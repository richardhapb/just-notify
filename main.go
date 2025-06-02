package main

import (
	"flag"
	"fmt"
	"just-notify/commands"
	"just-notify/config"
	"just-notify/database"
	"just-notify/notification"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type ArgsCli struct {
	time        string
	notif       string
	category    string
	description string
	useDatabase bool
	connString  string
	unlimited   bool
}

type app struct {
	wg          sync.WaitGroup
	closeSignal chan bool
	cfg         map[string]string
}

const (
	defaultCategory = "Unknown"
	defaultNotif    = "Time has been finalized"
)

func main() {
	app := &app{
		closeSignal: make(chan bool, 1),
		cfg:         config.LoadConfig(),
	}

	args := parseArgs(app.cfg)
	if err := validateArgs(&args, app.cfg); err != nil {
		flag.PrintDefaults()
		log.Fatalln(err)
	}

	var millis int64
	if !args.unlimited {
		var err error
		millis, err = commands.GetTime(args.time)
		if err != nil {
			log.Fatalf("Error scheduling task: %v", err)
		}
		fmt.Printf("Alert scheduled for %s\n", args.time)
	}

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Create error channel for goroutine errors
	errChan := make(chan error, 1)

	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		notification.Schedule(!args.unlimited, app.closeSignal, millis, func(now, epochMillis int64) {
			notification.Notify(args.notif, fmt.Sprintf("Time completed: %s", args.category))

			var logger database.Logger
			var err error
			if args.useDatabase {
				logger, err = database.NewLogger(args.connString, true)
			} else {
				logger, err = database.NewLogger(app.cfg["CSV_PATH"], false)
			}

			if err != nil {
				errChan <- fmt.Errorf("failed to create logger: %w", err)
				return
			}
			defer logger.Close()

			if err := logger.Log(&database.LogEntry{
				InitTime:    now,
				EndTime:     epochMillis,
				Category:    args.category,
				Description: args.description,
			}); err != nil {
				errChan <- fmt.Errorf("failed to log entry: %w", err)
				return
			}

			log.Println("Entry logged successfully.")
		})
	}()

	// Wait with timeout for goroutines to finish
	done := make(chan struct{})
	go func() {
		app.wg.Wait()
		close(done)
	}()

	go func() {
		// Wait for either signal or error
		startTime := time.Now()
		select {
		case sig := <-sigChan:
			log.Printf("\nReceived signal: %v", sig)

			// Send close signal with timeout
			select {
			case app.closeSignal <- true:
				elapsed := time.Since(startTime)
				log.Printf("Time elapsed: %.2f minutes", elapsed.Minutes())
			case <-time.After(3 * time.Second):
				log.Printf("Warning: Failed to send close signal (timeout)")
			}
		case err := <-errChan:
			log.Printf("Error during execution: %v", err)
		}
	}()

	select {
	case <-done:
		log.Println("Shutdown successfully")
	}
}

func parseArgs(cfg map[string]string) ArgsCli {
	var rawTime string
	var notif string
	var category string
	var description string
	var useDatabase bool
	var connString string
	var unlimited bool

	flag.StringVar(&rawTime, "t", "", "Time scheduled for the notification (e.g. <mm>m = Time and suffix \"m\" for minutes, or <hh:mm>Hour:minute")
	flag.StringVar(&category, "c", "", "Category: The category of the task to be executed during focus time. e.g. work.")
	flag.StringVar(&notif, "n", "", "Notification title: The title for the notification to be shown")
	flag.BoolVar(&useDatabase, "d", false, "Indicate whether a SQL database will be used")
	flag.StringVar(&connString, "s", "", "Connection string used to connect to the database; it only works if the database flag is enabled.")
	flag.StringVar(&description, "l", "", "Optional: Details of the task")
	flag.BoolVar(&unlimited, "u", false, "Unlimited time")
	flag.Parse()

	if category == "" {
		category = cfg["DEFAULT_CATEGORY"]
	}

	if notif == "" {
		notif = cfg["DEFAULT_NOTIFICATION"]
	}

	if category == "" {
		category = defaultCategory
	}

	if notif == "" {
		notif = defaultNotif
	}

	return ArgsCli{
		time:        rawTime,
		notif:       notif,
		category:    category,
		description: description,
		useDatabase: useDatabase,
		connString:  connString,
		unlimited:   unlimited,
	}
}

func validateArgs(args *ArgsCli, cfg map[string]string) error {
	if args.time == "" && !args.unlimited {
		return fmt.Errorf("\nERROR: Time argument is required")
	}

	if args.category == "" {
		return fmt.Errorf("\nERROR: Category is required")
	}

	if args.useDatabase {
		if args.connString == "" && cfg["CONN"] == "" {
			return fmt.Errorf("Database enabled but no connection string provided")
		}
		if args.connString == "" {
			args.connString = cfg["CONN"]
		}
	}

	return nil
}
