package fetcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"googlemaps.github.io/maps"
)

// Fetcher handles commute time fetching
type Fetcher struct {
	client  *maps.Client
	dataDir string
}

// New creates a new Fetcher instance
func New(apiKey, dataDir string) (*Fetcher, error) {
	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create maps client: %w", err)
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}

	return &Fetcher{
		client:  client,
		dataDir: dataDir,
	}, nil
}

// FetchAndSave gets commute time and appends to CSV file
func (f *Fetcher) FetchAndSave(ctx context.Context, from, to, outputFile string) error {
	// Create distance matrix request
	req := &maps.DistanceMatrixRequest{
		Origins:       []string{from},
		Destinations:  []string{to},
		DepartureTime: "now",
	}

	// Call API
	routes, err := f.client.DistanceMatrix(ctx, req)
	if err != nil {
		return fmt.Errorf("distance matrix API error: %w", err)
	}

	// Extract duration
	if len(routes.Rows) == 0 || len(routes.Rows[0].Elements) == 0 {
		return fmt.Errorf("no route found from %s to %s", from, to)
	}

	element := routes.Rows[0].Elements[0]
	if element.Status != "OK" {
		return fmt.Errorf("route status: %s", element.Status)
	}

	// Format CSV line
	timestamp := time.Now().Format(time.RFC3339)
	duration := element.DurationInTraffic.Minutes()
	line := fmt.Sprintf("%s,%f\n", timestamp, duration)

	// Append to file
	filePath := filepath.Join(f.dataDir, outputFile)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(line); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// Fetch gets commute time without saving (for fetch subcommand)
func (f *Fetcher) Fetch(ctx context.Context, from, to string) (float64, error) {
	// Create distance matrix request
	req := &maps.DistanceMatrixRequest{
		Origins:       []string{from},
		Destinations:  []string{to},
		DepartureTime: "now",
	}

	// Call API
	routes, err := f.client.DistanceMatrix(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("distance matrix API error: %w", err)
	}

	// Extract duration
	if len(routes.Rows) == 0 || len(routes.Rows[0].Elements) == 0 {
		return 0, fmt.Errorf("no route found from %s to %s", from, to)
	}

	element := routes.Rows[0].Elements[0]
	if element.Status != "OK" {
		return 0, fmt.Errorf("route status: %s", element.Status)
	}

	return element.DurationInTraffic.Minutes(), nil
}
