package notification

import (
	"fmt"
	"just-notify/ui"
	"os/exec"
	"runtime"
	"time"
)

func Notify(title, message string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("terminal-notifier", "-title", title, "-message", message).Run()
	case "linux":
		return exec.Command("notify-send", title, message).Run()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func Schedule(enableProgressBar bool, closeSignal chan bool, epochMillis int64, action func(int64, int64)) {
	now := time.Now().UnixMilli()

	if epochMillis != 0 && epochMillis < now {
		fmt.Println("Warning: Target time is in the past")
		return
	}

	// Calculate total duration
	duration := time.Duration(epochMillis-now) * time.Millisecond
	fmt.Printf("Scheduling task for %v from now\n", duration.Round(time.Second))

	if enableProgressBar {
		go ui.ProgressBar(closeSignal, now, epochMillis)
	} else {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-closeSignal:
				action(now, time.Now().UnixMilli())
				return
			case t := <-ticker.C:
				current := t.UnixMilli()
				elapsed := time.Since(time.UnixMilli(now)).Round(time.Second)
				hours := int(elapsed.Hours())
				minutes := int(elapsed.Minutes()) % 60
				seconds := int(elapsed.Seconds()) % 60
				fmt.Printf("\rTime elapsed: %02d:%02d:%02d", hours, minutes, seconds)

				if epochMillis != 0 && current >= epochMillis {
					fmt.Println() // Add newline before exiting
					action(now, current)
					return
				}
			}
		}
	}

	// For progress bar mode, wait for completion before running action
	if enableProgressBar {
		time.Sleep(duration)
		action(now, time.Now().UnixMilli())
	}
}
