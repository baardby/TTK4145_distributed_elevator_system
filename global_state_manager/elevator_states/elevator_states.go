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

func (elevatorStates *ElevatorStates) UpdatePeer(elevatorPeer ElevatorPeer, myId int) {
	if elevatorPeer.ID == myId {
		elevatorPeer.WorkingStatus = elevatorStates.Peers[myId-1].WorkingStatus // Keep the old working status for myself, it should only be updated by the supervisor
	}
	elevatorStates.Peers[elevatorPeer.ID-1] = elevatorPeer
}

func GenerateNewElevatorStates(myId int) ElevatorStates {
	var elevatorStates ElevatorStates
	for i := 0; i < N_ELEVATORS; i++ {
		elevatorStates.Peers[i] = ElevatorPeer{
			Floor:         -1,
			Direction:     MD_Stop,
			Behaviour:     EB_Idle,
			WorkingStatus: StatusLostConnection,
			ID:            i + 1,
		}
	}
	elevatorStates.Peers[myId-1].WorkingStatus = StatusOK
	return elevatorStates
}

func ThisElevatorToElevatorPeer(elevator Elevator, myId int) ElevatorPeer {
	return ElevatorPeer{
		Floor:     elevator.Floor,
		Direction: elevator.Direction,
		Behaviour: elevator.Behaviour,
		ID:        myId,
	}
}
