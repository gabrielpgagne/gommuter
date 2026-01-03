package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the entire application configuration
type Config struct {
	API         APIConfig   `yaml:"api"`
	DataDir     string      `yaml:"data_dir"`
	Itineraries []Itinerary `yaml:"itineraries"`
}

// APIConfig holds Google Maps API settings
type APIConfig struct {
	Key string `yaml:"key"`
}

// Itinerary represents a single route to monitor
type Itinerary struct {
	ID         string     `yaml:"id"`
	Name       string     `yaml:"name"`
	From       string     `yaml:"from"`
	To         string     `yaml:"to"`
	OutputFile string     `yaml:"output_file"`
	Schedules  []Schedule `yaml:"schedules"`
}

// Schedule defines when to fetch commute times
type Schedule struct {
	Name            string   `yaml:"name"`
	Days            []string `yaml:"days"`
	StartTime       string   `yaml:"start_time"`
	EndTime         string   `yaml:"end_time"`
	IntervalMinutes int      `yaml:"interval_minutes"`
}

// LoadConfig reads and parses the config file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Override API key with environment variable if present
	if envKey := os.Getenv("GOOGLE_MAPS_API_KEY"); envKey != "" {
		cfg.API.Key = envKey
	}

	return &cfg, nil
}

// Validate checks config for errors
func (c *Config) Validate() error {
	// Check API key
	if c.API.Key == "" {
		return fmt.Errorf("API key is required (set in config or GOOGLE_MAPS_API_KEY env var)")
	}

	// Check data directory
	if c.DataDir == "" {
		return fmt.Errorf("data_dir is required")
	}

	// Check itineraries
	if len(c.Itineraries) == 0 {
		return fmt.Errorf("at least one itinerary is required")
	}

	// Track unique IDs and output files
	seenIDs := make(map[string]bool)
	seenFiles := make(map[string]bool)

	for i, itin := range c.Itineraries {
		// Check required fields
		if itin.ID == "" {
			return fmt.Errorf("itinerary %d: id is required", i)
		}
		if itin.Name == "" {
			return fmt.Errorf("itinerary %s: name is required", itin.ID)
		}
		if itin.From == "" {
			return fmt.Errorf("itinerary %s: from address is required", itin.ID)
		}
		if itin.To == "" {
			return fmt.Errorf("itinerary %s: to address is required", itin.ID)
		}
		if itin.OutputFile == "" {
			return fmt.Errorf("itinerary %s: output_file is required", itin.ID)
		}

		// Check for duplicate IDs
		if seenIDs[itin.ID] {
			return fmt.Errorf("duplicate itinerary ID: %s", itin.ID)
		}
		seenIDs[itin.ID] = true

		// Check for duplicate output files
		if seenFiles[itin.OutputFile] {
			return fmt.Errorf("duplicate output_file: %s (used by multiple itineraries)", itin.OutputFile)
		}
		seenFiles[itin.OutputFile] = true

		// Validate schedules
		if len(itin.Schedules) == 0 {
			return fmt.Errorf("itinerary %s: at least one schedule is required", itin.ID)
		}

		for j, sched := range itin.Schedules {
			if err := validateSchedule(sched, itin.ID, j); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateSchedule checks a single schedule for errors
func validateSchedule(sched Schedule, itinID string, schedIndex int) error {
	if sched.Name == "" {
		return fmt.Errorf("itinerary %s, schedule %d: name is required", itinID, schedIndex)
	}

	// Validate days
	if len(sched.Days) == 0 {
		return fmt.Errorf("itinerary %s, schedule %s: at least one day is required", itinID, sched.Name)
	}
	for _, day := range sched.Days {
		if _, err := DayNameToWeekday(day); err != nil {
			return fmt.Errorf("itinerary %s, schedule %s: %w", itinID, sched.Name, err)
		}
	}

	// Validate start time
	startHour, startMin, err := ParseTime(sched.StartTime)
	if err != nil {
		return fmt.Errorf("itinerary %s, schedule %s: invalid start_time: %w", itinID, sched.Name, err)
	}

	// Validate end time
	endHour, endMin, err := ParseTime(sched.EndTime)
	if err != nil {
		return fmt.Errorf("itinerary %s, schedule %s: invalid end_time: %w", itinID, sched.Name, err)
	}

	// Check start < end
	startMinutes := startHour*60 + startMin
	endMinutes := endHour*60 + endMin
	if startMinutes >= endMinutes {
		return fmt.Errorf("itinerary %s, schedule %s: start_time must be before end_time", itinID, sched.Name)
	}

	// Validate interval
	if sched.IntervalMinutes <= 0 {
		return fmt.Errorf("itinerary %s, schedule %s: interval_minutes must be positive", itinID, sched.Name)
	}
	if sched.IntervalMinutes > 1440 {
		return fmt.Errorf("itinerary %s, schedule %s: interval_minutes cannot exceed 1440 (1 day)", itinID, sched.Name)
	}

	return nil
}

// ParseTime converts HH:MM string to hour and minute components
func ParseTime(timeStr string) (hour, minute int, err error) {
	var h, m int
	_, err = fmt.Sscanf(timeStr, "%d:%d", &h, &m)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid time format '%s' (expected HH:MM)", timeStr)
	}

	if h < 0 || h > 23 {
		return 0, 0, fmt.Errorf("hour must be 0-23, got %d", h)
	}
	if m < 0 || m > 59 {
		return 0, 0, fmt.Errorf("minute must be 0-59, got %d", m)
	}

	return h, m, nil
}

// DayNameToWeekday converts day names to time.Weekday
func DayNameToWeekday(day string) (time.Weekday, error) {
	dayLower := strings.ToLower(day)
	switch dayLower {
	case "sunday", "sun":
		return time.Sunday, nil
	case "monday", "mon":
		return time.Monday, nil
	case "tuesday", "tue", "tues":
		return time.Tuesday, nil
	case "wednesday", "wed":
		return time.Wednesday, nil
	case "thursday", "thu", "thurs":
		return time.Thursday, nil
	case "friday", "fri":
		return time.Friday, nil
	case "saturday", "sat":
		return time.Saturday, nil
	default:
		return time.Sunday, fmt.Errorf("invalid day name: %s", day)
	}
}
