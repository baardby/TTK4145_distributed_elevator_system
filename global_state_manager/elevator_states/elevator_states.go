package elevator_states

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
)

type ElevatorStatus int

const (
	StatusOK ElevatorStatus = iota
	StatusHardwareFault
	StatusLostConnection
)

type ElevatorPeer struct {
	Floor         int
	Direction     MotorDirection
	Behaviour     ElevatorBehaviour
	WorkingStatus ElevatorStatus
	ID            int
}

type ElevatorStates struct {
	Peers [N_ELEVATORS]ElevatorPeer
}

func (elevatorStates *ElevatorStates) UpdatePeer(elevatorPeer ElevatorPeer, myID int) {
	if elevatorPeer.ID == myID {
		// Keep the old working status for myself, it should only be updated by the supervisor
		elevatorPeer.WorkingStatus = elevatorStates.Peers[myID].WorkingStatus
	}
	elevatorStates.Peers[elevatorPeer.ID] = elevatorPeer
}

func GenerateNewElevatorStates(myID int) ElevatorStates {
	var elevatorStates ElevatorStates
	for i := 0; i < N_ELEVATORS; i++ {
		elevatorStates.Peers[i] = ElevatorPeer{
			Floor:         -1,
			Direction:     MD_Stop,
			Behaviour:     EB_Idle,
			WorkingStatus: StatusLostConnection,
			ID:            i,
		}
	}
	elevatorStates.Peers[myID].WorkingStatus = StatusOK
	return elevatorStates
}

func ThisElevatorToElevatorPeer(elevator Elevator, myID int) ElevatorPeer {
	return ElevatorPeer{
		Floor:     elevator.Floor,
		Direction: elevator.Direction,
		Behaviour: elevator.Behaviour,
		ID:        myID,
	}
}
