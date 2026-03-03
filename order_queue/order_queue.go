package order_queue // å endres hvis det puttes inn i en mappe

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/supervisor"
	. "distributed_elevator/elevio"
)

type OrderState int
const buttonsPerFloor = 2
const noElevatorAssigned = -1

const (
	None        OrderState = iota   // Order completed by all
	Unconfirmed                     // Order confirmed by at least one
	Confirmed                       // Order confirmed by all
	Completed                       // Order completed by at least one
)

// OrderStates stores the state of each order.
//
// OrderAssignedTo stores the assigned elevator ID for each order.
// The ID mapping is:
//   -1 : no elevator assigned
//    0 : elevator 1
//    1 : elevator 2
//    2 : elevator 3
type OrderQueue struct {
	HallOrderStates      [N_FLOORS][buttonsPerFloor]OrderState
	CabOrderStates       [N_FLOORS][N_ELEVATORS]OrderState
	HallOrdersAssignedTo [N_FLOORS][buttonsPerFloor]int
	CabOrdersAssignedTo  [N_FLOORS][N_ELEVATORS]int
}

func initOrderQueue() (rq OrderQueue) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < buttonsPerFloor; btn++ {
			rq.HallOrdersAssignedTo[floor][btn] = noElevatorAssigned
		}
		for elevator := 0; elevator < N_ELEVATORS; elevator++ {
			rq.CabOrdersAssignedTo[floor][elevator] = noElevatorAssigned
		}
	}
	return
}

//func restoreQueueAfterDisconnect 		//For å fikse når man kommer tilbake på nettet
//		input: None
//		Listen for other queue
//		Figure out if two different queues are recieved.
//		if riecieved:
//			Adopt
//		else:
//			init_queue()
//		return: None

func QueueUnion
		input: QueueFromElevator
		figure out logic with Union
		Barriers?
		Update Orders_queue
		return: none

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

func (currentOrderQueue *OrderQueue) AppendNewOrder(newButtonEvent ButtonEvent, assignedElevator ElevatorPeer) {
	floor := newButtonEvent.Floor
	button := newButtonEvent.Button
	ID := assignedElevator.ID

	if floor < 0 || floor >= N_FLOORS {
		fmt.Println("Invalid floor")
		return
	}

	// Sanity for button also?

	if button == BT_Cab {
		if currentcurrentOrderQueue.CabOrderStates[floor][ID] != Confirmed {
			currentcurrentOrderQueue.CabOrderStates[newButtonEvent.Floor][ID] = Unconfirmed
			currentcurrentOrderQueue.CabOrdersAssignedTo[newButtonEvent.Floor][ID] = ID
		}
	} else {
		if currentOrderQueue.HallOrderStates[newButtonEvent.Floor][newButtonEvent.Button] != Confirmed {
			currentcurrentOrderQueue.HallOrderStates[newButtonEvent.Floor][ID] = Unconfirmed
			currentcurrentOrderQueue.HallOrdersAssignedTo[newButtonEvent.Floor][ID] = ID
		}
	}
}

select {
	case newOrder := <-channelFromSupervisor:
		do stuff
	case sendQueueToSender<- bigQueue:
}


// Elevator controller:

// Sjekk elevator_controller for info om struktur av channels
// Skal sende cab og hall som nye bool matriser på updateQueueCh
// Skal bare sende cab og hall orders som er assignedTo denne heisen

// Skal også motta nye knappetrykk fra elevator controller

// Sender and listener:

// Select example in network sender
// Use ticker to periodically send to to network sender

// Standard case with channel to read from listener

// Main for loop needs to listen for timeout from supervisor on minimum 2 channels (other 2 elevators)