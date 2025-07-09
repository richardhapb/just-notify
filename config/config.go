package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fred1268/go-clap/clap"
)

type ArgsCli struct {
	Time        string `clap:"--time,-t"`
	Notif       string `clap:"--notif,-n"`
	Category    string `clap:"--cat,-c"`
	Description string `clap:"--description,-l"`
	UseDatabase bool   `clap:"--database,-d"`
	ConnString  string `clap:"--conn,-s"`
	Unlimited   bool   `clap:"--unlimited,-u"`
	Headless    bool   `clap:"--headless,-H"`
	CsvPath     string `clap:"--csvpath,-C"`
	Kill        bool   `clap:"--kill,-k"`
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

func ParseArgs(cfg map[string]string) (*ArgsCli, error) {
	args := os.Args
	cli := &ArgsCli{}

	var err error
	var results *clap.Results

	if results, err = clap.Parse(args[1:], cli); err != nil {
		if len(results.Mandatory) > 0 {
			for i, mandatory := range results.Mandatory {
				// For some reason each mandatory is repeated
				if i%2 != 0 {
					fmt.Printf("Error: %s is required\n\n", mandatory)
				}
			}
		}

		PrintUsage()

		os.Exit(1)
	}

	if cli.Category == "" {
		cli.Category = cfg["DEFAULT_CATEGORY"]
	}

	if cli.CsvPath == "" {
		cli.CsvPath = cfg["CSV_PATH"]
	}

	if cli.Notif == "" {
		cli.Notif = cfg["DEFAULT_NOTIFICATION"]
	}

	if !cli.UseDatabase {
		cli.UseDatabase = cfg["USE_DATABASE"] == "true"
	}

	if !cli.Headless {
		cli.Headless = cfg["HEADLESS"] == "true"
	}

	if cli.Category == "" {
		cli.Category = defaultCategory
	}

	if cli.Notif == "" {
		cli.Notif = defaultNotif
	}

	return cli, nil
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

func PrintUsage() {
	fmt.Println("Usage: program [options]")
	fmt.Println("\nOptions:")
	fmt.Printf("  -t, --time         Time scheduled for the notification (required unless --unlimited or --kill)\n")
	fmt.Printf("                     Format: <mm>m for minutes, or <hh:mm> for hour:minute\n")
	fmt.Printf("  -c, --cat         Category of the task (e.g., 'work')\n")
	fmt.Printf("  -n, --notif       Notification title to be shown\n")
	fmt.Printf("  -l, --description Optional details of the task\n")
	fmt.Printf("  -d, --database    Enable SQL database usage\n")
	fmt.Printf("  -s, --conn        Database connection string (required if --database is set)\n")
	fmt.Printf("  -u, --unlimited   Set unlimited time\n")
	fmt.Printf("  -H, --headless    Disable notifications\n")
	fmt.Printf("  -C, --csvpath     CSV file path (ignored if database is enabled)\n")
	fmt.Printf("  -k, --kill        Kill a task (requires category)\n")
	fmt.Println("\nConfiguration:")
	fmt.Printf("  Config file: ~/.jnconfig\n")
	fmt.Printf("  Supported config keys: DEFAULT_CATEGORY, CSV_PATH, DEFAULT_NOTIFICATION,\n")
	fmt.Printf("                        USE_DATABASE, HEADLESS, CONN\n")
	fmt.Println()
}
