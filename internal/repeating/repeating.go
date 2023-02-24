package repeating

import "time"

func Repeat(action func(), interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		<-ticker.C
		action()
	}
}
