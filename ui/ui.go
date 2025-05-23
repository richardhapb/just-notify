package ui

import (
	"fmt"
	"strings"
	"time"
)

func ProgressBar(init, end int64) {
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

