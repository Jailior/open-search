package utils

import (
	"fmt"
	"log"
	"time"
)

// Retry a function N times with exponential backoff
// Use label for logging
func RetryWithBackoff(operation func() error, maxRetries int, label string) error {
	var err error
	backoff := 500 * time.Millisecond

	for i := 1; i <= maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		log.Printf("[Retry][%s] Attempt %d failed: %v", label, i, err)
		time.Sleep(backoff)
		backoff *= 2
	}

	return fmt.Errorf("Retry failed for %s failed after %d attempts: %w", label, maxRetries, err)
}
