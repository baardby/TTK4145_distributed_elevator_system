 package global_state_manager // å endres hvis det puttes inn i en mappe

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/supervisor"
	. "distributed_elevator/elevio"
	"fmt"
)

const noElevatorAssigned = -1
const allFlagsSet uint8 = 0b00000111	// Assumes N_ELEVATORS = 3

type OrderState int

const (
	None        OrderState = iota   // Order completed by all
	Unconfirmed                     // Order confirmed by at least one
	Confirmed                       // Order confirmed by all
	Completed                       // Order completed by at least one
)

type HallOrder struct {
	AssignedTo	int		// Corresponding to elevator ID, -1 if unassigned
	Active		bool	// Order exists (state is not None)
	SeenBy		uint8	// Bit i represents whether elevator i has seen order
	CompletedBy uint8	// Bit i represents whether elevator i has completed order
	ClearedBy 	uint8	// Bit i represents whether elevator i is ready to clear order
}

type CabOrder struct {
	AssignedTo	int		// Will always be self elevator
	Active		bool	// Order exists (state is not None)
	SeenBy		uint8	// Bit i represents whether elevator i has seen order
	CompletedBy uint8	// Bit i represents whether elevator i has completed order
	ClearedBy 	uint8	// Bit i represents whether elevator i has cleared order
}

const buttonsPerFloor = 2

type OrderQueue struct {
	HallOrderList	[N_FLOORS][buttonsPerFloor]HallOrder
}

func GenerateNewOrderQueue() OrderQueue {
	newQueue := OrderQueue{}
	
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < buttonsPerFloor; btn++ {
			newQueue.HallOrderList[floor][btn] = HallOrder {
				AssignedTo: noElevatorAssigned,
				Active: false,
				SeenBy: 0,
				CompletedBy: 0,
				ClearedBy: 0,
			}
		}
	}

	return newQueue
}

func (order HallOrder) DerivedOrderState() OrderState {
	if !order.Active {
		return None
	}

	// Only the order's assignedTo elevator should set its bit in CompletedBy
	if order.CompletedBy != 0 {
		return Completed
	}

	if order.SeenBy == allFlagsSet {
		return Confirmed
	}

	return Unconfirmed
}

// Not needed, already part of other functions
// func (order *HallOrder) MarkSeenByMe(myID int) {
// 	if !order.Active {
// 		return
// 	}
// 	order.SeenBy |= 1 << myID
// }

// New idea with matrices for state
// func MergeHallOrder(myHallOrder HallOrder, peerHallOrder1 HallOrder, peerHallOrder2 HallOrder, floor int, button int) HallOrder {
// 	out.Active = myHallOrder.Active || peerHallOrder1.Active || peerHallOrder2.Active	// If any order active, keep this information

// 	// If order not active, return a clean inactive state
// 	if !out.Active {	
// 		out.AssignedTo = noElevatorAssigned
// 		out.SeenBy = 0
// 		out.CompletedBy = 0
// 		out.ClearedBy = 0
// 		return out
// 	}

// 	currentState = myHallOrder.HallOrderList[floor][button]
// 	switch currentState{
// 		case None:
// 			if anyInNextState() && noneInPrevState() {
// 				inputQueueinputQueue.HallOrderList[floor][button] = Unconfirmed
// 				return inputQueue
// 			}
// 		case Unconfirmed:
// 			if allAreUnconfirmed() || allAreUnconfirmedOrConfirmed() {
// 				inputQueueinputQueue.HallOrderList[floor][button] = Unconfirmed
// 				return inputQueue
// 			}	
// 	}

// 	switch {
// 	case h1.AssignedTo != noElevatorAssigned && h2.AssignedTo == noElevatorAssigned:
// 		out.AssignedTo = h1.AssignedTo
// 	case h1.AssignedTo == noElevatorAssigned:
// 		out.AssignedTo = h2.AssignedTo
// 	case h1.AssignedTo == h2.AssignedTo:
// 		out.AssignedTo = h1.AssignedTo
// 	default:
// 		// To prevent conflicting messages, choose lowest ID
// 		if h1.AssignedTo < h2.AssignedTo {
// 			out.AssignedTo = h1.AssignedTo
// 		} else {
// 			out.AssignedTo = h2.AssignedTo
// 		}
// 	}
// 	return out
// }

// !!! Better variable names
// Function that merges HallOrder h1 and h2, call function with h1 as own order in normal use case 
func MergeHallOrder(h1 HallOrder, h2 HallOrder, myID int) HallOrder {
	out := h1
	out.Active |= h2.Active	// If any order active, keep this information

	// If order not active, return a clean inactive state
	if !out.Active {	
		out.AssignedTo = noElevatorAssigned
		out.SeenBy = 0
		out.CompletedBy = 0
		out.ClearedBy = 0
		return out
	}

	out.SeenBy = h1.SeenBy | h2.SeenBy
	out.CompletedBy = h1.CompletedBy | h2.CompletedBy
	out.ClearedBy = h1.ClearedBy | h2.ClearedBy

	switch {
	case h1.AssignedTo != noElevatorAssigned && h2.AssignedTo == noElevatorAssigned:
		out.AssignedTo = h1.AssignedTo
	case h1.AssignedTo == noElevatorAssigned:
		out.AssignedTo = h2.AssignedTo
	case h1.AssignedTo == h2.AssignedTo:
		out.AssignedTo = h1.AssignedTo
	default:
		// To prevent conflicting messages, choose lowest ID
		if h1.AssignedTo < h2.AssignedTo {
			out.AssignedTo = h1.AssignedTo
		} else {
			out.AssignedTo = h2.AssignedTo
		}
	}
	out.SeenBy |= 1 << myID
	return out
}

// Function that merges queues q1 and q2
func MergeQueue(q1 OrderQueue, q2 OrderQueue, myId int) OrderQueue {
	out := q1
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < buttonsPerFloor; btn++ {
			h1 := q1.HallOrderList[floor][btn]
			h2 := q2.HallOrderList[floor][btn]
			out.HallOrderList[floor][btn] = MergeHallOrder(h1, h2, myID)
		}
	}
	return out
}

// Function that appends new order, must call cost function to find assignTo parameter
func (q *OrderQueue) AppendNewOrder(btnEv ButtonEvent, assignTo int, myID int) {
	f := btnEv.Floor
	btn := int(btnEv.Button)

	if f < 0 || f >= N_FLOORS {
		fmt.Println("Attempted to append order at invalid floor: ", f)
		return
	}
	if btnEv.Button != BT_HallUp && btnEv.Button != BT_HallDown {
		fmt.Println("Not hall button.")
		return
	}
	if assignTo < 0 || assignTo >= N_ELEVATORS {
		fmt.Println("Attempted to append invalid assignedTo: ", assignTo)
		return
	}

	order := &q.HallOrderList[f][btn]
	order.Active = true
	if order.AssignedTo == noElevatorAssigned {
		order.AssignedTo = assignTo
	} else if order.AssignedTo != assignTo {
		if assignTo < order.AssignedTo {
			order.AssignedTo = assignTo // In case of conflict (simultaneous button press),
										// let lowest ID elevator keep order
		}
	}
	order.SeenBy |= 1 << myID
}

// !!! Must fix event parameter, maybe take in button on that floor when completing?
// Function that marks an order as completed by myself if myID corresponds with the order's assignedTo
func (q *OrderQueue) MarkOrderCompletedByMe(event Event?, myID int) {
	f := event.Floor
	btn := int(event.Button)

	if f < 0 || f >= N_FLOORS {
		fmt.Println("Attempted to clear order at invalid floor: ", f)
		return
	}
	if event.Button != BT_HallUp && event.Button != BT_HallDown {
		fmt.Println("Not hall button.")
		return
	}

	order := &q.HallOrderList[f][btn]

	if !order.Active {
		return
	}
	if myID != order.AssignedTo { // Preventing elevators from marking complete if not assigned
		return
	}
	
	order.CompletedBy |= 1 << myID
	order.SeenBy |= 1 << myID
}
// !!! Must fix event parameter, maybe take in button on that floor when clearing?
// Function that marks an order as cleared by myself if order's assignedTo has already marked complete
func (q *OrderQueue) MarkOrderClearedByMe(event Event?, myID int) {
	f := event.Floor
	btn := int(event.Button)

	if f < 0 || f >= N_FLOORS {
		fmt.Println("Attempted to clear order at invalid floor: ", f)
		return
	}
	if event.Button != BT_HallUp && event.Button != BT_HallDown {
		fmt.Println("Not hall button.")
		return
	}

	order := &q.HallOrderList[f][btn]

	if !order.Active {
		return
	}
	if order.AssignedTo == noElevatorAssigned {
		return
	}
	// Check if order's assignedTo elevator has completed it
	if order.CompletedBy == (1 << order.AssignedTo) {
		order.ClearedBy |= 1 << myID
	}
}

// !!! Again, need to fix event parameter
// Function that attempts to reset order if cleared by all
func (q *OrderQueue) resetOrder(event Event?) {
	f := event.Floor
	btn := int(event.Button)

	if f < 0 || f >= N_FLOORS {
		fmt.Println("Attempted to clear order at invalid floor: ", f)
		return
	}
	if event.Button != BT_HallUp && event.Button != BT_HallDown {
		fmt.Println("Not hall button.")
		return
	}

	order := &q.HallOrderList[f][btn]

	if !order.Active {
		return
	}
	if order.AssignedTo == noElevatorAssigned {
		return
	}
	if order.CompletedBy != (1 << order.AssignedTo) {
		return
	}

	// Check if all elevators are ready to clear
	if order.ClearedBy == allFlagsSet {
		order.Active =  false
		order.AssignedTo = noElevatorAssigned
		order.SeenBy = 0
		order.CompletedBy = 0
		order.ClearedBy=  0
	}
}

// !!! NOTES !!!

// If using flags instead of matrices, 
// make a general function that does the shifting to improve readability

// Ensure that btnEvent: {floor = 0, btnType = BT_HALL_DOWN} and {floor = N_FLOORS, btnType = BT_HALL_UP}
// is not possible

// Implement for Cab orders as well

// When making global_state_manager loop in other go file, talk to Magnus.
// Idea: First step is to check if queue exists, if not wait 2 seconds and the generate
// In 2 second interval, if some other queue is sent to listener, make this the queue

/* ---------------------------------------------------------------------------------*/

/* START COMMENT

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

END COMMENT */


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