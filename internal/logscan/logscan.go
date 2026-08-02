package logscan

import "strings"

// Match returns the first configured failure pattern found in logs.
func Match(logs string, patterns []string) (string, bool) {
	lowerLogs := strings.ToLower(logs)
	for _, pattern := range patterns {
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}
		if strings.Contains(lowerLogs, pattern) {
			return pattern, true
		}
	}
	return "", false
}
