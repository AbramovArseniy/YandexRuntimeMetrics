// Package repeating repeats an action
package repeating

import (
	"os"
	"time"
)

// Repeat repears action every interval seconds
func Repeat(sigs chan os.Signal, action func(), interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-sigs:
			return
		case <-ticker.C:
			action()
		}
	}
}
