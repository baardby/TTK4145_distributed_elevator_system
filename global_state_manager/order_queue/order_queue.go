package order_queue

import (
	. "distributed_elevator/elevio"
	. "distributed_elevator/global_state_manager/elevator_states"
	"fmt"
)

// !!! Should we drop GetOrder functions? Use more often?

// !!! Variable and function names can be improved

const (
	NoElevatorAssigned  = -1
	HallButtonsPerFloor = 2
)

type OrderState int

// !!! Better comments
// !!! Switch back to None = 0?
const (
	Completed   OrderState = iota // Order completed by at least one elevator.
	None                          // Order completed by all elevators, or order never existed.
	Unconfirmed                   // Order confirmed by at least one elevator.
	Confirmed                     // Order confirmed by all connected elevators.
)

type HallOrder struct {
	State      OrderState
	AssignedTo int
}

// Matrix of all hall orders, from the perspective of a single elevator.
// Each order has a state and an assigned elevator ID.
type AllHallOrders [N_FLOORS][HallButtonsPerFloor]HallOrder

// Matrix of all cab orders, from the perspective of a single elevator.
// Each order has a state, and is assigned to the elevator corresponding to its column index.
type AllCabOrders [N_FLOORS][N_ELEVATORS]OrderState

type OrderQueue struct {
	Hall map[int]AllHallOrders // elevatorID -> that elevator's view of all hall orders.
	Cab  map[int]AllCabOrders  // elevatorID -> that elevator's view of all cab orders.
}

func GenerateNewOrderQueue() OrderQueue {
	queue := OrderQueue{
		Hall: make(map[int]AllHallOrders),
		Cab:  make(map[int]AllCabOrders),
	}
	// Initialize the order queue with empty orders for each elevator.
	for viewerID := 0; viewerID < N_ELEVATORS; viewerID++ {
		var hallOrders AllHallOrders
		var cabOrders AllCabOrders
		for floor := 0; floor < N_FLOORS; floor++ {
			for btn := 0; btn < HallButtonsPerFloor; btn++ {
				hallOrders[floor][btn] = HallOrder{
					State:      None,
					AssignedTo: NoElevatorAssigned,
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

func IsElevatorInQueue(queue *OrderQueue, viewerID int) bool {
	_, inHall := queue.Hall[viewerID]
	_, inCab := queue.Cab[viewerID]
	return inHall && inCab
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
			case BT_Cab:
				if GetCabOrder(myQueue, myID, floor, myID) == Confirmed {
					orders[floor][btn] = true
				}

			case BT_HallUp, BT_HallDown:
				hallOrder := GetHallOrder(myQueue, myID, floor, btn)
				if hallOrder.AssignedTo == myID && hallOrder.State == Confirmed {
					orders[floor][btn] = true
				}
			}
		}
	}
	return orders
}

// UpdateOrderQueue updates the local replica of the distributed order queue
// using information received from another elevator over the network.
//
// Each elevator broadcasts its view of assigned orders. When such a message
// is received, this function updates the local queue entries corresponding
// to the sending elevator (viewerID). // !!! Too much?
//
// Only the slice belonging to viewerID is updated.
func (queue *OrderQueue) UpdateOrderQueue(otherHallOrders AllHallOrders, otherCabOrders AllCabOrders, viewerID int) {
	if !IsElevatorInQueue(queue, viewerID) {
		fmt.Println("Attempted to update order queue with elevator not in queue: ", viewerID)
		return
	}
	queue.Hall[viewerID] = otherHallOrders
	queue.Cab[viewerID] = otherCabOrders
}

func IsOrderInProgress(queue *OrderQueue, elevatorStates ElevatorStates, btnEv ButtonEvent) bool {
	floor := btnEv.Floor
	btn := int(btnEv.Button)

	for _, elevatorPeer := range elevatorStates.Peers {
		if elevatorPeer.WorkingStatus == StatusLostConnection {
			continue
		}
		elevatorID := elevatorPeer.ID
		switch btnEv.Button {
		case BT_HallUp, BT_HallDown:
			order := GetHallOrder(queue, elevatorID, floor, btn)
			if order.State == Unconfirmed || order.State == Confirmed {
				return true
			}

		case BT_Cab:
			// Cab orders belong to a specific elevator, so we check the cab order for this elevator at this floor.
			order := GetCabOrder(queue, elevatorID, floor, elevatorID)
			if order == Unconfirmed || order == Confirmed {
				return true
			}
		}
	}
	return false
}

// AppendNewOrder initiates a new local order in the order queue.
// For hall orders, assignTo must be chosen by the caller beforehand.
// For cab orders, the order always belongs to myID.
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
		cabOrders[floor][myID] = Unconfirmed
		queue.Cab[myID] = cabOrders

	case BT_HallUp, BT_HallDown:
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
			if GetCabOrder(myQueue, elevatorID, floor, myID) != Confirmed {
				fmt.Println("Some elevator(s) not in Confirmed.") // !!! Might need to also allow complete order
				return false
			}
		}
		cabOrders := myQueue.Cab[myID]
		cabOrders[floor][myID] = Completed
		myQueue.Cab[myID] = cabOrders

	case BT_HallUp, BT_HallDown:
		for _, elevatorPeer := range elevatorStates.Peers {
			if elevatorPeer.WorkingStatus == StatusLostConnection {
				continue
			}
			elevatorID := elevatorPeer.ID
			// Checks every active elevator's view of my cab order at this floor.
			// If any of them is not in Confirmed -> I cannot mark this order as Completed.
			if GetHallOrder(myQueue, elevatorID, floor, btn).State != Confirmed {
				fmt.Println("Some elevator(s) not in Confirmed.") // !!! Might need to also allow complete order
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
			AssignedTo: NoElevatorAssigned,
		}
		myQueue.Hall[myID] = hallOrders
	}
	return true
}

// RedistributeHallOrders reassigns hall orders that were assigned to elevators that are no longer operational or connected to the network.
func (myQueue *OrderQueue) RedistributeHallOrders(
	myID int,
	elevatorStates ElevatorStates,
	assignNewOrder func(ButtonEvent, ElevatorStates, AllCabOrders, int) int) {

	myHallOrders := myQueue.Hall[myID]

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS-1; btn++ {
			order := myHallOrders[floor][btn]

			if order.AssignedTo == NoElevatorAssigned || elevatorStates.Peers[order.AssignedTo].WorkingStatus == StatusOK {
				continue
			}
			buttonEvent := ButtonEvent{
				Floor:  floor,
				Button: ButtonType(btn),
			}
			newID := assignNewOrder(buttonEvent, elevatorStates, myQueue.Cab[myID], myID)
			order.AssignedTo = newID
			myHallOrders[floor][btn] = order
		}
	}
	myQueue.Hall[myID] = myHallOrders
}

// TransitionSingleHallOrder updates this elevator's view of one hall order by
// comparing it with the corresponding order in other active elevators' views.
// The transition rules depend on the current local state and are intended to
// converge towards a shared state across elevators.
func (myQueue *OrderQueue) TransitionSingleHallOrder(
	myID int,
	elevatorStates ElevatorStates,
	hallOrders *AllHallOrders,
	floor int,
	btn int) {

	myHallOrder := GetHallOrder(myQueue, myID, floor, btn)
	currentState := myHallOrder.State
	currentAssignedTo := myHallOrder.AssignedTo
	var otherHallOrder HallOrder

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
			assignedToWorkingElevator := (otherHallOrder.AssignedTo > NoElevatorAssigned) && (elevatorStates.Peers[otherHallOrder.AssignedTo].WorkingStatus == StatusOK)
			shouldISwitchAssigned := assignedToWorkingElevator && (otherHallOrder.AssignedTo != currentAssignedTo) && (elevatorID < myID)

			if shouldISwitchAssigned {
				currentAssignedTo = otherHallOrder.AssignedTo
			}
		}
		hallOrders[floor][btn].State = Confirmed
		hallOrders[floor][btn].AssignedTo = currentAssignedTo
		myQueue.Hall[myID] = *hallOrders

	case Confirmed:
		canComplete := false
		for _, elevatorPeer := range elevatorStates.Peers {
			elevatorID := elevatorPeer.ID
			if elevatorID == myID || elevatorPeer.WorkingStatus == StatusLostConnection {
				continue
			}

			otherHallOrder = GetHallOrder(myQueue, elevatorID, floor, btn)
			switch otherHallOrder.State {
			case None, Unconfirmed:
				return
			case Completed:
				hallOrders[floor][btn].State = Completed
				currentAssignedTo = NoElevatorAssigned
				canComplete = true
				continue
			}
			assignedToWorkingElevator := (otherHallOrder.AssignedTo > NoElevatorAssigned) && (elevatorStates.Peers[otherHallOrder.AssignedTo].WorkingStatus == StatusOK)
			shouldISwitchAssigned := !canComplete && assignedToWorkingElevator && (otherHallOrder.AssignedTo != currentAssignedTo) && (elevatorID < myID)

			if shouldISwitchAssigned {
				currentAssignedTo = otherHallOrder.AssignedTo
			}
		}
		hallOrders[floor][btn].AssignedTo = currentAssignedTo
		myQueue.Hall[myID] = *hallOrders

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
			hallOrders[floor][btn].AssignedTo = NoElevatorAssigned
		} else if hallOrders[floor][btn].State == Completed {
			hallOrders[floor][btn].State = None
		}
		myQueue.Hall[myID] = *hallOrders
	default:
		fmt.Println("Undefined order state: ", currentState)
	}
}

func (myQueue *OrderQueue) TransitionAllHallOrders(myID int, elevatorStates ElevatorStates) {
	hallOrders := myQueue.Hall[myID]
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < HallButtonsPerFloor; btn++ {
			myQueue.TransitionSingleHallOrder(myID, elevatorStates, &hallOrders, floor, btn)
		}
	}
}

// TransitionSingleCabOrder updates this elevator's view of one cab order by
// comparing it against the same order in other active elevators' views.
// The transition depends on the local current state and is intended to help
// the system converge to a shared cab-order state.
func (myQueue *OrderQueue) TransitionSingleCabOrder(
	myID int,
	elevatorStates ElevatorStates,
	cabOrders *AllCabOrders,
	assignedElevatorID int,
	floor int,
) {

	myCabOrder := GetCabOrder(myQueue, myID, floor, assignedElevatorID)
	var otherCabOrder OrderState

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

	case Confirmed:
		for _, elevatorPeer := range elevatorStates.Peers {
			elevatorID := elevatorPeer.ID
			if elevatorID == myID || elevatorPeer.WorkingStatus == StatusLostConnection {
				continue
			}

			otherCabOrder = GetCabOrder(myQueue, elevatorID, floor, assignedElevatorID)
			switch otherCabOrder {
			case None, Unconfirmed:
				return
			case Completed:
				cabOrders[floor][assignedElevatorID] = Completed
			}
		}
		myQueue.Cab[myID] = *cabOrders

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
		if amIAlone || cabOrders[floor][assignedElevatorID] == Completed {
			cabOrders[floor][assignedElevatorID] = None
		}
		myQueue.Cab[myID] = *cabOrders
	default:
		fmt.Println("Undefined order state: ", myCabOrder)
	}
}

func (myQueue *OrderQueue) TransitionAllCabOrders(myID int, elevatorStates ElevatorStates) {
	cabOrders := myQueue.Cab[myID]

	for assignedElevatorID := 0; assignedElevatorID < N_ELEVATORS; assignedElevatorID++ {
		for floor := 0; floor < N_FLOORS; floor++ {
			myQueue.TransitionSingleCabOrder(myID, elevatorStates, &cabOrders, assignedElevatorID, floor)
		}
	}
}

// TESTING

func TestOrderQueue() {
	myId := 1
	yourId := 2
	hisId := 3

	viewOfQueue := GenerateNewOrderQueue()

	elevatorStates := GenerateNewElevatorStates(myId)
	elevatorStates.Peers[myId].WorkingStatus = StatusOK
	elevatorStates.Peers[myId].ID = myId
	elevatorStates.Peers[myId].Floor = 0
	elevatorStates.Peers[yourId].WorkingStatus = StatusLostConnection
	elevatorStates.Peers[yourId].ID = yourId
	elevatorStates.Peers[yourId].Floor = 0
	elevatorStates.Peers[hisId].WorkingStatus = StatusLostConnection
	elevatorStates.Peers[hisId].ID = hisId
	elevatorStates.Peers[hisId].Floor = 0

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

	elevatorStates.Peers[yourId].WorkingStatus = StatusOK
	elevatorStates.Peers[hisId].WorkingStatus = StatusOK
	elevatorStates.Peers[myId].WorkingStatus = StatusLostConnection

	viewOfQueue.TransitionAllHallOrders(yourId, elevatorStates)

	viewOfQueue.TransitionAllHallOrders(hisId, elevatorStates)

	elevatorStates.Peers[myId].WorkingStatus = StatusOK

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
