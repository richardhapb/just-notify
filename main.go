package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("\tjn [Time option]")
	fmt.Println()
	fmt.Println("Time options:")
	fmt.Println("\t<ss>s\tTime and suffix \"s\": seconds")
	fmt.Println("\t<mm>m\tTime and suffix \"m\": minutes")
	fmt.Println("\t<hh>h\tTime and suffix \"h\": hours")
	fmt.Println("\t<hh:mm>\tHour:minute")
	fmt.Println()
}

func main() {
	args := os.Args

	if len(args) < 2 {
		fmt.Println("\nERROR: Time argument is required")
		printUsage()
		os.Exit(1)
	}

	timeArg := args[1]
	taskName := timeArg
	title := "Unknown"

	if len(args) > 2 {
		taskName = args[2]
		title = taskName
	}

	millis, err := getTime(timeArg)

	if err != nil {
		fmt.Printf("Error scheduling task: %s\n", err.Error())
		os.Exit(1)
	}

	wg.Add(1)

	schedule(millis, func(now, epochMillis int64) {
		notify(title, fmt.Sprintf("Time completed: %s", taskName))
		logData([][]string{{strconv.Itoa(int(now)), strconv.Itoa(int(epochMillis)), title}})
		wg.Done()
	})

	fmt.Printf("Alert scheduled for %s\n", timeArg)

	wg.Wait()
}

func schedule(epochMillis int64, action func(int64, int64)) {
	now := time.Now().UnixMilli()
	delayMillis := epochMillis - now

	fmt.Printf("Scheduling task to %d seconds later\n", delayMillis/1000)

	if delayMillis < 0 {
		fmt.Println("epochMillis is in the past in schedule function")
		return
	}

	go progressBar(now, epochMillis)

	go func() {
		time.Sleep(time.Duration(delayMillis) * time.Millisecond)
		action(now, epochMillis)
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func logData(data [][]string) error {
	config := loadConfig()

	filePath := config["CSV_PATH"]

	// Default path
	if filePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		filePath = filepath.Join(home, ".jn.csv")
	}

	// Check if file exists to determine if headers are needed
	needsHeader := !fileExists(filePath)

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers if new file
	if needsHeader {
		headers := []string{"init_time_ms", "end_time_ms", "task_name"}
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("error writing headers: %w", err)
		}
	}

	// Write the actual data
	return writer.WriteAll(data)
}

func progressBar(init, end int64) {
	if init >= end {
		return
	}

	const width = 50
	bar := fmt.Sprintf("[%s]", strings.Repeat(" ", width))

	fmt.Println()
	for {
		now := time.Now().UnixMilli()
		progress := float64(now-init) / float64(end-init)

		filled := int(progress * float64(width))
		progressBar := bar[:1] + strings.Repeat("█", filled) +
			strings.Repeat("░", width-filled) + bar[width+1:]

		// Avoid round issues
		percentage := min(progress*100, 100.0)

		fmt.Printf("\r%s %.1f%%", progressBar, percentage)

		if progress >= 1.0 {
			break
		}

		time.Sleep(time.Millisecond * 500)
	}
	fmt.Println() // Add newline at the end
}

func loadConfig() map[string]string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home dir:", err)
		return nil
	}

	configPath := filepath.Join(home, ".jnconfig")

	content, err := os.ReadFile(configPath)
	if err != nil {
		// No config file found, fallback
		return nil
	}

	lines := strings.Split(string(content), "\n")
	config := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		config[key] = val
	}

	return config
}
