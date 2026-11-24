package runtimeenv

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Load reads a KEY=value file; process environment variables take precedence.
func Load(path string) (func(string) string, error) {
	values := make(map[string]string)
	if path != "" {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for number := 1; scanner.Scan(); number++ {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, value, ok := strings.Cut(line, "=")
			if !ok || strings.TrimSpace(key) == "" {
				return nil, fmt.Errorf("invalid environment setting in %s on line %d", path, number)
			}
			values[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	return func(key string) string {
		if value, ok := os.LookupEnv(key); ok {
			return value
		}
		return values[key]
	}, nil
}
