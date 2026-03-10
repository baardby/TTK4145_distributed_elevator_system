package order_queue

import (
	. "distributed_elevator/elevio"
	"fmt"
)

// !!! Should we drop GetOrder functions

// !!! Variable and function names can be improved

// !!! Check if aliveElevators map must be changed to connectedElevators in some functions

const (
	noElevatorAssigned  = 0
	hallButtonsPerFloor = 2
	numOfOrderStates    = 4
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

type Order struct {
	State      OrderState
	AssignedTo int
}

type AllHallOrders [N_FLOORS][hallButtonsPerFloor]Order
type AllCabOrders [N_FLOORS][N_ELEVATORS]OrderState

type OrderQueue struct {
	Hall map[int]AllHallOrders // elevatorID -> that elevator's view of all hall orders
	Cab  map[int]AllCabOrders  // elevatorID -> that elevator's view of cab orders
}

func GenerateEmptyOrderQueue() OrderQueue {
	return OrderQueue{
		Hall: make(map[int]AllHallOrders),
		Cab:  make(map[int]AllCabOrders),
	}
}

func IsElevatorInQueue(queue *OrderQueue, ID int) bool {
	_, inHall := queue.Hall[ID]
	_, inCab := queue.Cab[ID]
	return inHall && inCab
}

// Function must be called with any elevator that is added to the system, before it can be assigned orders
func AddElevatorToQueue(queue *OrderQueue, ID int) {
	if !IsElevatorInQueue(queue, ID) {
		queue.Hall[ID] = AllHallOrders{}
		queue.Cab[ID] = AllCabOrders{}
	}
}

func GetHallOrder(queue *OrderQueue, ID int, floor int, btn int) Order {
	return queue.Hall[ID][floor][btn]
}

func GetCabOrder(queue *OrderQueue, ID int, floor int) Order {
	return queue.Cab[ID][floor]
}

func (myQueue *OrderQueue) RetrieveMyOrders(myID int) [N_FLOORS][N_BUTTONS]bool {
	var orders [N_FLOORS][N_BUTTONS]bool
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if GetHallOrder(myQueue, myID, floor, btn).AssignedTo == myID || GetCabOrder(myQueue, myID, floor).AssignedTo == myID {
				orders[floor][btn] = true
			}
		}
	}
	return orders
}

// When a message is received with an order queue, this function should be called to update the local order queue with the new information. Only updates the orders for the elevator that sent the message, identified by ID.
func (myQueue *OrderQueue) UpdateOrderQueue(otherQueue OrderQueue, ID int) {
	if !IsElevatorInQueue(myQueue, ID) {
		fmt.Println("Attempted to update order queue with elevator not in queue: ", ID)
		return
	}
	myQueue.Hall[ID] = otherQueue.Hall[ID]
	myQueue.Cab[ID] = otherQueue.Cab[ID]
}

func IsOrderInProgress(
	queue *OrderQueue,
	aliveElevators map[int]bool,
	btnEv ButtonEvent,
) bool {
	floor := btnEv.Floor
	btn := int(btnEv.Button)

	for ID, alive := range aliveElevators {
		if !alive {
			continue
		}
		if btnEv.Button != BT_Cab {
			order := GetHallOrder(queue, ID, floor, btn)
			if order.State == Unconfirmed || order.State == Confirmed {
				return true
			}
		} else {
			order := GetCabOrder(queue, ID, floor)
			if order.State == Unconfirmed || order.State == Confirmed {
				return true
			}
		}
	}
	return false
}

// Function that appends new order, must call cost function to find assignTo parameter
func (queue *OrderQueue) AppendNewOrder(btnEv ButtonEvent, myID int, aliveElevators map[int]bool, assignTo int) {
	floor := btnEv.Floor
	btn := int(btnEv.Button)

	if floor < 0 || floor >= N_FLOORS {
		fmt.Println("Attempted to append order at invalid floor: ", floor)
		return
	}
	if IsOrderInProgress(queue, aliveElevators, btnEv) {
		fmt.Println("Order is already in progress.")
		return
	}

	if btnEv.Button != BT_Cab { // !!! Use switch case?

		if assignTo < 0 || assignTo >= N_ELEVATORS {
			fmt.Println("Attempted to append invalid assignedTo: ", assignTo)
			return
		}

		hallOrders := queue.Hall[myID]

		hallOrders[floor][btn] = Order{
			State:      Unconfirmed,
			AssignedTo: assignTo,
		}
		queue.Hall[myID] = hallOrders
	} else {
		cabOrders := queue.Cab[myID]
		cabOrders[floor] = Order{
			State:      Unconfirmed,
			AssignedTo: myID,
		}
		queue.Cab[myID] = cabOrders
	}
}

func (myQueue *OrderQueue) CompleteMyOrder(btnEvent ButtonEvent, aliveElevators map[int]bool, myID int) bool {
	floor := btnEvent.Floor
	btn := int(btnEvent.Button)

	if floor < 0 || floor >= N_FLOORS {
		fmt.Println("Attempted to append order at invalid floor: ", floor)
		return false
	}
	if !IsOrderInProgress(myQueue, aliveElevators, btnEvent) {
		fmt.Println("This order is not in progress.")
		return false
	}

	for ID, alive := range aliveElevators {
		if !alive || ID == myID {
			continue
		}
		if GetHallOrder(myQueue, ID, floor, btn).State != Confirmed {
			fmt.Print("Some elevator(s) no in Confirmed.")
			return false
		}
	}

	if btnEvent.Button != BT_Cab { // !!! Use switch case?
		hallOrders := myQueue.Hall[myID]

		if hallOrders[floor][btn].AssignedTo != myID {
			fmt.Println("Attempted to mark order completed by wrong elevator. Order assigned to: ", hallOrders[floor][btn].AssignedTo, ", myID: ", myID)
			return false
		}

		hallOrders[floor][btn] = Order{
			State:      Completed,
			AssignedTo: noElevatorAssigned,
		}
		myQueue.Hall[myID] = hallOrders
	} else {
		cabOrders := myQueue.Cab[myID]

		if cabOrders[floor].AssignedTo != myID {
			fmt.Println("Attempted to mark order completed by wrong elevator. Order assigned to: ", cabOrders[floor].AssignedTo, ", myID: ", myID)
			return false
		}

		cabOrders[floor] = Order{
			State:      Completed,
			AssignedTo: noElevatorAssigned,
		}
		myQueue.Cab[myID] = cabOrders
	}
	return true
}

/* START
func CanHallOrderTransitionState(
	queue *OrderQueue,
	myID int,
	aliveElevators map[int]bool,
	floor int,
	btn int,
) bool {
	currentState := GetHallOrder(queue, myID, floor, btn).State

	switch currentState {
	case None:
		for ID, alive := range aliveElevators {
			if !alive || ID == myID {
				continue
			}
			if GetHallOrder(queue, ID, floor, btn).State == Unconfirmed {
				return true
			}
		}
		return false

	case Unconfirmed:
		for ID, alive := range aliveElevators {
			if !alive || ID == myID {
				continue
			}
			if GetHallOrder(queue, ID, floor, btn).State == None || GetHallOrder(queue, ID, floor, btn).State == Completed {
				return false
			}
		}
		return true

	case Confirmed:
		for ID, alive := range aliveElevators {
			if !alive || ID == myID {
				continue
			}
			if GetHallOrder(queue, ID, floor, btn).State == Completed { // Must double check this
				return true
			}
		}
		return false

	case Completed:
		for ID, alive := range aliveElevators {
			if !alive || ID == myID {
				continue
			}
			if GetHallOrder(queue, ID, floor, btn).State == Confirmed {
				return false
			}
		}
		return true
	default:
		fmt.Println("Undefined order state: ", currentState)
		return false
	}
}

func CanCabOrderTransitionState(
	queue *OrderQueue,
	myID int,
	aliveElevators map[int]bool,
	floor int,
) bool {
	currentState := GetCabOrder(queue, myID, floor).State

	switch currentState {
	case None:
		for ID, alive := range aliveElevators {
			if !alive || ID == myID {
				continue
			}
			if GetCabOrder(queue, ID, floor).State == Unconfirmed {
				return true
			}
		}
		return false

	case Unconfirmed:
		for ID, alive := range aliveElevators {
			if !alive || ID == myID {
				continue
			}
			if GetCabOrder(queue, ID, floor).State == None || GetCabOrder(queue, ID, floor).State == Completed {
				return false
			}
		}
		return true

	case Confirmed:
		for ID, alive := range aliveElevators {
			if !alive || ID == myID {
				continue
			}
			if GetCabOrder(queue, ID, floor).State == Completed { // Must double check this
				return true
			}
		}
		return false

	case Completed:
		for ID, alive := range aliveElevators {
			if !alive || ID == myID {
				continue
			}
			if GetCabOrder(queue, ID, floor).State == Confirmed {
				return false
			}
		}
		return true
	default:
		fmt.Println("Undefined order state: ", currentState)
		return false
	}
}

func (myQueue *OrderQueue) TransitionMyQueue(myID int, aliveElevators map[int]bool, otherID int) {
	hallOrders := myQueue.Hall[myID]
	cabOrders := myQueue.Cab[myID]

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < hallButtonsPerFloor; btn++ {
			if CanHallOrderTransitionState(myQueue, myID, aliveElevators, floor, btn) {
				currentState := GetHallOrder(myQueue, myID, floor, btn).State
				nextState := (currentState + 1) % numOfOrderStates
				hallOrders[floor][btn].State = nextState

			}
		}
		if CanCabOrderTransitionState(myQueue, myID, aliveElevators, floor) {
			currentState := GetCabOrder(myQueue, myID, floor).State
			nextState := (currentState + 1) % numOfOrderStates
			cabOrders[floor].State = nextState
		}
	}
	myQueue.Hall[myID] = hallOrders
	myQueue.Cab[myID] = cabOrders
} END */

func (myQueue *OrderQueue) TransitionHallOrders(
	myID int,
	aliveElevators map[int]bool,
	floor int,
	btn int,
) bool {
	hallOrders := myQueue.Hall[myID]

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < hallButtonsPerFloor; btn++ {
			myHallOrder := GetHallOrder(myQueue, myID, floor, btn)
			currentState := myHallOrder.State
			expectedAssignedTo := myHallOrder.AssignedTo
			otherHallOrder := myHallOrder // Initialized as my own

			switch currentState {
			case None:
				amIAlone := true
				for ID, alive := range aliveElevators {
					if !alive || ID == myID {
						continue
					}
					amIAlone = false
					otherHallOrder = GetHallOrder(myQueue, ID, floor, btn)
					if otherHallOrder.State == Completed {
						return false
					}
					if hallOrders[floor][btn].State < otherHallOrder.State {
						hallOrders[floor][btn].State = otherHallOrder.State
						hallOrders[floor][btn].AssignedTo = otherHallOrder.AssignedTo
					}
				}
				if amIAlone {
					hallOrders[floor][btn].State = Unconfirmed
					hallOrders[floor][btn].AssignedTo = otherHallOrder.AssignedTo
				}
				myQueue.Hall[myID] = hallOrders
				return true

			case Unconfirmed:
				for ID, alive := range aliveElevators {
					if !alive || ID == myID {
						continue
					}
					otherHallOrder = GetHallOrder(myQueue, ID, floor, btn)
					if otherHallOrder.State == None || otherHallOrder.State == Completed {
						myQueue.Hall[myID] = hallOrders // Ensuring we keep the lowest assignedTo ID even in transition failure
						return false
					}

					if otherHallOrder.AssignedTo < expectedAssignedTo && otherHallOrder.AssignedTo > noElevatorAssigned {
						expectedAssignedTo = otherHallOrder.AssignedTo
					}
				}

				hallOrders[floor][btn].State = Confirmed
				hallOrders[floor][btn].AssignedTo = expectedAssignedTo
				myQueue.Hall[myID] = hallOrders
				return true

			// Continue from here
			case Confirmed:
				for ID, alive := range aliveElevators {
					if !alive || ID == myID {
						continue
					}
					otherHallOrder = GetHallOrder(myQueue, ID, floor, btn)
					if otherHallOrder.State != Completed { // Must double check this
						return false
					}
				}
				hallOrders[floor][btn].State = Completed
				hallOrders[floor][btn].AssignedTo = noElevatorAssigned
				myQueue.Hall[myID] = hallOrders
				return true

			case Completed:
				amIAlone := true
				for ID, alive := range aliveElevators {
					if !alive || ID == myID {
						continue
					}
					amIAlone = false
					otherHallOrder = GetHallOrder(myQueue, ID, floor, btn)
					if otherHallOrder.State == Confirmed {
						return false
					}
					if hallOrders[floor][btn].State < otherHallOrder.State {
						hallOrders[floor][btn].State = otherHallOrder.State
						hallOrders[floor][btn].AssignedTo = otherHallOrder.AssignedTo
					}
				}
				if amIAlone {
					hallOrders[floor][btn].State = None
					hallOrders[floor][btn].AssignedTo = noElevatorAssigned
				}
				myQueue.Hall[myID] = hallOrders
				return true
			default:
				fmt.Println("Undefined order state: ", currentState)
				return false
			}
		}
	}
	return false
}

func (myQueue *OrderQueue) TransitionCabOrders(
	myID int,
	aliveElevators map[int]bool,
	floor int,
) bool {
	cabOrders := myQueue.Cab[myID]

	for floor := 0; floor < N_FLOORS; floor++ {
		myCabOrder := GetCabOrder(myQueue, myID, floor)
		currentState := myCabOrder.State
		expectedAssignedTo := myCabOrder.AssignedTo
		otherCabOrder := myCabOrder // Initialized as my own

		switch currentState {
		case None:
			amIAlone := true
			for ID, alive := range aliveElevators {
				if !alive || ID == myID {
					continue
				}
				amIAlone = false
				otherCabOrder = GetCabOrder(myQueue, ID, floor)
				if otherCabOrder.State == Completed {
					return false
				}
				if cabOrders[floor].State < otherCabOrder.State {
					cabOrders[floor].State = otherCabOrder.State
				}
			}
			if amIAlone {
				cabOrders[floor].State = Unconfirmed
			}
			myQueue.Cab[myID] = cabOrders
			return true

		case Unconfirmed:
			for ID, alive := range aliveElevators {
				if !alive || ID == myID {
					continue
				}
				otherCabOrder = GetCabOrder(myQueue, ID, floor)
				if otherCabOrder.State == None || otherCabOrder.State == Completed {
					return false
				}
			}

			cabOrders[floor].State = Confirmed
			myQueue.Cab[myID] = cabOrders
			return true

		// Continue from here
		case Confirmed:
			for ID, alive := range aliveElevators {
				if !alive || ID == myID {
					continue
				}
				otherCabOrder = GetCabOrder(myQueue, ID, floor)
				if otherCabOrder.State != Completed { // Must double check this
					return false
				}
			}
			cabOrders[floor].State = Completed
			myQueue.Cab[myID] = cabOrders
			return true

		case Completed:
			amIAlone := true
			for ID, alive := range aliveElevators {
				if !alive || ID == myID {
					continue
				}
				amIAlone = false
				otherCabOrder = GetCabOrder(myQueue, ID, floor)
				if otherCabOrder.State == Confirmed {
					return false
				}
				if cabOrders[floor].State < otherCabOrder.State {
					cabOrders[floor].State = otherCabOrder.State
				}
			}
			if amIAlone {
				cabOrders[floor].State = None
			}
			myQueue.Cab[myID] = cabOrders
			return true
		default:
			fmt.Println("Undefined order state: ", currentState)
			return false
		}
	}
	return false
}
