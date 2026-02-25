package supervisor //Må endres hvis det puttes inn i en mappe

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
)

const N_ELEVATORS = 3

type ElevatorPeer struct {
	Floor     int
	Direction MotorDirection
	Behaviour ElevatorBehaviour
	Alive     bool
	//Bør ha en elevtor ID ellerno sånt
}

type ElevatorStates struct {
	Peers [N_ELEVATORS]ElevatorPeer
}

func (elevatorStates *ElevatorStates) UpdateElevatorStates(elevatorNum int, floor int, direction MotorDirection, behaviour ElevatorBehaviour) {
	elevatorStates.Peers[elevatorNum] = ElevatorPeer{
		Floor:     floor,
		Direction: direction,
		Behaviour: behaviour,
		Alive:     true,
	}
}
