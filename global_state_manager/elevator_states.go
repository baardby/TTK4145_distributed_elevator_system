package global_state_manager //Må endres hvis det puttes inn i en mappe

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

func (elevatorStates *ElevatorStates) updateElevatorState(elevatorPeer ElevatorPeer) {
	elevatorStates.Peers[elevatorPeer.ID] = elevatorPeer
}

func generateNewElevatorStates() ElevatorStates {
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

func updateAliveElevatorsMap(elevatorStates ElevatorStates, aliveElevatorsMap map[int]bool) {
	for i := 0; i < N_ELEVATORS; i++ {
		aliveElevatorsMap[elevatorStates.Peers[i].ID] = elevatorStates.Peers[i].Alive
	}
}
