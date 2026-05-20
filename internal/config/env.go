package config

import (
	"bufio"
	"log"
	"os"
	"strings"
)

func LoadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		log.Println("Cannot Open Environment Variables")
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		value = strings.Trim(value, `"'`)

		if os.Getenv(key) == "" {
			if err = os.Setenv(key, value); err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
