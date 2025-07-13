# Just-Notify

Just-Notify is a lightweight application designed for scheduling notifications, making it ideal for Pomodoro sessions, reminders, or focus tasks. It supports flexible time formats, categories, and descriptions, and integrates with both CSV and SQL databases for logging.

---

## Features

- **Flexible Scheduling**: Set notifications using relative time (`30m`, `1h`) or absolute time (`12:30`).
- **Task Categorization**: Assign categories to tasks for better organization.
- **Persistent Logging**: Log tasks to a CSV file or SQL database for tracking and analysis.
- **Cross-Platform Notifications**: Supports macOS (`terminal-notifier`) and Linux (`notify-send`).
- **Headless Mode**: Disable notifications for silent operation.
- **Progress Bar**: Visualize time remaining for scheduled tasks.
- **Kill Tasks**: Terminate tasks by category.

---

## Installation

### Prerequisites

- **macOS**: Install `terminal-notifier` via Homebrew:
  ```bash
  brew install terminal-notifier
  ```
- **Linux**: Ensure `notify-send` is available (part of `libnotify-bin`):
  ```bash
  sudo apt install libnotify-bin
  ```

### Build from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/richardhapb/just-notify.git
   cd just-notify
   ```

2. Build the application:
   ```bash
   go build -o jn
   ```

3. Move the binary to a directory in your PATH:
   ```bash
   mv jn /usr/local/bin/
   ```

---

## Usage

### Basic Examples

- Schedule a notification in 30 minutes:
  ```bash
  jn -t 30m -c "Focus" -n "Break Time"
  ```

- Schedule a notification in 1 hour with a description:
  ```bash
  jn -t 1h -c "Work" -n "Call Customer" -l "Discuss project updates"
  ```

- Schedule a notification at 12:30 PM:
  ```bash
  jn -t 12:30 -c "Personal" -n "Send Email" -l "Email bank about loan details"
  ```

### Advanced Options

- Enable database logging:
  ```bash
  jn -t 45m -c "Study" -d -s "postgresql://user:password@localhost/dbname"
  ```

- Use a custom CSV file for logging:
  ```bash
  jn -t 20m -c "Exercise" -C "/path/to/log.csv"
  ```

- Run in headless mode (no notifications):
  ```bash
  jn -t 1h -c "Silent Task" -H
  ```

- Kill a task by category:
  ```bash
  jn -k -c "Focus"
  ```

---

## Configuration

Just-Notify supports a configuration file located at `~/.jnconfig`. This file allows you to set default values for various options.

### Example Configuration

```ini
DEFAULT_CATEGORY=Work
DEFAULT_NOTIFICATION=Task Completed
CSV_PATH=/path/to/log.csv
USE_DATABASE=true
HEADLESS=false
CONN=postgresql://user:password@localhost/dbname
```

---

## Logging

### CSV Logging

Tasks are logged in a CSV file with the following columns:
- `init_time_ms`: Start time in milliseconds.
- `end_time_ms`: End time in milliseconds (if completed).
- `category`: Task category.
- `description`: Task description.

### SQL Logging

Tasks are stored in a database table with the following schema:
```sql
CREATE TABLE logs (
    id SERIAL PRIMARY KEY,
    init_time_ms BIGINT NOT NULL,
    end_time_ms BIGINT,
    category TEXT NOT NULL,
    description TEXT,
    UNIQUE (init_time_ms, category)
);
```

---

## Development

### Run Tests

To run tests for the project:
```bash
go test ./...
```

### Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

---

## License

This project is licensed under the MIT License.


