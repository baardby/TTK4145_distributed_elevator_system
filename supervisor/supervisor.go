package supervisor

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
	"time"
)

type SupervisorEventType int

const (
	TimerElevatorTimeout SupervisorEventType = iota
	SupervisorHardwareFault
	SupervisorHardwareRecovered
)

type TimerEvent struct {
	Type       SupervisorEventType
	ElevatorID int
}

type timer struct {
	startTime time.Time
	active    bool
}

type elevatorTimers [N_ELEVATORS]timer

type movingTimer timer

type supervisor struct {
	elevatorTimers          elevatorTimers
	movingTimer             movingTimer
	stuckDetected           bool
	recoveryFromMovingStuck bool
	recoveryPrevFloor       int
	lastFloor               int
	obstruction             bool
	doorOpen                bool
}

func initSupervisor() supervisor {
	return supervisor{
		elevatorTimers: elevatorTimers{
			{startTime: time.Now(), active: false},
			{startTime: time.Now(), active: false},
			{startTime: time.Now(), active: false},
		},
		movingTimer: movingTimer{startTime: time.Now(), active: false},
		obstruction: false,
		doorOpen:    false,
	}
}

// func checkElevatorTimers returnerer ID til timeren som har gått ut, ellers -1
func (elevatorTimers *elevatorTimers) lostConnectionToElevator() int {
	for elevator := 0; elevator < N_ELEVATORS; elevator++ {
		if elevatorTimers[elevator].active && time.Since(elevatorTimers[elevator].startTime) > 5*time.Second {
			elevatorTimers[elevator].active = false
			return elevator
		}
	}
	return -1
}

func amIStuck(supervisor supervisor) bool {
	if supervisor.movingTimer.active && time.Since(supervisor.movingTimer.startTime) > 5*time.Second {
		return true
	}
	return false
}

func amIObstructed(supervisor supervisor) bool {
	return supervisor.obstruction && supervisor.doorOpen
}

func updateElevatorTimer(elevatorTimers *elevatorTimers, elevatorID int) {
	elevatorTimers[elevatorID].startTime = time.Now()
	elevatorTimers[elevatorID].active = true
}

func updateMovingTimer(supervisor *supervisor, elevator Elevator) {
	if supervisor.movingTimer.active {
		if elevator.Floor != supervisor.lastFloor {
			supervisor.movingTimer.startTime = time.Now()
		}
		if elevator.Behaviour != EB_Moving {
			supervisor.movingTimer.active = false
		}
	} else if elevator.Behaviour == EB_Moving {
		supervisor.movingTimer.startTime = time.Now()
		supervisor.movingTimer.active = true
	}
}

func handleElevatorUpdate(supervisor *supervisor, elevator Elevator) {
	updateMovingTimer(supervisor, elevator)

	supervisor.obstruction = elevator.Obstruction
	supervisor.doorOpen = (elevator.Behaviour == EB_DoorOpen)
	supervisor.lastFloor = elevator.Floor
}

func haveIRecovered(supervisor supervisor, elevator Elevator) bool {
	if supervisor.recoveryFromMovingStuck {
		if elevator.Floor != supervisor.recoveryPrevFloor {
			return true
		}
	} else {
		if !(elevator.Obstruction && elevator.Behaviour == EB_DoorOpen) {
			return true
		}
	}
	return false
}

func Supervisor(peerAliveCh <-chan int, updateElevatorEvt <-chan Elevator, TimerEventChan chan<- TimerEvent) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	sup := initSupervisor()
	// Lag løkke for å lage timers og holde styr på forskjellig stuff.
	for {
		select {
		case peerAlive := <-peerAliveCh:
			updateElevatorTimer(&sup.elevatorTimers, peerAlive)
		case elevator := <-updateElevatorEvt:
			if sup.stuckDetected {
				if haveIRecovered(sup, elevator) {
					sup.stuckDetected = false
					sup.recoveryFromMovingStuck = false
					TimerEventChan <- TimerEvent{Type: SupervisorHardwareRecovered}
				}
			}
			handleElevatorUpdate(&sup, elevator)

		case <-ticker.C:
			if id := sup.elevatorTimers.lostConnectionToElevator(); id != -1 {
				TimerEventChan <- TimerEvent{
					Type:       TimerElevatorTimeout,
					ElevatorID: id,
				}
			}
			if amIStuck(sup) && !sup.stuckDetected {
				sup.stuckDetected = true
				sup.recoveryFromMovingStuck = true
				sup.movingTimer.active = false
				sup.recoveryPrevFloor = sup.lastFloor
				TimerEventChan <- TimerEvent{Type: SupervisorHardwareFault}
			}
			if amIObstructed(sup) && !sup.stuckDetected {
				TimerEventChan <- TimerEvent{Type: SupervisorHardwareFault}
				sup.stuckDetected = true
			}
		}
	}
}
