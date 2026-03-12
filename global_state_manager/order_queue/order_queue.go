package order_queue

import (
	. "distributed_elevator/elevio"
	. "distributed_elevator/global_state_manager/elevator_states"
	"fmt"
)

// !!! Should we drop GetOrder functions

// !!! Variable and function names can be improved

// !!! Check if aliveElevators map must be changed to connectedElevators in some functions

const (
	noElevatorAssigned  = 0
	hallButtonsPerFloor = 2
	numOfOrderStates    = 4 // !!! Needed?
)

type OrderState int

// !!! Better comments
// !!! Switch back to None = 0?
const (
	Completed   OrderState = iota // Order completed by at least one
	None                          // Order completed by all
	Unconfirmed                   // Order confirmed by at least one
	Confirmed                     // Order confirmed by all
)

type HallOrder struct {
	State      OrderState
	AssignedTo int
}

type AllHallOrders [N_FLOORS][hallButtonsPerFloor]HallOrder
type AllCabOrders [N_FLOORS][N_ELEVATORS]OrderState

type OrderQueue struct {
	Hall map[int]AllHallOrders // elevatorID -> that elevator's view of all hall orders
	Cab  map[int]AllCabOrders  // elevatorID -> that elevator's view of cab orders
}

// func GenerateEmptyOrderQueue() OrderQueue {
// 	return OrderQueue{
// 		Hall: make(map[int]AllHallOrders),
// 		Cab:  make(map[int]AllCabOrders),
// 	}
// }

func IsElevatorInQueue(queue *OrderQueue, viewerID int) bool {
	_, inHall := queue.Hall[viewerID]
	_, inCab := queue.Cab[viewerID]
	return inHall && inCab
}

// Function must be called with any elevator that is added to the system, before it can be assigned orders
// func AddElevatorToQueue(queue *OrderQueue, viewerID int) {
// 	if !IsElevatorInQueue(queue, viewerID) {
// 		queue.Hall[viewerID] = AllHallOrders{
// 			{{State: None, AssignedTo: noElevatorAssigned}, {State: None, AssignedTo: noElevatorAssigned}},
// 			{{State: None, AssignedTo: noElevatorAssigned}, {State: None, AssignedTo: noElevatorAssigned}},
// 			{{State: None, AssignedTo: noElevatorAssigned}, {State: None, AssignedTo: noElevatorAssigned}},
// 			{{State: None, AssignedTo: noElevatorAssigned}, {State: None, AssignedTo: noElevatorAssigned}},
// 		}
// 		queue.Cab[viewerID] = AllCabOrders{
// 			{None, None, None},
// 			{None, None, None},
// 			{None, None, None},
// 			{None, None, None},
// 		}
// 	}
// }

func GenerateNewOrderQueue() OrderQueue {
	queue := OrderQueue{
		Hall: make(map[int]AllHallOrders),
		Cab:  make(map[int]AllCabOrders),
	}

	for viewerID := 0; viewerID < N_ELEVATORS; viewerID++ {
		var hallOrders AllHallOrders
		var cabOrders AllCabOrders
		for floor := 0; floor < N_FLOORS; floor++ {
			for btn := 0; btn < hallButtonsPerFloor; btn++ {
				hallOrders[floor][btn] = HallOrder{
					State:      None,
					AssignedTo: noElevatorAssigned,
				}
			}
			for elevatorID := 0; elevatorID < N_ELEVATORS; elevatorID++ {
				cabOrders[floor][elevatorID] = None
			}
		}
		queue.Hall[viewerID] = hallOrders
		queue.Cab[viewerID] = cabOrders
	}

	return queue
}

func GetHallOrder(queue *OrderQueue, viewerID int, floor int, btn int) HallOrder {
	return queue.Hall[viewerID][floor][btn]
}

func GetCabOrder(queue *OrderQueue, viewerID int, floor int, assignedElevatorID int) OrderState {
	return queue.Cab[viewerID][floor][assignedElevatorID]
}

func (myQueue *OrderQueue) RetrieveMyOrders(myID int) [N_FLOORS][N_BUTTONS]bool {
	var orders [N_FLOORS][N_BUTTONS]bool
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			switch ButtonType(btn) {
			case BT_HallDown, BT_HallUp:
				hallOrder := GetHallOrder(myQueue, myID, floor, btn)
				if hallOrder.AssignedTo == myID && hallOrder.State == Confirmed {
					orders[floor][btn] = true
				}
			case BT_Cab:
				if GetCabOrder(myQueue, myID, floor, myID) == Confirmed {
					orders[floor][myID] = true
				}
			}
		}
	}
	return orders
}

// When a message is received with an order queue, this function should be called to update the local order queue with the new information.
// Only updates the orders for the elevator that sent the message, identified by ID.
func (myQueue *OrderQueue) UpdateOrderQueue(otherHallOrders AllHallOrders, otherCabOrders AllCabOrders, viewerID int) {
	if !IsElevatorInQueue(myQueue, viewerID) {
		fmt.Println("Attempted to update order queue with elevator not in queue: ", viewerID)
		return
	}
	myQueue.Hall[viewerID] = otherHallOrders
	myQueue.Cab[viewerID] = otherCabOrders
}

func IsOrderInProgress(
	queue *OrderQueue,
	elevatorStates ElevatorStates,
	btnEv ButtonEvent,
) bool {
	floor := btnEv.Floor
	btn := int(btnEv.Button)

	for _, elevatorPeer := range elevatorStates.Peers {
		if elevatorPeer.WorkingStatus == StatusLostConnection {
			continue
		}
		elevatorID := elevatorPeer.ID
		if btnEv.Button != BT_Cab {
			order := GetHallOrder(queue, elevatorID, floor, btn)
			if order.State == Unconfirmed || order.State == Confirmed {
				return true
			}
		} else {
			order := GetCabOrder(queue, elevatorID, floor, elevatorID)
			if order == Unconfirmed || order == Confirmed {
				return true
			}
		}
	}
	return false
}

// Function that appends new order, must call cost function to find assignTo parameter
func (queue *OrderQueue) AppendNewOrder(btnEv ButtonEvent, myID int, elevatorStates ElevatorStates, assignTo int) {
	floor := btnEv.Floor
	btn := int(btnEv.Button)

	if floor < 0 || floor >= N_FLOORS {
		fmt.Println("Attempted to append order at invalid floor: ", floor)
		return
	}
	if IsOrderInProgress(queue, elevatorStates, btnEv) {
		fmt.Println("Order is already in progress.")
		return
	}

	switch btnEv.Button {
	case BT_Cab:
		cabOrders := queue.Cab[myID]
		cabOrders[floor][assignTo] = Unconfirmed // AssignTo should always be myID for cab orders
		queue.Cab[myID] = cabOrders
	default:
		if assignTo < 0 || assignTo >= N_ELEVATORS {
			fmt.Println("Attempted to append invalid assignedTo: ", assignTo)
			return
		}

		hallOrders := queue.Hall[myID]

		hallOrders[floor][btn] = HallOrder{
			State:      Unconfirmed,
			AssignedTo: assignTo,
		}
		queue.Hall[myID] = hallOrders
	}
}

func (myQueue *OrderQueue) CompleteMyOrder(btnEvent ButtonEvent, elevatorStates ElevatorStates, myID int) bool {
	floor := btnEvent.Floor
	btn := int(btnEvent.Button)
	if floor < 0 || floor >= N_FLOORS {
		fmt.Println("Attempted to complete order at invalid floor: ", floor)
		return false
	}
	if elevatorStates.Peers[myID].WorkingStatus != StatusOK {
		fmt.Println("Attempted to complete an order for non-working elevator.")
		return false
	}
	switch btnEvent.Button {
	case BT_Cab:
		for _, elevatorPeer := range elevatorStates.Peers {
			if elevatorPeer.WorkingStatus == StatusLostConnection {
				continue
			}
			elevatorID := elevatorPeer.ID
			if GetCabOrder(myQueue, elevatorID, floor, elevatorID) != Confirmed {
				fmt.Println("Some elevator(s) not in Confirmed.") // Might need to also allow complete order
				return false
			}
		}
		cabOrders := myQueue.Cab[myID]
		cabOrders[floor][myID] = Completed
		myQueue.Cab[myID] = cabOrders
	default:
		for _, elevatorPeer := range elevatorStates.Peers {
			if elevatorPeer.WorkingStatus == StatusLostConnection {
				continue
			}
			elevatorID := elevatorPeer.ID
			if GetHallOrder(myQueue, elevatorID, floor, btn).State != Confirmed {
				fmt.Println("Some elevator(s) not in Confirmed.") // Might need to also allow complete order
				return false
			}
		}
		hallOrders := myQueue.Hall[myID]
		if hallOrders[floor][btn].AssignedTo != myID {
			fmt.Println("Attempted to mark order completed by wrong elevator. Order assigned to: ", hallOrders[floor][btn].AssignedTo, ", myID: ", myID)
			return false
		}
		hallOrders[floor][btn] = HallOrder{
			State:      Completed,
			AssignedTo: noElevatorAssigned,
		}
		myQueue.Hall[myID] = hallOrders
	}
	return true
}

// Assigner interface to gain access to AssignNewOrder behaviour, which is needed in RedistributeHallOrders
type Assigner interface {
	AssignNewOrder(ButtonEvent, ElevatorStates, AllCabOrders, int) int
}

func RedistributeHallOrders(myQueue *OrderQueue, myID int, elevatorStates ElevatorStates, assigner Assigner) {
	myHallOrders := myQueue.Hall[myID]
	status := make(map[int]bool)

	for _, elevatorPeer := range elevatorStates.Peers {
		if elevatorPeer.WorkingStatus == StatusOK {
			status[elevatorPeer.ID] = true
		} else {
			status[elevatorPeer.ID] = false
		}
	}
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			myHallOrder := myHallOrders[floor][btn]

			if status[myHallOrder.AssignedTo] { // If order's assigned elevator is working -> go to next order
				continue
			}
			buttonEvent := ButtonEvent{Floor: floor, Button: ButtonType(btn)}
			newID := assigner.AssignNewOrder(buttonEvent, elevatorStates, myQueue.Cab[myID], myID) // !!! Correct usage?
			myHallOrder.AssignedTo = newID
			myHallOrders[floor][btn] = myHallOrder
		}
	}
	myQueue.Hall[myID] = myHallOrders
}

func (myQueue *OrderQueue) TransitionSingleHallOrder(
	myID int,
	elevatorStates ElevatorStates,
	hallOrders *AllHallOrders,
	floor int,
	btn int,
) {
	myHallOrder := GetHallOrder(myQueue, myID, floor, btn)
	currentState := myHallOrder.State
	expectedAssignedTo := myHallOrder.AssignedTo
	otherHallOrder := myHallOrder // Initialized as my own

	switch currentState {
	case None:
		for _, elevatorPeer := range elevatorStates.Peers {
			elevatorID := elevatorPeer.ID
			if elevatorID == myID || elevatorPeer.WorkingStatus == StatusLostConnection {
				continue
			}
			otherHallOrder = GetHallOrder(myQueue, elevatorID, floor, btn)
			if otherHallOrder.State == Completed {
				return
			}
			if hallOrders[floor][btn].State < otherHallOrder.State {
				hallOrders[floor][btn].State = otherHallOrder.State
				hallOrders[floor][btn].AssignedTo = otherHallOrder.AssignedTo
			}
		}
		myQueue.Hall[myID] = *hallOrders
		return

	case Unconfirmed:
		for _, elevatorPeer := range elevatorStates.Peers {
			elevatorID := elevatorPeer.ID
			if elevatorID == myID || elevatorPeer.WorkingStatus == StatusLostConnection {
				continue
			}
			otherHallOrder = GetHallOrder(myQueue, elevatorID, floor, btn)
			if otherHallOrder.State == None || otherHallOrder.State == Completed {
				myQueue.Hall[myID] = *hallOrders // Ensuring we keep the lowest assignedTo ID even in transition failure
				return
			}
			shouldISwitchAssigned := (otherHallOrder.AssignedTo != expectedAssignedTo && otherHallOrder.AssignedTo > noElevatorAssigned && elevatorID < myID)
			if shouldISwitchAssigned {
				expectedAssignedTo = otherHallOrder.AssignedTo
			}
		}
		hallOrders[floor][btn].State = Confirmed
		hallOrders[floor][btn].AssignedTo = expectedAssignedTo
		myQueue.Hall[myID] = *hallOrders
		// SetButtonLamp(ButtonType(btn), floor, true)	// !!! Ensure this works correctly
		return

	case Confirmed:
		for _, elevatorPeer := range elevatorStates.Peers {
			elevatorID := elevatorPeer.ID
			if elevatorID == myID || elevatorPeer.WorkingStatus == StatusLostConnection {
				continue
			}
			otherHallOrder = GetHallOrder(myQueue, elevatorID, floor, btn)
			switch otherHallOrder.State {
			case None, Unconfirmed: // Double check
				return
			case Completed:
				hallOrders[floor][btn].State = Completed
				hallOrders[floor][btn].AssignedTo = noElevatorAssigned
				// SetButtonLamp(ButtonType(btn), floor, false)	// !!! Ensure this works correctly
			}
			shouldISwitchAssigned := (otherHallOrder.AssignedTo != expectedAssignedTo && otherHallOrder.AssignedTo > noElevatorAssigned && elevatorID < myID)
			if shouldISwitchAssigned {
				expectedAssignedTo = otherHallOrder.AssignedTo
			}
			// if otherHallOrder.State == None || otherHallOrder.State == Unconfirmed { // Must double check this
			// 	return
			// } else if otherHallOrder.State == Completed {
			// 	hallOrders[floor][btn].State = Completed
			// 	hallOrders[floor][btn].AssignedTo = noElevatorAssigned
			// }
		}
		myQueue.Hall[myID] = *hallOrders
		return

	case Completed:
		amIAlone := true
		for _, elevatorPeer := range elevatorStates.Peers {
			elevatorID := elevatorPeer.ID
			fmt.Println("Checking queue for ID: ", elevatorID)
			if elevatorID == myID || elevatorPeer.WorkingStatus == StatusLostConnection {
				continue
			}
			amIAlone = false
			otherHallOrder = GetHallOrder(myQueue, elevatorID, floor, btn)
			if otherHallOrder.State == Confirmed {
				fmt.Println("Other elevator was in Confirmed.")
				return
			}
			fmt.Println(otherHallOrder.State)
			if hallOrders[floor][btn].State < otherHallOrder.State {
				hallOrders[floor][btn].State = otherHallOrder.State
				hallOrders[floor][btn].AssignedTo = otherHallOrder.AssignedTo
			}
		}
		if amIAlone {
			hallOrders[floor][btn].State = None
			hallOrders[floor][btn].AssignedTo = noElevatorAssigned
		} else if hallOrders[floor][btn].State == Completed {
			hallOrders[floor][btn].State = None
		}
		myQueue.Hall[myID] = *hallOrders
		return
	default:
		fmt.Println("Undefined order state: ", currentState)
		return
	}
}

func (myQueue *OrderQueue) TransitionAllHallOrders(
	myID int,
	elevatorStates ElevatorStates,
) {
	hallOrders := myQueue.Hall[myID]
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < hallButtonsPerFloor; btn++ {
			myQueue.TransitionSingleHallOrder(myID, elevatorStates, &hallOrders, floor, btn)
		}
	}
}

func (myQueue *OrderQueue) TransitionSingleCabOrder(
	myID int,
	elevatorStates ElevatorStates,
	cabOrders *AllCabOrders,
	assignedElevatorID int,
	floor int,
) {
	myCabOrder := GetCabOrder(myQueue, myID, floor, assignedElevatorID)
	otherCabOrder := myCabOrder // Initialized as my own

	switch myCabOrder {
	case None:
		for _, elevatorPeer := range elevatorStates.Peers {
			elevatorID := elevatorPeer.ID
			if elevatorID == myID || elevatorPeer.WorkingStatus == StatusLostConnection {
				continue
			}
			otherCabOrder = GetCabOrder(myQueue, elevatorID, floor, assignedElevatorID)
			if otherCabOrder == Completed {
				return
			}
			if cabOrders[floor][assignedElevatorID] < otherCabOrder {
				cabOrders[floor][assignedElevatorID] = otherCabOrder
			}
		}
		myQueue.Cab[myID] = *cabOrders
		return

	case Unconfirmed:
		for _, elevatorPeer := range elevatorStates.Peers {
			elevatorID := elevatorPeer.ID
			if elevatorID == myID || elevatorPeer.WorkingStatus == StatusLostConnection {
				continue
			}
			otherCabOrder = GetCabOrder(myQueue, elevatorID, floor, assignedElevatorID)
			if otherCabOrder == None || otherCabOrder == Completed {
				return
			}
		}
		cabOrders[floor][assignedElevatorID] = Confirmed
		myQueue.Cab[myID] = *cabOrders
		return

	case Confirmed:
		for _, elevatorPeer := range elevatorStates.Peers {
			elevatorID := elevatorPeer.ID
			if elevatorID == myID || elevatorPeer.WorkingStatus == StatusLostConnection {
				continue
			}
			otherCabOrder = GetCabOrder(myQueue, elevatorID, floor, assignedElevatorID)
			switch otherCabOrder {
			case None, Unconfirmed: // Must check
				return
			case Completed:
				cabOrders[floor][assignedElevatorID] = Completed
			}
		}
		myQueue.Cab[myID] = *cabOrders
		return

	case Completed:
		amIAlone := true
		for _, elevatorPeer := range elevatorStates.Peers {
			elevatorID := elevatorPeer.ID
			if elevatorID == myID || elevatorPeer.WorkingStatus == StatusLostConnection {
				continue
			}
			amIAlone = false
			otherCabOrder = GetCabOrder(myQueue, elevatorID, floor, assignedElevatorID)
			if otherCabOrder == Confirmed {
				return
			}
			if cabOrders[floor][assignedElevatorID] < otherCabOrder {
				cabOrders[floor][assignedElevatorID] = otherCabOrder
			}
		}
		if amIAlone {
			cabOrders[floor][assignedElevatorID] = None
		} else if cabOrders[floor][assignedElevatorID] == Completed {
			cabOrders[floor][assignedElevatorID] = None
		}
		myQueue.Cab[myID] = *cabOrders
		return
	default:
		fmt.Println("Undefined order state: ", myCabOrder)
		return
	}
}

func (myQueue *OrderQueue) TransitionAllCabOrders(
	myID int,
	elevatorStates ElevatorStates,
) {
	cabOrders := myQueue.Cab[myID]

	for assignedElevatorID := 0; assignedElevatorID < N_ELEVATORS; assignedElevatorID++ {
		for floor := 0; floor < N_FLOORS; floor++ {
			myQueue.TransitionSingleCabOrder(myID, elevatorStates, &cabOrders, assignedElevatorID, floor)
		}
	}
}

func TestOrderQueue() {
	myId := 1
	yourId := 2
	hisId := 3

	viewOfQueue := GenerateEmptyOrderQueue()
	AddElevatorToQueue(&viewOfQueue, myId)
	AddElevatorToQueue(&viewOfQueue, yourId)
	AddElevatorToQueue(&viewOfQueue, hisId)

	elevatorStates := GenerateNewElevatorStates(myId)
	elevatorStates.Peers[myId-1].WorkingStatus = StatusOK
	elevatorStates.Peers[myId-1].ID = myId
	elevatorStates.Peers[myId-1].Floor = 0
	elevatorStates.Peers[yourId-1].WorkingStatus = StatusLostConnection
	elevatorStates.Peers[yourId-1].ID = yourId
	elevatorStates.Peers[yourId-1].Floor = 0
	elevatorStates.Peers[hisId-1].WorkingStatus = StatusLostConnection
	elevatorStates.Peers[hisId-1].ID = hisId
	elevatorStates.Peers[hisId-1].Floor = 0

	newButtonPress := ButtonEvent{
		Floor:  0,
		Button: ButtonType(0),
	}

	assignTo := myId

	// Test reconnection behaviour
	viewOfQueue.TransitionAllHallOrders(myId, elevatorStates)
	viewOfQueue.AppendNewOrder(newButtonPress, myId, elevatorStates, assignTo)

	for k, v := range viewOfQueue.Hall {
		fmt.Printf("%6v :  %+v\n", k, v)
	}

	viewOfQueue.TransitionAllHallOrders(myId, elevatorStates)

	for k, v := range viewOfQueue.Hall {
		fmt.Printf("%6v :  %+v\n", k, v)
	}

	elevatorStates.Peers[yourId-1].WorkingStatus = StatusOK
	elevatorStates.Peers[hisId-1].WorkingStatus = StatusOK
	elevatorStates.Peers[myId-1].WorkingStatus = StatusLostConnection

	viewOfQueue.TransitionAllHallOrders(yourId, elevatorStates)

	viewOfQueue.TransitionAllHallOrders(hisId, elevatorStates)

	elevatorStates.Peers[myId-1].WorkingStatus = StatusOK

	viewOfQueue.TransitionAllHallOrders(myId, elevatorStates)

	viewOfQueue.TransitionAllHallOrders(yourId, elevatorStates)

	viewOfQueue.TransitionAllHallOrders(hisId, elevatorStates)

	for k, v := range viewOfQueue.Hall {
		fmt.Printf("%6v :  %+v\n", k, v)
	}

	// Test normal behaviour
	/*
		viewOfQueue.TransitionAllHallOrders(myId, elevatorStates)

		viewOfQueue.TransitionAllHallOrders(yourId, elevatorStates)

		viewOfQueue.TransitionAllHallOrders(hisId, elevatorStates)

		for k, v := range viewOfQueue.Hall {
			fmt.Printf("%6v :  %+v\n", k, v)
		}

		viewOfQueue.AppendNewOrder(newButtonPress, myId, elevatorStates, assignTo)

		viewOfQueue.TransitionAllHallOrders(yourId, elevatorStates)

		viewOfQueue.TransitionAllHallOrders(hisId, elevatorStates)

		for k, v := range viewOfQueue.Hall {
			fmt.Printf("%6v :  %+v\n", k, v)
		}

		viewOfQueue.TransitionAllHallOrders(myId, elevatorStates)

		viewOfQueue.TransitionAllHallOrders(yourId, elevatorStates)

		viewOfQueue.TransitionAllHallOrders(hisId, elevatorStates)

		for k, v := range viewOfQueue.Hall {
			fmt.Printf("%6v :  %+v\n", k, v)
		}

		viewOfQueue.TransitionAllHallOrders(yourId, elevatorStates)

		viewOfQueue.TransitionAllHallOrders(myId, elevatorStates)

		viewOfQueue.TransitionAllHallOrders(hisId, elevatorStates)

		for k, v := range viewOfQueue.Hall {
			fmt.Printf("%6v :  %+v\n", k, v)
		}

		viewOfQueue.CompleteMyOrder(newButtonPress, elevatorStates, assignTo)

		viewOfQueue.TransitionAllHallOrders(yourId, elevatorStates)

		viewOfQueue.TransitionAllHallOrders(hisId, elevatorStates)

		for k, v := range viewOfQueue.Hall {
			fmt.Printf("%6v :  %+v\n", k, v)
		}
	*/
	newButtonPress = ButtonEvent{
		Floor:  2,
		Button: ButtonType(2),
	}

	fmt.Printf("output: \n")
	for k, v := range viewOfQueue.Cab {
		fmt.Printf("%6v :  %+v\n", k, v)
	}
}
