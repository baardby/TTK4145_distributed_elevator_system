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

type obstructionTimer timer

type supervisor struct {
	elevatorTimers          elevatorTimers
	movingTimer             movingTimer
	obstructionTimer        obstructionTimer
	stuckDetected           bool
	recoveryFromMovingStuck bool
	recoveryPrevFloor       int
	lastFloor               int
	obstructed              bool
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
		obstructionTimer:        obstructionTimer{startTime: time.Now(), active: false},
		stuckDetected:           false,
		recoveryFromMovingStuck: false,
		lastFloor:               -1,
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
	if supervisor.movingTimer.active && time.Since(supervisor.movingTimer.startTime) > 4*time.Second { // !!! Spec / FAT says 4 seconds
		return true
	}
	return false
}

func obstructionTimedOut(supervisor supervisor) bool {
	if supervisor.obstructionTimer.active && time.Since(supervisor.obstructionTimer.startTime) > 8*time.Second {
		return true
	}
	return false
}

func updateElevatorTimer(elevatorTimers *elevatorTimers, elevatorID int) {
	elevatorTimers[elevatorID].startTime = time.Now()
	elevatorTimers[elevatorID].active = true
}

func updateMovingTimer(supervisor *supervisor, elevator Elevator) {
	if supervisor.movingTimer.active {
		if elevator.Floor != supervisor.lastFloor {
			supervisor.movingTimer.startTime = time.Now()
			supervisor.lastFloor = elevator.Floor
		}
		if elevator.Behaviour != EB_Moving {
			supervisor.movingTimer.active = false
		}
	} else if elevator.Behaviour == EB_Moving {
		supervisor.movingTimer.startTime = time.Now()
		supervisor.movingTimer.active = true
		supervisor.lastFloor = elevator.Floor
	}
}

func updateObstructionTimer(supervisor *supervisor, elevator Elevator) {
	if elevator.Obstruction && elevator.Behaviour == EB_DoorOpen {
		if !supervisor.obstructionTimer.active {
			supervisor.obstructionTimer.active = true
			supervisor.obstructionTimer.startTime = time.Now()
		}
	} else {
		supervisor.obstructionTimer.active = false
	}
}

func handleElevatorUpdate(supervisor *supervisor, elevator Elevator) {
	updateMovingTimer(supervisor, elevator)
	updateObstructionTimer(supervisor, elevator)
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

func Supervisor(
	peerAliveCh <-chan int,
	updateElevatorEvt <-chan Elevator,
	SupervisorEventChan chan<- SupervisorEvent) {

	// Wait for elevator to find floor
	elevatorStartState := <-updateElevatorEvt

	healthCheckTicker := time.NewTicker(100 * time.Millisecond)
	defer healthCheckTicker.Stop()

	sup := initSupervisor()
	sup.lastFloor = elevatorStartState.Floor

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

		case <-healthCheckTicker.C:
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
			if obstructionTimedOut(sup) && !sup.stuckDetected {
				sup.stuckDetected = true
				sup.recoveryFromMovingStuck = false
				sup.obstructionTimer.active = false
				SupervisorEventChan <- SupervisorEvent{Type: SupervisorHardwareFault}
			}
		}
	}
}
