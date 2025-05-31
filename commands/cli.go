package commands

import (
	"fmt"
	"strings"
	"strconv"
	"time"
)

func GetTime(timeArg string) (int64, error) {

	result, err := int64(0), fmt.Errorf("Unexpected time argument: %s", timeArg)

	if len(timeArg) < 2 {
		return result, err
	}

	if strings.Contains(timeArg, ":") {

		parts := strings.Split(timeArg, ":")

		if len(parts) != 2 {
			return result, fmt.Errorf("Incorrect time format: %v", parts)
		}

		h, err := strconv.Atoi(parts[0])

		if err != nil {
			return result, err
		}

		m, err := strconv.Atoi(parts[1])

		if err != nil {
			return result, err
		}

		now := time.Now()

		target := time.Date(
			now.Year(), now.Month(), now.Day(),
			h, m, 0, 0,
			now.Location(),
		)

		if target.Before(now) {
			target = target.Add(24 * time.Hour)
		}

		result = target.UnixMilli()

	} else {

		suffix := timeArg[len(timeArg)-1:]

		if !strings.Contains("hms", suffix) {
			return result, fmt.Errorf("Incorrect time suffix: \"%s\"", suffix)
		}

		numberStr := timeArg[:len(timeArg)-1]
		number, err := strconv.Atoi(numberStr)

		if err != nil {
			return result, err
		}

		switch suffix {
		case "h":
			result = time.Now().Add(time.Duration(number) * time.Hour).UnixMilli()
		case "m":
			result = time.Now().Add(time.Duration(number) * time.Minute).UnixMilli()
		case "s":
			result = time.Now().Add(time.Duration(number) * time.Second).UnixMilli()
		}
	}

	return result, nil
}

