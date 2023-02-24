package agent

import "time"

func Repeat(action func(), interval time.Duration) *time.Ticker {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			action()
		}
	}()
	return ticker
}
