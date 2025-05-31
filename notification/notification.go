package notification

import (
	"just-notify/ui"
	"os/exec"
	"runtime"
	"time"
	"fmt"
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



func Schedule(closeSignal chan bool, epochMillis int64, action func(int64, int64)) {
	now := time.Now().UnixMilli()
	delayMillis := epochMillis - now

	fmt.Printf("Scheduling task to %d seconds later\n", delayMillis/1000)

	if delayMillis < 0 {
		fmt.Println("epochMillis is in the past in schedule function")
		return
	}

	ui.ProgressBar(closeSignal, now, epochMillis)

	action(now, epochMillis)
}
