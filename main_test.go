package main

import (
	"os"
	"testing"
)

func buildCliArgs() ArgsCli {
	return ArgsCli{
		time:        "",
		notif:       "",
		category:    "testing",
		description: "",
		useDatabase: false,
		connString:  "",
		unlimited:   false,
		headless:    true,
	}

}

func TestValidateArgs(t *testing.T) {
	args := buildCliArgs()

	cfg := map[string]string{"CONN": "testing"}

	if err := validateArgs(&args, cfg); err == nil {
		t.Fatalf("Execution must fail; all fields are empty.")
	}

	args.unlimited = true

	if err := validateArgs(&args, cfg); err != nil {
		t.Fatalf("The application must execute.")
	}

	args.time = "1h"

	if err := validateArgs(&args, cfg); err != nil {
		t.Fatalf("The application must execute.")
	}

	args.useDatabase = true

	if err := validateArgs(&args, map[string]string{}); err == nil {
		t.Fatalf("Execution must fail; connection string is required when database is enabled.")
	}

	if err := validateArgs(&args, cfg); err != nil {
		t.Fatalf("The application must execute. CONN config is present.")
	}

	args.category = ""

	if err := validateArgs(&args, cfg); err == nil {
		t.Fatalf("Execution must fail; category is required.")
	}
}

func TestParseArgs(t *testing.T) {
	cfg := map[string]string{
		"CONN":                 "testing-conn",
		"DEFAULT_CATEGORY":     "testing-category",
		"DEFAULT_NOTIFICATION": "testing-notification",
		"USE_DATABASE":         "true",
		"HEADLESS":             "true",
	}

	expectedTime := "1h"
	os.Args[1] = "-t"
	os.Args[2] = expectedTime
	parsedArgs := parseArgs(cfg)

	if parsedArgs.time != expectedTime {
		t.Fatalf("Time expected: %s, received %s", expectedTime, parsedArgs.time)
	}

	if parsedArgs.category != cfg["DEFAULT_CATEGORY"] {
		t.Fatalf("Category expected: %s, received %s", cfg["DEFAULT_CATEGORY"], parsedArgs.category)
	}

	if parsedArgs.notif != cfg["DEFAULT_NOTIFICATION"] {
		t.Fatalf("Notification expected: %s, received %s", cfg["DEFAULT_NOTIFICATION"], parsedArgs.notif)
	}

	if !parsedArgs.useDatabase {
		t.Fatalf("Use database expected: %v, received %v", true, parsedArgs.useDatabase)
	}

	if !parsedArgs.headless {
		t.Fatalf("Headless expected: %v, received %v", true, parsedArgs.headless)
	}
}
