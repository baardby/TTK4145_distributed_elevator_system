package global_state_manager

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
	. "distributed_elevator/supervisor"
	"fmt"
)

// !!! Variable and function names are hard

const (
	noElevatorAssigned  = 0
	hallButtonsPerFloor = 2
	numOfOrderStates    = 4
)

type OrderState int

// !!! Better comments
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
type AllCabOrders [N_FLOORS]Order

type OrderQueue struct {
	Hall map[int]AllHallOrders // elevatorID -> that elevator's view of all hall orders
	Cab  map[int]AllCabOrders  // elevatorID -> that elevator's view of cab orders
}

type MyOrderList struct {
	MyOrders [N_FLOORS][N_BUTTONS]bool
}

func (globalQueue *OrderQueue) GenerateMyOrderList(myID int) MyOrderList {
	var myOrderList MyOrderList
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if GetHallOrder(globalQueue, myID, floor, btn).AssignedTo == myID || GetCabOrder(globalQueue, myID, floor).AssignedTo == myID {
				myOrderList.MyOrders[floor][btn] = true
			}
		}
	}
	return myOrderList
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
func (queue *OrderQueue) AppendNewOrder(btnEv ButtonEvent, myID int, aliveElevators map[int]bool) {
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

	if btnEv.Button != BT_Cab {
		// !!! Placeholder function
		assignTo := calculateCost(queue, aliveElevators, floor, btn)

		if assignTo < 0 || assignTo >= N_ELEVATORS {
			fmt.Println("Attempted to append invalid assignedTo: ", assignTo)
			return
		}

		queue.Hall[myID][floor][btn] = Order{
			State:      Unconfirmed,
			AssignedTo: assignTo,
		}
	} else {
		queue.Cab[myID][floor][btn] = Order{
			State:      Unconfirmed,
			AssignedTo: myID,
		}
	}
}

func CanOrderTransitionState(
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
			if !alive {
				continue
			}
			if GetHallOrder(queue, ID, floor, btn).State == Unconfirmed {
				return true
			}
		}
		return false

	case Unconfirmed:
		for ID, alive := range aliveElevators {
			if !alive {
				continue
			}
			if GetHallOrder(queue, ID, floor, btn).State == None || GetHallOrder(queue, ID, floor, btn).State == Completed {
				return false
			}
		}
		return true

	case Confirmed:
		for ID, alive := range aliveElevators {
			if !alive {
				continue
			}
			if GetHallOrder(queue, ID, floor, btn).State == Completed { // Must double check this
				return true
			}
		}
		return false

	case Completed:
		for ID, alive := range aliveElevators {
			if !alive {
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

func (myQueue *OrderQueue) TransitionQueue(myID int, aliveElevators map[int]bool) {
	hallOrders := myQueue.Hall[myID]

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < hallButtonsPerFloor; btn++ {
			if CanOrderTransitionState(myQueue, myID, aliveElevators, floor, btn) {
				currentState := GetHallOrder(myQueue, myID, floor, btn).State
				nextState := (currentState + 1) % numOfOrderStates
				hallOrders[floor][btn].State = nextState
			}
		}
	}
	myQueue.Hall[myID] = hallOrders
}

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