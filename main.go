package main

import (
	"context"
	"flag"
	"fmt"
	"just-notify/commands"
	"just-notify/config"
	"just-notify/database"
	"just-notify/notification"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type ArgsCli struct {
	time        string
	notif       string
	category    string
	description string
	useDatabase bool
	connString  string
}

var wg sync.WaitGroup

func main() {
	args := parseArgs()

	if args.time == "" {
		fmt.Fprintln(os.Stderr, "\nERROR: Time argument is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if args.useDatabase && args.connString == "" {
		if config.LoadConfig()["CONN"] == "" {
			fmt.Fprintf(os.Stderr, "The database is enabled, but there is no connection string provided.")
		}
	}

	millis, err := commands.GetTime(args.time)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scheduling task: %s\n", err.Error())
		os.Exit(1)
	}

	wg.Add(1)

	closeSignal := make(chan bool, 1)

	go notification.Schedule(closeSignal, millis, func(now, epochMillis int64) {
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
		wg.Done()
	})

	fmt.Printf("Alert scheduled for %s\n", args.time)

	go func() {
		cx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
		defer stop()

		<-cx.Done()
		closeSignal <- true

		wg.Wait()
		os.Exit(0)
	}()

	wg.Wait()

}

func parseArgs() ArgsCli {
	var rawTime string
	var notif string
	var category string
	var description string
	var useDatabase bool
	var connString string

	flag.StringVar(&rawTime, "t", "", "Time scheduled for the notification (e.g. <mm>m = Time and suffix \"m\" for minutes, or <hh:mm>Hour:minute")
	flag.StringVar(&category, "c", "", "Category: The category of the task to be executed during focus time. e.g. work.")
	flag.StringVar(&notif, "n", "", "Notification title: The title for the notification to be shown")
	flag.BoolVar(&useDatabase, "d", false, "Indicate whether a SQL database will be used")
	flag.StringVar(&connString, "s", "", "Connection string used to connect to the database; it only works if the database flag is enabled.")
	flag.StringVar(&description, "l", "", "Optional: Details of the task")
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
	}
}
