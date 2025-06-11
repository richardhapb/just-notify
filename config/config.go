package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ArgsCli struct {
	Time        string
	Notif       string
	Category    string
	Description string
	UseDatabase bool
	ConnString  string
	Unlimited   bool
	Headless    bool
	CsvPath     string
	Kill        bool
}

const (
	defaultCategory = "Unknown"
	defaultNotif    = "Time has been finalized"
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

func ParseArgs(cfg map[string]string) ArgsCli {
	var rawTime string
	var notif string
	var category string
	var description string
	var useDatabase bool
	var connString string
	var unlimited bool
	var headless bool
	var csvPath string
	var kill bool

	flag.StringVar(&rawTime, "t", "", "Time scheduled for the notification (e.g. <mm>m = Time and suffix \"m\" for minutes, or <hh:mm>Hour:minute")
	flag.StringVar(&category, "c", "", "Category: The category of the task to be executed during focus time. e.g. work.")
	flag.StringVar(&notif, "n", "", "Notification title: The title for the notification to be shown")
	flag.BoolVar(&useDatabase, "d", false, "Indicate whether a SQL database will be used")
	flag.StringVar(&connString, "s", "", "Connection string used to connect to the database; it only works if the database flag is enabled.")
	flag.StringVar(&description, "l", "", "Optional: Details of the task")
	flag.BoolVar(&kill, "k", false, "Kill a task, an category should be passed")
	flag.StringVar(&csvPath, "C", "", "CSV path; this will be ignored if the database is enabled, but can be used as a fallback.")
	flag.BoolVar(&unlimited, "u", false, "Unlimited time")
	flag.BoolVar(&headless, "H", false, "Headless, disable notifications")
	flag.Parse()

	if category == "" {
		category = cfg["DEFAULT_CATEGORY"]
	}

	if csvPath == "" {
		csvPath = cfg["CSV_PATH"]
	}

	if notif == "" {
		notif = cfg["DEFAULT_NOTIFICATION"]
	}

	if !useDatabase {
		useDatabase = cfg["USE_DATABASE"] == "true"
	}

	if !headless {
		headless = cfg["HEADLESS"] == "true"
	}

	if category == "" {
		category = defaultCategory
	}

	if notif == "" {
		notif = defaultNotif
	}

	return ArgsCli{
		Time:        rawTime,
		Notif:       notif,
		Category:    category,
		Description: description,
		UseDatabase: useDatabase,
		ConnString:  connString,
		Unlimited:   unlimited,
		Headless:    headless,
		CsvPath:     csvPath,
		Kill:        kill,
	}
}

func ValidateArgs(args *ArgsCli, cfg map[string]string) error {
	if !args.Kill && args.Time == "" && !args.Unlimited {
		return fmt.Errorf("\nERROR: Time argument is required")
	}

	if args.Category == "" {
		if args.Kill {
			return fmt.Errorf("\nERROR: A category is required to terminate the process")
		}
		return fmt.Errorf("\nERROR: Category is required")
	}

	if args.UseDatabase {
		if args.ConnString == "" && cfg["CONN"] == "" {
			return fmt.Errorf("Database enabled but no connection string provided")
		}
		if args.ConnString == "" {
			args.ConnString = cfg["CONN"]
		}
	}

	return nil
}
