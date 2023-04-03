// Module repeating repeats an action
package repeating

import "time"

// Repeat repears action every interval seconds
func Repeat(action func(), interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		<-ticker.C
		action()
	}
}
