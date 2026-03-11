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

type SupervisorEvent struct {
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
		movingTimer:             movingTimer{startTime: time.Now(), active: false},
		stuckDetected:           false,
		recoveryFromMovingStuck: false,
		obstruction:             false,
		doorOpen:                false,

		lastFloor: -1,
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

func Supervisor(peerAliveCh <-chan int, updateElevatorEvt <-chan Elevator, SupervisorEventChan chan<- SupervisorEvent) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	sup := initSupervisor()

	for {
		select {
		case peerAlive := <-peerAliveCh:
			updateElevatorTimer(&sup.elevatorTimers, peerAlive)
		case elevator := <-updateElevatorEvt:
			if sup.stuckDetected {
				if haveIRecovered(sup, elevator) {
					sup.stuckDetected = false
					sup.recoveryFromMovingStuck = false
					SupervisorEventChan <- SupervisorEvent{Type: SupervisorHardwareRecovered}
				}
			}
			handleElevatorUpdate(&sup, elevator)

		case <-ticker.C:
			if id := sup.elevatorTimers.lostConnectionToElevator(); id != -1 {
				SupervisorEventChan <- SupervisorEvent{
					Type:       TimerElevatorTimeout,
					ElevatorID: id,
				}
			}
			if amIStuck(sup) && !sup.stuckDetected {
				sup.stuckDetected = true
				sup.recoveryFromMovingStuck = true
				sup.movingTimer.active = false
				sup.recoveryPrevFloor = sup.lastFloor
				SupervisorEventChan <- SupervisorEvent{Type: SupervisorHardwareFault}
			}
			if amIObstructed(sup) && !sup.stuckDetected {
				SupervisorEventChan <- SupervisorEvent{Type: SupervisorHardwareFault}
				sup.stuckDetected = true
			}
		}
	}
}
