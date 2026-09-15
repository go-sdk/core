package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-sdk/core/testx"
)

func TestLoadYAMLAndEnvironment(t *testing.T) {
	t.Setenv("APP__DATABASE__TYPE", "mysql")
	filename := writeConfigFile(t, "config.yaml", `
database:
  type: postgres
  port: 5432
message: "${database.type}:${database.port}"
`)
	c := New(WithFile(filename))
	testx.NoError(t, c.Load())

	databaseType, ok := c.Get[string]("database.type")
	testx.True(t, ok)
	testx.Equal(t, "mysql", databaseType)
	testx.Equal(t, "mysql:5432", c.MustGet[string]("message"))
	testx.True(t, c.Exists("database.port"))
	testx.False(t, c.Exists("database.host"))

	raw := c.Raw()
	database := raw["database"].(map[string]any)
	testx.Equal(t, "mysql", database["type"])
	delete(database, "type")
	testx.True(t, c.Exists("database.type"))
}

func TestLoadExplicitFileType(t *testing.T) {
	filename := writeConfigFile(t, "config.conf", `{"server":{"port":8080}}`)
	c := New(WithFile(filename, "json"))
	testx.NoError(t, c.Load())
	port, ok := c.Get[int]("server.port")
	testx.True(t, ok)
	testx.Equal(t, 8080, port)
}

func TestLoadInvalidOptionsAndTypes(t *testing.T) {
	filename := writeConfigFile(t, "config.data", `{}`)
	testx.Error(t, New(WithFile("")).Load())
	testx.ErrorContains(t, New(WithFile(filename)).Load(), "unsupported file extension")
	testx.ErrorContains(t, New(WithFile(filename, "toml")).Load(), "unsupported file type")
	testx.ErrorContains(t, New(WithFile(filename, "json", "yaml")).Load(), "at most one")
	testx.ErrorContains(t, New(WithFileWatch(true)).Load(), "requires WithFile")
}

func TestGetAndMustGet(t *testing.T) {
	c := New()
	testx.NoError(t, c.Load())
	value, ok := c.Get[string]("missing")
	testx.False(t, ok)
	testx.Empty(t, value)
	testx.Equal(t, "default", c.MustGet("missing", "default"))
	testx.Equal(t, "first", c.MustGet("missing", "first", "second"))
	testx.Panics(t, func() { c.MustGet[string]("missing") })
}

func TestDecodeToUsesJSONTags(t *testing.T) {
	filename := writeConfigFile(t, "config.json", `{"database":{"database_type":"mysql","port":"3306"}}`)
	c := New(WithFile(filename))
	testx.NoError(t, c.Load())

	var target struct {
		Database struct {
			Type string `json:"database_type"`
			Port int    `json:"port"`
		} `json:"database"`
	}
	testx.NoError(t, c.DecodeTo(&target))
	testx.Equal(t, "mysql", target.Database.Type)
	testx.Equal(t, 3306, target.Database.Port)
	testx.Error(t, c.DecodeTo(target))
}

func TestDecodeToParsesDurationAndTime(t *testing.T) {
	filename := writeConfigFile(t, "config.json", `{"timeout":"24h","started_at":"2026-09-13T08:30:00Z"}`)
	c := New(WithFile(filename))
	testx.NoError(t, c.Load())

	var target struct {
		Timeout   time.Duration `json:"timeout"`
		StartedAt time.Time     `json:"started_at"`
	}

	testx.NoError(t, c.DecodeTo(&target))
	testx.Equal(t, 24*time.Hour, target.Timeout)
	testx.Equal(
		t,
		time.Date(2026, time.September, 13, 8, 30, 0, 0, time.UTC),
		target.StartedAt,
	)
}

func TestDecodeToRejectsInvalidDurationAndTime(t *testing.T) {
	filename := writeConfigFile(t, "config.json", `{"timeout":"tomorrow","started_at":"not-a-time"}`)
	c := New(WithFile(filename))
	testx.NoError(t, c.Load())

	var target struct {
		Timeout   time.Duration `json:"timeout"`
		StartedAt time.Time     `json:"started_at"`
	}

	testx.Error(t, c.DecodeTo(&target))
}

func TestVariableErrors(t *testing.T) {
	missing := writeConfigFile(t, "missing.yaml", `value: "${missing}"`)
	testx.ErrorContains(t, New(WithFile(missing)).Load(), "does not exist")

	cycle := writeConfigFile(t, "cycle.yaml", `
first: "${second}"
second: "${third}"
third: "${first}"
`)
	testx.ErrorContains(t, New(WithFile(cycle)).Load(), "cyclic reference")

	composite := writeConfigFile(t, "composite.yaml", `
items:
  - one
value: "${items}"
`)
	testx.ErrorContains(t, New(WithFile(composite)).Load(), "not a scalar")
}

func TestSetDefault(t *testing.T) {
	previous := defaultConfig.Load()
	defer defaultConfig.Store(previous)

	filename := writeConfigFile(t, "default.yaml", `name: example`)
	c := New(WithFile(filename))
	testx.NoError(t, c.Load())
	SetDefault(c)

	name, ok := Get[string]("name")
	testx.True(t, ok)
	testx.Equal(t, "example", name)
	testx.Equal(t, "example", MustGet[string]("name"))
	testx.Equal(t, 42, MustGet("missing", 42))
	testx.True(t, Exists("name"))
	testx.False(t, Exists("missing"))
	raw := Raw()
	testx.Equal(t, "example", raw["name"])
	raw["name"] = "changed"
	testx.Equal(t, "example", Raw()["name"])
	var target struct {
		Name string `json:"name"`
	}
	testx.NoError(t, DecodeTo(&target))
	testx.Equal(t, "example", target.Name)
	testx.Panics(t, func() { SetDefault(nil) })
}

func TestFileWatchReloadsValidSnapshot(t *testing.T) {
	filename := writeConfigFile(t, "watch.yaml", `value: one`)
	c := New(WithFile(filename), WithFileWatch(true))
	testx.NoError(t, c.Load())
	t.Cleanup(func() { testx.NoError(t, c.closeWatcher()) })

	testx.NoError(t, os.WriteFile(filename, []byte("value: two\n"), 0o600))
	waitForValue(t, c, "two")

	testx.NoError(t, os.WriteFile(filename, []byte("value: [\n"), 0o600))
	time.Sleep(2 * watchDebounce)
	testx.Equal(t, "two", c.MustGet[string]("value"))
}

func writeConfigFile(t *testing.T, name, content string) string {
	t.Helper()
	filename := filepath.Join(t.TempDir(), name)
	testx.NoError(t, os.WriteFile(filename, []byte(content), 0o600))
	return filename
}

func waitForValue(t *testing.T, c *Config, expected string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if value, ok := c.Get[string]("value"); ok && value == expected {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	testx.Fail(t, "configuration was not reloaded")
}
