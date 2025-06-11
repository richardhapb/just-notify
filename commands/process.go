package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const pidPathPattern = "/tmp/jn.%s.pid"

func KillProcess(category string) (int, error) {
	pidFile := fmt.Sprintf(pidPathPattern, strings.TrimSpace(category))
	
	pid, err := os.ReadFile(pidFile)
	if err != nil {
		return -1, fmt.Errorf("reading PID file: %w", err)
	}

	pidNum, err := strconv.Atoi(strings.TrimSpace(string(pid)))
	if err != nil {
		return -1, fmt.Errorf("parsing PID: %w", err)
	}

	if err := exec.Command("pkill", "-SIGTERM", "-F", pidFile).Run(); err != nil {
		return -1, fmt.Errorf("killing process: %w", err)
	}

	// Clean up PID file after successful termination
	if err := os.Remove(pidFile); err != nil {
		return pidNum, fmt.Errorf("removing PID file: %w", err)
	}

	return pidNum, nil
}

func StorePID(category string) error {
	pidFile := fmt.Sprintf(pidPathPattern, strings.TrimSpace(category))
	
	pid := os.Getpid()
	
	err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0640)
	if err != nil {
		return fmt.Errorf("writing PID file: %w", err)
	}
	
	return nil
}

