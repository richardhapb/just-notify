package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

func notify(title, message string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("terminal-notifier", "-title", title, "-message", message).Run()
	case "linux":
		return exec.Command("notify-send", title, message).Run()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("\tjn <hh:mm>")
	fmt.Println("\tjn <ss>s | <mm>m | <hh>h")
}

func main() {
	args := os.Args

	if len(args) < 2 {
		log.Fatalf("Time argument is required")
		printUsage()
	}

	timeArg := args[1]
	taskName := timeArg
	title := "Time completed"

	if len(args) > 2 {
		taskName = args[2]
		title = taskName
	}

	millis, err := getTime(timeArg)

	if err != nil {
		log.Fatalf("Error scheduling task: %s", err.Error())
	}

	wg.Add(1)
	schedule(millis, func() {
		notify(title, fmt.Sprintf("Time completed: %s", taskName))
		wg.Done()
	})

	log.Printf("Alert scheduled for %s\n", timeArg)

	wg.Wait()
}

func schedule(epochMillis int64, action func()) {
	delayMillis := epochMillis - time.Now().UnixMilli()

	log.Println(fmt.Sprintf("Scheduling task to %d seconds later", delayMillis/1000))

	if delayMillis < 0 {
		log.Println("epochMillis is in the past in schedule function")
		return
	}

	go func() {
		time.Sleep(time.Duration(delayMillis) * time.Millisecond)
		action()
	}()
}

func getTime(timeArg string) (int64, error) {

	result, err := int64(0), fmt.Errorf("Unexpected time argument: %s", timeArg)

	if len(timeArg) < 2 {
		return result, err
	}

	if strings.Contains(timeArg, ":") {

		parts := strings.Split(timeArg, ":")

		if len(parts) != 2 {
			return result, fmt.Errorf("Incorrect time format: %v", parts)
		}

		h, err := strconv.Atoi(parts[0])

		if err != nil {
			return result, err
		}

		m, err := strconv.Atoi(parts[1])

		if err != nil {
			return result, err
		}

		now := time.Now()

		target := time.Date(
			now.Year(), now.Month(), now.Day(),
			h, m, 0, 0,
			now.Location(),
		)

		if target.Before(now) {
			target = target.Add(24 * time.Hour)
		}

		result = target.UnixMilli()

	} else {

		suffix := timeArg[len(timeArg)-1:]

		if !strings.Contains("hms", suffix) {
			return result, fmt.Errorf("Incorrect time suffix: \"%s\"", suffix)
		}

		numberStr := timeArg[:len(timeArg)-1]
		number, err := strconv.Atoi(numberStr)

		if err != nil {
			return result, err
		}

		switch suffix {
		case "h":
			result = time.Now().Add(time.Duration(number) * time.Hour).UnixMilli()
		case "m":
			result = time.Now().Add(time.Duration(number) * time.Minute).UnixMilli()
		case "s":
			result = time.Now().Add(time.Duration(number) * time.Second).UnixMilli()
		}
	}

	return result, nil
}
