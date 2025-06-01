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

	if epochMillis != 0 {
		delayMillis := epochMillis - now
		if delayMillis < 0 {
			fmt.Println("epochMillis is in the past in schedule function")
			return
		}

		fmt.Printf("Scheduling task to %d seconds later\n", delayMillis/1000)
	}

	if enableProgressBar {
		ui.ProgressBar(closeSignal, now, epochMillis)
	} else {
		exit := false
		for {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-closeSignal:
					exit = true
				case <-ticker.C:
					elapsed := time.Since(time.UnixMilli(now)).Round(time.Second)
					hours := int(elapsed.Hours())
					minutes := int(elapsed.Minutes()) % 60
					seconds := int(elapsed.Seconds()) % 60
					fmt.Printf("\rTime elapsed: %02d:%02d:%02d", hours, minutes, seconds)
				}
				if exit {
					break
				}

			}
			if exit {
				break
			}
		}
	}

	action(now, epochMillis)
}
