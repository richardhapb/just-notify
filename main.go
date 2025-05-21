package main

import (
	"fmt"
	"just-notify/commands"
	"just-notify/config"
	"just-notify/database"
	"just-notify/notification"
	"os"
	"sync"
)

var wg sync.WaitGroup

func main() {
	args := os.Args

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "\nERROR: Time argument is required")
		commands.PrintUsage()
		os.Exit(1)
	}

	config := config.LoadConfig()

	timeArg := args[1]
	notificationName := config["DEFAULT_NOTIFICATION_NAME"]
	title := config["DEFAULT_TITLE"]

	if notificationName == "" {
		notificationName = "Time has been finalized"
	}

	if title == "" {
		title = "Unknown"
	}

	if len(args) > 2 {
		notificationName = args[2]
	}

	if len(args) > 3 {
		title = args[3]
	}

	millis, err := commands.GetTime(timeArg)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scheduling task: %s\n", err.Error())
		os.Exit(1)
	}

	wg.Add(1)

	notification.Schedule(millis, func(now, epochMillis int64) {
		notification.Notify(notificationName, fmt.Sprintf("Time completed: %s", title))
		if err := database.LogData(database.LogEntry{
			InitTime:    now,
			EndTime: epochMillis,
			TaskName:       title,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "error inserting data in database: %s", err)
			os.Exit(1)
		}
		wg.Done()
	})

	fmt.Printf("Alert scheduled for %s\n", timeArg)

	wg.Wait()
}
