package elevator_states //Må endres hvis det puttes inn i en mappe

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
)

type ElevatorPeer struct {
	Floor     int
	Direction MotorDirection
	Behaviour ElevatorBehaviour
	Alive     bool
	ID        int
}

type ElevatorStates struct {
	Peers [N_ELEVATORS]ElevatorPeer
}

func (elevatorStates *ElevatorStates) UpdateElevatorState(elevatorPeer ElevatorPeer) {
	elevatorStates.Peers[elevatorPeer.ID] = elevatorPeer
}

func GenerateNewElevatorStates() ElevatorStates {
	var elevatorStates ElevatorStates
	for i := 0; i < N_ELEVATORS; i++ {
		elevatorStates.Peers[i] = ElevatorPeer{
			Floor:     -1,
			Direction: MD_Stop,
			Behaviour: EB_Idle,
			Alive:     false,
			ID:        i,
		}
	}
	return elevatorStates
}

func UpdateAliveElevatorsMap(elevatorStates ElevatorStates, aliveElevatorsMap map[int]bool) {
	for i := 0; i < N_ELEVATORS; i++ {
		aliveElevatorsMap[elevatorStates.Peers[i].ID] = elevatorStates.Peers[i].Alive
	}
}

func ThisElevatorToElevatorPeer(elevator Elevator, MyId int) ElevatorPeer {
	return ElevatorPeer{
		Floor:     elevator.Floor,
		Direction: elevator.Direction,
		Behaviour: elevator.Behaviour,
		Alive:     true,
		ID:        MyId,
	}
}
