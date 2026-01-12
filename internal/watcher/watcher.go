package watcher

import (
	"context"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"gommutetime/internal/config"
)

// Watcher monitors config file for changes
type Watcher struct {
	configPath string
	watcher    *fsnotify.Watcher
	onReload   func(*config.Config) error
}

// New creates a new config file watcher
func New(configPath string, onReload func(*config.Config) error) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Get absolute path for more reliable watching
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		watcher.Close()
		return nil, err
	}

	// Watch both the file and its directory for maximum compatibility
	// This handles both direct edits and atomic editor rewrites
	dir := filepath.Dir(absPath)
	if err := watcher.Add(dir); err != nil {
		watcher.Close()
		return nil, err
	}

	// Also watch the file directly (helps with Docker bind mounts)
	if err := watcher.Add(absPath); err != nil {
		log.Printf("Warning: Could not watch file directly: %v", err)
		// Continue anyway, directory watch might be sufficient
	}

	return &Watcher{
		configPath: absPath,
		watcher:    watcher,
		onReload:   onReload,
	}, nil
}

// Start begins watching for config changes
func (w *Watcher) Start(ctx context.Context) error {
	log.Printf("Watching for config changes: %s", w.configPath)

	for {
		select {
		case <-ctx.Done():
			return w.watcher.Close()

		case event, ok := <-w.watcher.Events:
			if !ok {
				return nil
			}

			// Get absolute path of the event for comparison
			eventPath, _ := filepath.Abs(event.Name)

			// Log all events for debugging (can be removed later)
			log.Printf("File event: %s %s", event.Op, event.Name)

			// Reload on Write, Create, or Chmod events for our config file
			// Chmod is included because some editors change permissions during save
			if eventPath == w.configPath &&
				(event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create ||
					event.Op&fsnotify.Chmod == fsnotify.Chmod) {

				log.Println("Config file changed, reloading...")

				cfg, err := config.LoadConfig(w.configPath)
				if err != nil {
					log.Printf("ERROR: Failed to reload config: %v", err)
					log.Println("Keeping previous configuration")
					continue
				}

				if err := cfg.Validate(); err != nil {
					log.Printf("ERROR: Invalid new config: %v", err)
					log.Println("Keeping previous configuration")
					continue
				}

				if err := w.onReload(cfg); err != nil {
					log.Printf("ERROR: Failed to apply new config: %v", err)
					continue
				}

				log.Println("Config reloaded successfully")
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}
