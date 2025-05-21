package config

import (
	"os"
	"path/filepath"
	"fmt"
	"strings"
)

func LoadConfig() map[string]string {
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

