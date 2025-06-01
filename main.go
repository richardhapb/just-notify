package main

import (
	"context"
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

	args := parseArgs()

	if args.time == "" && !args.unlimited {
		fmt.Fprintln(os.Stderr, "\nERROR: Time argument is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if args.useDatabase {
		if args.connString == "" && app.cfg["CONN"] == "" {
			log.Fatal("Database enabled but no connection string provided")
		}
		if args.connString == "" {
			args.connString = app.cfg["CONN"]
		}
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

	app.wg.Add(1)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go notification.Schedule(!args.unlimited, app.closeSignal, millis, func(now, epochMillis int64) {
		notification.Notify(args.notif, fmt.Sprintf("Time completed: %s", args.category))
		if err := database.LogData(database.LogEntry{
			InitTime:    now,
			EndTime:     epochMillis,
			Category:    args.category,
			Description: args.description,
		}, args.useDatabase, args.connString); err != nil {
			fmt.Fprintf(os.Stderr, "error inserting data in database: %s", err)
			os.Exit(1)
		}
		app.wg.Done()
	})

	go func() {
		startTime := time.Now()
		select {
		case <-ctx.Done():
			app.closeSignal <- true
			elapsed := time.Since(startTime)
			time.Sleep(time.Duration(100) * time.Millisecond)
			fmt.Printf("\n\nTime elapsed: %.2f minutes\n", elapsed.Minutes())
		}
	}()

	app.wg.Wait()

}

func parseArgs() ArgsCli {
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

	config := config.LoadConfig()

	if category == "" {
		category = config["DEFAULT_TITLE"]
	}

	if notif == "" {
		notif = config["DEFAULT_NOTIFICATION_NAME"]
	}

	if category == "" {
		category = "Unknown"
	}

	if notif == "" {
		notif = "Time has been finalized"
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
