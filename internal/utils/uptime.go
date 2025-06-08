package utils

import (
	"fmt"
	"time"
)

// startTime stores the time when the application started.
var startTime = time.Now()

// GetUptime calculates the time elapsed since the application started.
func GetUptime() string {
	uptimeDuration := time.Since(startTime)
	return fmt.Sprintf("%dd %dh %dm",
		int(uptimeDuration.Hours()/24),
		int(uptimeDuration.Hours())%24,
		int(uptimeDuration.Minutes())%60,
	)
}

// InitUptime (Optional) can be called once in main to explicitly set the start time
// if needed for specific scenarios, though `var startTime = time.Now()` is usually fine.
func InitUptime() {
	startTime = time.Now()
}

// You can add other utility functions here, e.g.:
// func ValidateEmail(email string) bool { ... }
