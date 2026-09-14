package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	jsoncodec "github.com/go-sdk/core/codec/json"
	yamlcodec "github.com/go-sdk/core/codec/yaml"
	"github.com/go-sdk/core/errx"
)

// Load 重新读取全部配置来源，并在校验成功后一次性发布新快照。
func (c *Config) Load() error {
	c.loadMu.Lock()
	defer c.loadMu.Unlock()

	if c.fileWatch && !c.hasFile {
		return errx.New("config: file watch requires WithFile")
	}
	if err := c.load(); err != nil {
		return err
	}
	if c.fileWatch {
		if err := c.ensureWatcher(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) load() (err error) {
	if c.optionErr != nil {
		return c.optionErr
	}
	if c.hasFile {
		slog.Info("loading config file", "file", c.filename)
		defer func() {
			if err != nil {
				slog.Error("config file load failed", "file", c.filename, "error", err)
				return
			}
			slog.Info("config file loaded", "file", c.filename)
		}()
	}

	flat := map[string]any{}
	if c.hasFile {
		var fileData map[string]any
		fileData, err = c.loadFile()
		if err != nil {
			return err
		}
		if err = flattenMap("", fileData, flat); err != nil {
			return errx.Wrapf(err, "config: flatten file %q", c.filename)
		}
	}

	mergeEnv(flat)
	resolved, err := resolveVariables(flat)
	if err != nil {
		return errx.Wrap(err, "config: resolve variables")
	}
	nested, err := unflattenMap(resolved)
	if err != nil {
		return errx.Wrap(err, "config: restore nested data")
	}

	c.dataMu.Lock()
	c.flatData = resolved
	c.nestedData = nested
	c.dataMu.Unlock()
	return nil
}

func (c *Config) loadFile() (map[string]any, error) {
	bs, err := os.ReadFile(c.filename)
	if err != nil {
		return nil, errx.Wrapf(err, "config: read file %q", c.filename)
	}
	fileType, err := c.resolveFileType()
	if err != nil {
		return nil, err
	}
	data := map[string]any{}
	switch fileType {
	case "json":
		err = jsoncodec.Unmarshal(bs, &data)
	case "yaml":
		err = yamlcodec.Unmarshal(bs, &data)
	}
	if err != nil {
		return nil, errx.Wrapf(err, "config: parse %s file %q", fileType, c.filename)
	}
	return data, nil
}

func (c *Config) resolveFileType() (string, error) {
	fileType := c.fileType
	if fileType == "" {
		switch strings.ToLower(filepath.Ext(c.filename)) {
		case ".json":
			fileType = "json"
		case ".yaml", ".yml":
			fileType = "yaml"
		default:
			return "", errx.Newf("config: unsupported file extension for %q", c.filename)
		}
	}
	if fileType != "json" && fileType != "yaml" {
		return "", errx.Newf("config: unsupported file type %q", c.fileType)
	}
	return fileType, nil
}

func mergeEnv(flat map[string]any) {
	for _, item := range os.Environ() {
		key, value, ok := strings.Cut(item, "=")
		if !ok || !strings.HasPrefix(key, "APP__") {
			continue
		}
		key = strings.ToLower(strings.TrimPrefix(key, "APP__"))
		key = strings.ReplaceAll(key, "__", delimiter)
		if key == "" {
			continue
		}
		setFlat(flat, key, value)
	}
}

func flattenMap(prefix string, input map[string]any, output map[string]any) error {
	for key, value := range input {
		if key == "" {
			return errx.New("empty key is not supported")
		}
		path := key
		if prefix != "" {
			path = prefix + delimiter + key
		}
		switch child := value.(type) {
		case map[string]any:
			if len(child) == 0 {
				setFlat(output, path, map[string]any{})
				continue
			}
			if err := flattenMap(path, child, output); err != nil {
				return err
			}
		default:
			setFlat(output, path, cloneValue(value))
		}
	}
	return nil
}

func setFlat(data map[string]any, key string, value any) {
	for existing := range data {
		if strings.HasPrefix(key, existing+delimiter) || strings.HasPrefix(existing, key+delimiter) {
			delete(data, existing)
		}
	}
	data[key] = value
}

func unflattenMap(flat map[string]any) (map[string]any, error) {
	root := map[string]any{}
	for path, value := range flat {
		parts := strings.Split(path, delimiter)
		current := root
		for index, part := range parts {
			if part == "" {
				return nil, errx.Newf("path %q contains an empty segment", path)
			}
			if index == len(parts)-1 {
				current[part] = cloneValue(value)
				continue
			}
			next, ok := current[part]
			if !ok {
				child := map[string]any{}
				current[part] = child
				current = child
				continue
			}
			child, ok := next.(map[string]any)
			if !ok {
				return nil, errx.Newf("path %q conflicts with %q", path, strings.Join(parts[:index+1], delimiter))
			}
			current = child
		}
	}
	return root, nil
}

func cloneMap(input map[string]any) map[string]any {
	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = cloneValue(value)
	}
	return output
}

func cloneValue(value any) any {
	switch value := value.(type) {
	case map[string]any:
		return cloneMap(value)
	case []any:
		output := make([]any, len(value))
		for index := range value {
			output[index] = cloneValue(value[index])
		}
		return output
	default:
		return value
	}
}
