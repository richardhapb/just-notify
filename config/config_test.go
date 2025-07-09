package config

import (
	"os"
	"testing"
)

func buildCliArgs() ArgsCli {
	return ArgsCli{
		Time:        "",
		Notif:       "",
		Category:    "testing",
		Description: "",
		UseDatabase: false,
		ConnString:  "",
		Unlimited:   false,
		Headless:    true,
	}

}

func TestValidateArgs(t *testing.T) {
	args := buildCliArgs()

	cfg := map[string]string{"CONN": "testing"}

	if err := ValidateArgs(&args, cfg); err == nil {
		t.Fatalf("Execution must fail; all fields are empty.")
	}

	args.Unlimited = true

	if err := ValidateArgs(&args, cfg); err != nil {
		t.Fatalf("The application must execute.")
	}

	args.Time = "1h"

	if err := ValidateArgs(&args, cfg); err != nil {
		t.Fatalf("The application must execute.")
	}

	args.UseDatabase = true

	if err := ValidateArgs(&args, map[string]string{}); err == nil {
		t.Fatalf("Execution must fail; connection string is required when database is enabled.")
	}

	if err := ValidateArgs(&args, cfg); err != nil {
		t.Fatalf("The application must execute. CONN config is present.")
	}

	args.Category = ""

	if err := ValidateArgs(&args, cfg); err == nil {
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
	parsedArgs, err := ParseArgs(cfg)

	if err != nil {
		t.Fatalf("Error parsing arguments: %s", err)
	}

	if parsedArgs.Time != expectedTime {
		t.Fatalf("Time expected: %s, received %s", expectedTime, parsedArgs.Time)
	}

	if parsedArgs.Category != cfg["DEFAULT_CATEGORY"] {
		t.Fatalf("Category expected: %s, received %s", cfg["DEFAULT_CATEGORY"], parsedArgs.Category)
	}

	if parsedArgs.Notif != cfg["DEFAULT_NOTIFICATION"] {
		t.Fatalf("Notification expected: %s, received %s", cfg["DEFAULT_NOTIFICATION"], parsedArgs.Notif)
	}

	if !parsedArgs.UseDatabase {
		t.Fatalf("Use database expected: %v, received %v", true, parsedArgs.UseDatabase)
	}

	if !parsedArgs.Headless {
		t.Fatalf("Headless expected: %v, received %v", true, parsedArgs.Headless)
	}
}
