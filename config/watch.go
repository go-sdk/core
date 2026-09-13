package config

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/go-sdk/core/lifex"
	"github.com/go-sdk/core/logx"
)

const watchDebounce = 100 * time.Millisecond

func (c *Config) ensureWatcher() error {
	c.watchMu.Lock()
	defer c.watchMu.Unlock()
	if c.watcher != nil {
		return nil
	}

	filename, err := filepath.Abs(c.filename)
	if err != nil {
		return fmt.Errorf("config: resolve watched file %q: %w", c.filename, err)
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("config: create file watcher: %w", err)
	}
	if err = w.Add(filepath.Dir(filename)); err != nil {
		_ = w.Close()
		return fmt.Errorf("config: watch directory for %q: %w", filename, err)
	}

	done := make(chan struct{})
	c.watcher = w
	c.watchDone = done
	c.lifecycleOnce.Do(func() {
		lifex.OnDeinit(c.closeWatcher)
	})
	go c.watchFile(w, done, filepath.Clean(filename))
	return nil
}

func (c *Config) watchFile(w *fsnotify.Watcher, done chan struct{}, filename string) {
	defer close(done)
	var timer *time.Timer
	var timerC <-chan time.Time
	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			if filepath.Clean(event.Name) != filename || event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}
			if timer == nil {
				timer = time.NewTimer(watchDebounce)
			} else {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(watchDebounce)
			}
			timerC = timer.C
		case <-timerC:
			timerC = nil
			c.reloadWatchedFile()
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			logx.Error().Err(err).Str("file", c.filename).Msg("config file watch failed")
		}
	}
}

func (c *Config) reloadWatchedFile() {
	c.loadMu.Lock()
	defer c.loadMu.Unlock()
	if err := c.load(); err != nil {
		logx.Error().Err(err).Str("file", c.filename).Msg("config file reload failed")
	}
}

func (c *Config) closeWatcher() error {
	c.watchMu.Lock()
	w := c.watcher
	done := c.watchDone
	c.watcher = nil
	c.watchDone = nil
	c.watchMu.Unlock()
	if w == nil {
		return nil
	}
	err := w.Close()
	if done != nil {
		<-done
	}
	if err != nil {
		return fmt.Errorf("config: close file watcher: %w", err)
	}
	return nil
}
