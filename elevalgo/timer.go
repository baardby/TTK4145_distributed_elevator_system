package elevalgo

import (
	"time"
)

var (
	timerEndTime float64
	timerActive  bool
)

func timer_getWallTime() float64 {
	return float64(time.Now().UnixNano()) * 1e-9
}

func Timer_Start(duration float64) {
	timerEndTime = timer_getWallTime() + duration
	timerActive = true
}

func Timer_Stop() {
	timerActive = false
}

func Timer_TimedOut() bool {
	return timerActive && timer_getWallTime() > timerEndTime
}
