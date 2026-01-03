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

	// Watch the directory (not the file directly, handles editor rewrites)
	dir := filepath.Dir(configPath)
	if err := watcher.Add(dir); err != nil {
		watcher.Close()
		return nil, err
	}

	return &Watcher{
		configPath: configPath,
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

			// Only reload on Write or Create events for our config file
			if event.Name == w.configPath && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
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
