package ui

import (
	"fmt"
	"strings"
	"time"
)

func ProgressBar(closeSignal chan bool, init, end int64) {
	if init >= end {
		return
	}

	const width = 50
	bar := fmt.Sprintf("[%s]", strings.Repeat(" ", width))
	duration := time.Duration(end-init) * time.Millisecond
	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()

	fmt.Println()
	start := time.Now()

	for {

		select {
		case <-closeSignal:
			// Clean up the progress bar and exit
			fmt.Printf("\r%s 100.0%%\n\n",
				bar[:1]+strings.Repeat("█", width)+bar[width+1:])
			return
		case <-ticker.C:
			elapsed := time.Since(start)
			progress := float64(elapsed) / float64(duration)

			if progress >= 1.0 {
				fmt.Printf("\r%s 100.0%%\n",
					bar[:1]+strings.Repeat("█", width)+bar[width+1:])
				return
			}

			filled := int(progress * float64(width))
			progressBar := bar[:1] + strings.Repeat("█", filled) +
				strings.Repeat("░", width-filled) + bar[width+1:]

			fmt.Printf("\r%s %.1f%%", progressBar, progress*100)
		}
	}
}
