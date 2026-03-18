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
	elevatorTimers         elevatorTimers
	movingTimer            movingTimer
	obstructionTimer       obstructionTimer
	stuckDetected          bool
	recoveryFromImmobility bool //To distinguish between immobility and obstruction, which has different recovery methods
	recoveryPrevFloor      int
	lastFloor              int
}

func Supervisor(
	peerAlive <-chan int,
	updateMyElevator <-chan Elevator,
	supervisorEventChan chan<- SupervisorEvent) {

	// Wait for elevator to find floor
	sup := initSupervisor()
	elevatorStartState := <-updateMyElevator
	handleElevatorUpdate(&sup, elevatorStartState)

	healthCheckTicker := time.NewTicker(100 * time.Millisecond)
	defer healthCheckTicker.Stop()

	for {
		select {
		case peerAlive := <-peerAlive:
			updateElevatorTimer(&sup.elevatorTimers, peerAlive)

		case elevator := <-updateMyElevator:
			if sup.stuckDetected {
				if haveIRecovered(sup, elevator) {
					sup.stuckDetected = false
					sup.recoveryFromImmobility = false
					supervisorEventChan <- SupervisorEvent{Type: SupervisorHardwareRecovered}
				}
			}
			handleElevatorUpdate(&sup, elevator)

		case <-healthCheckTicker.C:
			if id := sup.elevatorTimers.lostConnectionToElevator(); id != -1 {
				supervisorEventChan <- SupervisorEvent{
					Type:       TimerElevatorTimeout,
					ElevatorID: id,
				}
			}
			if amIImmobile(sup) && !sup.stuckDetected {
				sup.stuckDetected = true
				sup.recoveryFromImmobility = true
				sup.movingTimer.active = false
				sup.recoveryPrevFloor = sup.lastFloor
				supervisorEventChan <- SupervisorEvent{Type: SupervisorHardwareFault}
			}
			if obstructionTimedOut(sup) && !sup.stuckDetected {
				sup.stuckDetected = true
				sup.recoveryFromImmobility = false
				sup.obstructionTimer.active = false
				supervisorEventChan <- SupervisorEvent{Type: SupervisorHardwareFault}
			}
		}
	}
}

func initSupervisor() supervisor {
	return supervisor{
		elevatorTimers: elevatorTimers{
			{startTime: time.Now(), active: false},
			{startTime: time.Now(), active: false},
			{startTime: time.Now(), active: false},
		},
		movingTimer:            movingTimer{startTime: time.Now(), active: false},
		obstructionTimer:       obstructionTimer{startTime: time.Now(), active: false},
		stuckDetected:          false,
		recoveryFromImmobility: false,
		recoveryPrevFloor:      -1,
		lastFloor:              -1,
	}
}

// Returns the ID of the elevator that has lost connection, or -1 if no elevator has lost connection
func (elevatorTimers *elevatorTimers) lostConnectionToElevator() int {
	for elevator := 0; elevator < N_ELEVATORS; elevator++ {
		if elevatorTimers[elevator].active && time.Since(elevatorTimers[elevator].startTime) > 5*time.Second {
			elevatorTimers[elevator].active = false
			return elevator
		}
	}
	return -1
}

func amIImmobile(supervisor supervisor) bool {
	if supervisor.movingTimer.active && time.Since(supervisor.movingTimer.startTime) > 4*time.Second {
		return true
	}
	return false
}

func obstructionTimedOut(supervisor supervisor) bool {
	if supervisor.obstructionTimer.active && time.Since(supervisor.obstructionTimer.startTime) > 6*time.Second {
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
	if supervisor.recoveryFromImmobility {
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
