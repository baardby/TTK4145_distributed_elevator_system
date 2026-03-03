package HealthTimers

type TimerEventType int

const (
	TimerElevatorTimeout TimerEventType = iota
	TimerMovementStuck
	TimerAcceptancetest
)

type TimerEvent struct {
	Type       TimerEventType
	ElevatorID string
}

type TimerCommandType int

const (
	ElevatorAlive TimerCommandType = iota
	Moving
	Stopped
	ArrivedAtFloor
)

type TimerCommand struct {
	Type       TimerCommandType
	ElevatorID string
}

func HealthTimers(TimerCommandChan <-chan TimerCommand, TimerEventChan chan<- TimerEvent) {
	// Lag løkke for å lage timers og holde styr på forskjellig stuff.
}
