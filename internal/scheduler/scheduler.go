package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	"gommutetime/internal/config"
	"gommutetime/internal/fetcher"
)

// Scheduler manages scheduled commute time fetches
type Scheduler struct {
	scheduler gocron.Scheduler
	fetcher   *fetcher.Fetcher
	config    *config.Config
}

// New creates a new scheduler instance
func New(cfg *config.Config, fetch *fetcher.Fetcher) (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	return &Scheduler{
		scheduler: s,
		fetcher:   fetch,
		config:    cfg,
	}, nil
}

// Start initializes all jobs from config and starts the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	// Create jobs for each itinerary/schedule combination
	jobCount := 0
	for _, itinerary := range s.config.Itineraries {
		for _, schedule := range itinerary.Schedules {
			count, err := s.addSchedule(ctx, itinerary, schedule)
			if err != nil {
				return fmt.Errorf("failed to add schedule %s for %s: %w",
					schedule.Name, itinerary.ID, err)
			}
			jobCount += count
		}
	}

	// Start the scheduler
	s.scheduler.Start()
	log.Printf("Scheduler started with %d jobs", jobCount)

	return nil
}

// addSchedule creates jobs for a single schedule configuration
func (s *Scheduler) addSchedule(ctx context.Context, itin config.Itinerary, sched config.Schedule) (int, error) {
	// Parse start and end times
	startHour, startMin, err := config.ParseTime(sched.StartTime)
	if err != nil {
		return 0, fmt.Errorf("invalid start time: %w", err)
	}

	endHour, endMin, err := config.ParseTime(sched.EndTime)
	if err != nil {
		return 0, fmt.Errorf("invalid end time: %w", err)
	}

	// Convert day names to weekdays
	weekdays := []time.Weekday{}
	for _, dayName := range sched.Days {
		day, err := config.DayNameToWeekday(dayName)
		if err != nil {
			return 0, err
		}
		weekdays = append(weekdays, day)
	}

	// Create the job task with panic recovery
	task := s.createTask(itin)

	// Generate time slots within the window
	slots := generateTimeSlots(startHour, startMin, endHour, endMin, sched.IntervalMinutes)

	// Create a job for each time slot
	jobCount := 0
	for _, slot := range slots {
		// Build cron expression for this specific time on specified days
		cronExpr := buildCronExpression(slot.hour, slot.minute, weekdays)

		_, err := s.scheduler.NewJob(
			gocron.CronJob(cronExpr, false),
			gocron.NewTask(task),
			gocron.WithName(fmt.Sprintf("%s-%s-%02d:%02d", itin.ID, sched.Name, slot.hour, slot.minute)),
		)

		if err != nil {
			return 0, fmt.Errorf("failed to create job for %02d:%02d: %w", slot.hour, slot.minute, err)
		}
		jobCount++
	}

	log.Printf("Created %d jobs for %s (%s)", jobCount, itin.ID, sched.Name)
	return jobCount, nil
}

// createTask creates a task function with panic recovery
func (s *Scheduler) createTask(itin config.Itinerary) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC in job %s: %v", itin.ID, r)
			}
		}()

		jobCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		log.Printf("Fetching: %s -> %s (%s)", itin.From, itin.To, itin.Name)

		if err := s.fetcher.FetchAndSave(jobCtx, itin.From, itin.To, itin.OutputFile); err != nil {
			log.Printf("ERROR fetching %s: %v", itin.ID, err)
		} else {
			log.Printf("Successfully saved to %s", itin.OutputFile)
		}
	}
}

// timeSlot represents a specific hour:minute
type timeSlot struct {
	hour   int
	minute int
}

// generateTimeSlots creates all time slots within a window at the specified interval
func generateTimeSlots(startHour, startMin, endHour, endMin, intervalMinutes int) []timeSlot {
	var slots []timeSlot

	startTotalMin := startHour*60 + startMin
	endTotalMin := endHour*60 + endMin

	for currentMin := startTotalMin; currentMin <= endTotalMin; currentMin += intervalMinutes {
		hour := currentMin / 60
		minute := currentMin % 60

		// Ensure we don't go past 23:59
		if hour > 23 {
			break
		}

		slots = append(slots, timeSlot{hour: hour, minute: minute})
	}

	return slots
}

// buildCronExpression creates a cron expression for specific time and days
func buildCronExpression(hour, minute int, weekdays []time.Weekday) string {
	// Cron format: minute hour day-of-month month day-of-week
	// Example: "15 6 * * 1-5" = 6:15 AM Monday-Friday

	// Convert weekdays to cron day numbers (0=Sunday, 1=Monday, etc.)
	dayNums := make([]string, len(weekdays))
	for i, day := range weekdays {
		dayNums[i] = fmt.Sprintf("%d", int(day))
	}

	// Join days with commas
	daysStr := ""
	for i, dayNum := range dayNums {
		if i > 0 {
			daysStr += ","
		}
		daysStr += dayNum
	}

	return fmt.Sprintf("%d %d * * %s", minute, hour, daysStr)
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() error {
	return s.scheduler.Shutdown()
}

// Reload reloads configuration and restarts scheduler
func (s *Scheduler) Reload(ctx context.Context, newConfig *config.Config) error {
	log.Println("Reloading scheduler configuration...")

	// Shutdown old scheduler
	if err := s.scheduler.Shutdown(); err != nil {
		log.Printf("Warning: error shutting down old scheduler: %v", err)
	}

	// Create new scheduler
	newScheduler, err := gocron.NewScheduler()
	if err != nil {
		return fmt.Errorf("failed to create new scheduler: %w", err)
	}

	s.scheduler = newScheduler
	s.config = newConfig

	return s.Start(ctx)
}
