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
	time  string
	notif string
	title string
}

var wg sync.WaitGroup

func main() {
	args := parseArgs()

	if args.time == "" {
		fmt.Fprintln(os.Stderr, "\nERROR: Time argument is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	millis, err := commands.GetTime(args.time)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scheduling task: %s\n", err.Error())
		os.Exit(1)
	}

	wg.Add(1)

	closeSignal := make(chan bool, 1)

	go notification.Schedule(closeSignal, millis, func(now, epochMillis int64) {
		notification.Notify(args.notif, fmt.Sprintf("Time completed: %s", args.title))
		if err := database.LogData(database.LogEntry{
			InitTime: now,
			EndTime:  epochMillis,
			TaskName: args.title,
		}); err != nil {
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
	var title string

	flag.StringVar(&rawTime, "time", "", "Time scheduled for the notification (e.g. <mm>m = Time and suffix \"m\" for minutes, or <hh:mm>Hour:minute")
	flag.StringVar(&title, "title", "", "Task title: The title of the task to be executed during focus time; this will be logged in the CSV file or database.")
	flag.StringVar(&notif, "notif", "", "Notification title: The title for the notification to be shown")
	flag.Parse()

	config := config.LoadConfig()

	if title == "" {
		title = config["DEFAULT_TITLE"]
	}

	if notif == "" {
		notif = config["DEFAULT_NOTIFICATION_NAME"]
	}

	if title == "" {
		title = "Unknown"
	}

	if notif == "" {
		notif = "Time has been finalized"
	}

	return ArgsCli{time: rawTime, notif: notif, title: title}
}
