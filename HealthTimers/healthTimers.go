package HealthTimers

import (
	"time"
)

const N_ELEVATORS = 3

type TimerEventType int

const (
	TimerElevatorTimeout TimerEventType = iota
	TimerMovementStuck
	TimerAcceptancetest
)

type TimerEvent struct {
	Type       TimerEventType
	ElevatorID int
}

type TimerCommandType int

const (
	ElevatorAlive TimerCommandType = iota
	MovingFromStopped
	Stopped
	DrovePastFloor
)

type TimerCommand struct {
	Type       TimerCommandType
	ElevatorID int
}

type countdown struct {
	startTime time.Time
	active    bool
}

type elevatorTimers [N_ELEVATORS]countdown

func (elevatorTimers *elevatorTimers) checkElevatorTimers() int { //sender -1 om ingen timere har gått ut
	for elevator := 0; elevator < N_ELEVATORS; elevator++ {
		if elevatorTimers[elevator].active && time.Since(elevatorTimers[elevator].startTime) > 5*time.Second {
			return elevator
		}
	}
	return -1
}

type movingTimer countdown

func (movingTimer *movingTimer) amIStuck() bool {
	if movingTimer.active && time.Since(movingTimer.startTime) > 5*time.Second {
		return true
	}
	return false
}

func (timerCommand *TimerCommand) perform(elevatorTimers *elevatorTimers, movingTimer *movingTimer) {
	switch timerCommand.Type {
	case ElevatorAlive:
		elevatorTimers[timerCommand.ElevatorID].startTime = time.Now()
		elevatorTimers[timerCommand.ElevatorID].active = true
	case MovingFromStopped:
		movingTimer.startTime = time.Now()
		movingTimer.active = true
	case Stopped:
		movingTimer.active = false
	case DrovePastFloor:
		movingTimer.startTime = time.Now()
	}
}

func HealthTimers(TimerCommandChan <-chan TimerCommand, TimerEventChan chan<- TimerEvent) {
	ticker := time.NewTicker(100 * time.Millisecond)

	movingTimer := movingTimer{startTime: time.Now(), active: false}

	elevatorTimers := elevatorTimers{
		{startTime: time.Now(), active: false},
		{startTime: time.Now(), active: false},
		{startTime: time.Now(), active: false},
	}
	// Lag løkke for å lage timers og holde styr på forskjellig stuff.
	for {
		select {
		case timerCommand := <-TimerCommandChan:
			timerCommand.perform(&elevatorTimers, &movingTimer)

		case <-ticker.C:
			if id := elevatorTimers.checkElevatorTimers(); id != -1 {
				TimerEventChan <- TimerEvent{
					Type:       TimerElevatorTimeout,
					ElevatorID: id,
				}
			}
			if movingTimer.amIStuck() {
				TimerEventChan <- TimerEvent{Type: TimerMovementStuck}
			}

			//LAG EN ACCEPTANCE TEST
		}
	}
}
