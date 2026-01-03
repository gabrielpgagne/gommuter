package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gommutetime/internal/config"
	"gommutetime/internal/fetcher"
	"gommutetime/internal/scheduler"
	"gommutetime/internal/watcher"
)

const (
	defaultConfigPath = "/app/config.yaml"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "schedule":
		runScheduler(os.Args[2:])
	case "fetch":
		runFetch(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("gommutetime - Google Maps commute time tracker")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gommutetime schedule [options]  Run scheduler with config file")
	fmt.Println("  gommutetime fetch [options]     Fetch commute time once")
	fmt.Println("  gommutetime help                Show this help")
	fmt.Println()
	fmt.Println("Schedule options:")
	fmt.Println("  -config string    Path to config file (default: /app/config.yaml)")
	fmt.Println()
	fmt.Println("Fetch options:")
	fmt.Println("  -from string      Starting point (required)")
	fmt.Println("  -to string        Destination (required)")
	fmt.Println("  -key string       Google Maps API key (optional, uses GOOGLE_MAPS_API_KEY env var)")
	fmt.Println()
}

func runScheduler(args []string) {
	fs := flag.NewFlagSet("schedule", flag.ExitOnError)
	configPath := fs.String("config", defaultConfigPath, "Path to config file")
	fs.Parse(args)

	// Load config
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	// Create fetcher
	apiKey := cfg.API.Key
	if envKey := os.Getenv("GOOGLE_MAPS_API_KEY"); envKey != "" {
		apiKey = envKey
	}

	fetch, err := fetcher.New(apiKey, cfg.DataDir)
	if err != nil {
		log.Fatalf("Failed to create fetcher: %v", err)
	}

	// Create scheduler
	sched, err := scheduler.New(cfg, fetch)
	if err != nil {
		log.Fatalf("Failed to create scheduler: %v", err)
	}

	// Start scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := sched.Start(ctx); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}

	// Setup config file watcher
	watch, err := watcher.New(*configPath, func(newCfg *config.Config) error {
		if err := newCfg.Validate(); err != nil {
			return err
		}
		return sched.Reload(ctx, newCfg)
	})
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}

	// Start watcher in goroutine
	go func() {
		if err := watch.Start(ctx); err != nil {
			log.Printf("Watcher stopped: %v", err)
		}
	}()

	// Wait for shutdown signal
	log.Println("Scheduler running. Press Ctrl+C to stop.")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	cancel()

	if err := sched.Stop(); err != nil {
		log.Printf("Error stopping scheduler: %v", err)
	}

	log.Println("Goodbye!")
}

func runFetch(args []string) {
	fs := flag.NewFlagSet("fetch", flag.ExitOnError)
	from := fs.String("from", "", "Starting point")
	to := fs.String("to", "", "Destination")
	key := fs.String("key", "", "Google Maps API Key (optional)")
	fs.Parse(args)

	if *from == "" || *to == "" {
		fmt.Println("Error: -from and -to are required")
		fmt.Println()
		fs.PrintDefaults()
		os.Exit(1)
	}

	// Get API key
	apiKey := *key
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_MAPS_API_KEY")
		if apiKey == "" {
			fmt.Println("Error: API key required (use -key or GOOGLE_MAPS_API_KEY env var)")
			os.Exit(1)
		}
	}

	// Create fetcher (with temp data dir, not used for fetch command)
	fetch, err := fetcher.New(apiKey, "/tmp")
	if err != nil {
		log.Fatalf("Failed to create fetcher: %v", err)
	}

	// Fetch commute time
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	duration, err := fetch.Fetch(ctx, *from, *to)
	if err != nil {
		log.Fatalf("Failed to fetch commute time: %v", err)
	}

	// Output in same CSV format as before
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Printf("%s,%f\n", timestamp, duration)
}
