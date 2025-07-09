package main

import (
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

type app struct {
	wg          sync.WaitGroup
	closeSignal chan bool
	cfg         map[string]string
}

func main() {

	app := &app{
		closeSignal: make(chan bool, 1),
		cfg:         config.LoadConfig(),
	}

	args, err := config.ParseArgs(app.cfg)

	if err != nil {
		log.Fatalf("error parsing the arguments: %s", err)
	}

	if err := config.ValidateArgs(args, app.cfg); err != nil {
		config.PrintUsage()
		log.Fatalln(err)
	}

	if args.Kill {
		pid, err := commands.KillProcess(args.Category)
		if err != nil {
			log.Fatalf("Error terminating the process: %s\n", err)
		}

		log.Printf("Process %d terminated sucessfully\n", pid)
		os.Exit(0)
	} else {
		// Create the pid file
		commands.StorePID(args.Category)
	}

	var millis int64
	if !args.Unlimited {
		var err error
		millis, err = commands.GetTime(args.Time)
		if err != nil {
			log.Fatalf("Error scheduling task: %v", err)
		}
		fmt.Printf("Alert scheduled for %s\n", args.Time)
	}

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Create error channel for goroutine errors
	errChan := make(chan error, 1)

	app.wg.Add(1)
	go func() {
		defer app.wg.Done()

		var logger database.Logger
		var err error
		if args.UseDatabase {
			logger, err = database.NewLogger(args.ConnString, true)
		} else {
			logger, err = database.NewLogger(args.CsvPath, false)
		}

		currentTime := time.Now().UnixMilli()

		exists, err := logger.Exists(&database.LogEntry{
			InitTime: currentTime,
			Category: args.Category,
		})

		if exists {
			log.Fatalf("The task with time %d and category %s already exists", currentTime, args.Category)
		}

		if err != nil {
			log.Fatalf("Error checking task: %s", err)
		}

		// Initialize the data before scheduling the task; this allows tracking if any
		// tasks exist and prevents data loss when the task is not finalized gracefully.
		if err := logger.Log(&database.LogEntry{
			InitTime:    currentTime,
			Category:    args.Category,
			Description: args.Description,
		}); err != nil {
			errChan <- fmt.Errorf("failed to log initial entry: %w", err)
			return
		}

		notification.Schedule(!args.Unlimited, app.closeSignal, currentTime, millis, func(now, epochMillis int64) {
			if !args.Headless {
				notification.Notify(args.Notif, fmt.Sprintf("Time completed: %s", args.Category))
			}

			if err != nil {
				errChan <- fmt.Errorf("failed to create logger: %w", err)
				return
			}
			defer logger.Close()

			if err := logger.Log(&database.LogEntry{
				InitTime:    now,
				EndTime:     epochMillis,
				Category:    args.Category,
				Description: args.Description,
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

	<-done
	log.Println("Shutdown successfully")
}
