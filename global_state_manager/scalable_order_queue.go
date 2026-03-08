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
)

type OrderState int

const (
	None        OrderState = iota // Order completed by all
	Unconfirmed                   // Order confirmed by at least one
	Confirmed                     // Order confirmed by all
	Completed                     // Order completed by at least one
)

type HallOrder struct {
	State      OrderState
	AssignedTo int
}

type AllHallOrders [N_FLOORS][hallButtonsPerFloor]HallOrder

type OrderQueue struct {
	Hall map[int]AllHallOrders        // elevatorID -> that elevator's view of all hall orders
	Cab  map[int][N_FLOORS]OrderState // elevatorID -> that elevator's view of cab orders
}

func GenerateEmptyOrderQueue() OrderQueue {
	return OrderQueue{
		Hall: make(map[int]AllHallOrders),
		Cab:  make(map[int][N_FLOORS]OrderState),
	}
}

// This function must be used before other functions to ensure no errors
func AddElevatorToQueue(queue *OrderQueue, ID int) {
	if _, ok := queue.Hall[ID]; !ok {
		queue.Hall[ID] = AllHallOrders{}
		queue.Cab[ID] = [N_FLOORS]OrderState{}
	}
}

func GetHallOrder(queue *OrderQueue, ID int, floor int, btn int) HallOrder {
	return queue.Hall[ID][floor][btn]
}

func UpdateHallOrder(queue *OrderQueue, ID int, floor int, btn int, order HallOrder) {
	hallOrders := queue.Hall[ID]
	hallOrders[floor][btn] = order
	queue.Hall[ID] = hallOrders
}

func CanOrderTransitionToNextState(
	queue *OrderQueue,
	aliveElevatorIDs []int,
	floor int,
	btn int,
	state OrderState,
) bool {
	for _, ID := range aliveElevatorIDs {
		order := GetHallOrder(queue, ID, floor, btn)
		if order.State == state {
			return false
		}
	}
	return true
}

// !!! Might need further implementation as currently only true
// if at least 1 elevator is None
func IsOrderAlreadyActive(
	queue *OrderQueue,
	aliveElevatorIDs []int,
	floor int,
	btn int,
	state OrderState,
) bool {
	for _, ID := range aliveElevatorIDs {
		order := GetHallOrder(queue, ID, floor, btn)
		if order.State != None {
			return true
		}
	}
	return false
}

// Function that appends new order, must call cost function to find assignTo parameter
func (queue *OrderQueue) AppendNewOrder(btnEv ButtonEvent, assignTo int, myID int, aliveElevatorIDs []int) {
	floor := btnEv.Floor
	btn := int(btnEv.Button)

	if floor < 0 || floor >= N_FLOORS {
		fmt.Println("Attempted to append order at invalid floor: ", floor)
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
	if IsOrderAlreadyActive(queue, aliveElevatorIDs, floor, btn) {
		fmt.Println("Order is already active.")
		return
	}

}
