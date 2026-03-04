package HealthTimers

import (
	"distributed_elevator/elevalgo"
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

type timer struct {
	startTime time.Time
	active    bool
}

type elevatorTimers [N_ELEVATORS]timer

type movingTimer timer

// func checkElevatorTimers returnerer ID til timeren som har gått ut, -1 ellers
func (elevatorTimers *elevatorTimers) checkElevatorTimers() int {
	for elevator := 0; elevator < N_ELEVATORS; elevator++ {
		if elevatorTimers[elevator].active && time.Since(elevatorTimers[elevator].startTime) > 5*time.Second {
			return elevator
		}
	}
	return -1
}

func (movingTimer *movingTimer) amIStuck() bool {
	if movingTimer.active && time.Since(movingTimer.startTime) > 5*time.Second {
		return true
	}
	return false
}

func updateElevatorTimer(elevatorTimers *elevatorTimers, elevatorID int) {
	elevatorTimers[elevatorID].startTime = time.Now()
	elevatorTimers[elevatorID].active = true
}

func updateMovingTimer(movingTimer *movingTimer, elevator elevalgo.Elevator) {
	if movingTimer.active {
		if elevator.Floor != -1 {
			movingTimer.startTime = time.Now()
		}
		if elevator.Behaviour != elevalgo.EB_Moving {
			movingTimer.active = false
		}
	} else {
		if elevator.Behaviour == elevalgo.EB_Moving {
			movingTimer.startTime = time.Now()
			movingTimer.active = true
		}
	}

}

func HealthTimers(peerAliveCh <-chan int, updateElevatorEvt <-chan elevalgo.Elevator, TimerEventChan chan<- TimerEvent) {
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
		case peerAlive := <-peerAliveCh:
			updateElevatorTimer(&elevatorTimers, peerAlive)
		case elevator := <-updateElevatorEvt:
			updateMovingTimer(&movingTimer, elevator)
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
