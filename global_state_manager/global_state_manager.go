package global_state_manager

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
	. "distributed_elevator/global_state_manager/cost_fns"
	. "distributed_elevator/global_state_manager/elevator_states"
	. "distributed_elevator/global_state_manager/order_queue"
	. "distributed_elevator/network/message"
	. "distributed_elevator/supervisor"
	"time"
)

func GlobalStateManager(
	myID int,
	supervisorEvent <-chan SupervisorEvent,
	receivedMessage <-chan Message,
	updateMyElevator <-chan Elevator,
	newButtonEvent <-chan ButtonEvent,
	updateMyQueue chan<- [N_FLOORS][N_BUTTONS]bool,
	newElevStateToSend chan<- ElevatorPeer,
	newOrderQueueToSend chan<- OrderQueue,
	heartBeatPing chan<- int) {

	globalQueue := GenerateNewOrderQueue()
	globalElevatorStates := GenerateNewElevatorStates(myID)
	prevMyElevatorQueue := [N_FLOORS][N_BUTTONS]bool{}

	updateOrderListTicker := time.NewTicker(100 * time.Millisecond)
	defer updateOrderListTicker.Stop()

	heartBeatTicker := time.NewTicker(500 * time.Millisecond)
	defer heartBeatTicker.Stop()

	for {
		select {
		case supervisorEvent := <-supervisorEvent:
			handleSupervisorEvent(myID, supervisorEvent, &globalQueue, &globalElevatorStates)
			newElevStateToSend <- globalElevatorStates.Peers[myID]

		case receivedMessage := <-receivedMessage:
			handleReceivedMessage(myID, receivedMessage, &globalQueue, &globalElevatorStates)
			updateMyQueue <- globalQueue.RetrieveMyOrders(myID)
			newOrderQueueToSend <- globalQueue

		case thisElevatorUpdate := <-updateMyElevator:
			couldCompleteOrder := handleThisElevatorUpdate(myID, thisElevatorUpdate, &globalQueue, &globalElevatorStates, &prevMyElevatorQueue)
			if !couldCompleteOrder {
				updateMyQueue <- prevMyElevatorQueue
			}
			newElevStateToSend <- globalElevatorStates.Peers[myID]
			newOrderQueueToSend <- globalQueue

		case buttonEvent := <-newButtonEvent:
			handleButtonEvent(myID, buttonEvent, &globalQueue, globalElevatorStates)
			updateMyQueue <- globalQueue.RetrieveMyOrders(myID)
			newOrderQueueToSend <- globalQueue

		case <-updateOrderListTicker.C:
			globalQueue.TransitionAllCabOrders(myID, globalElevatorStates)
			globalQueue.TransitionAllHallOrders(myID, globalElevatorStates)
			handleHallLights(myID, &globalQueue)
			updateMyQueue <- globalQueue.RetrieveMyOrders(myID)

		case <-heartBeatTicker.C:
			heartBeatPing <- myID
		}
	}
}

func handleSupervisorEvent(
	myID int,
	supervisorEvent SupervisorEvent,
	globalQueue *OrderQueue,
	globalElevatorStates *ElevatorStates) {

	switch supervisorEvent.Type {
	case TimerElevatorTimeout:
		globalElevatorStates.Peers[supervisorEvent.ElevatorID].WorkingStatus = StatusLostConnection
		if lowestIDOnNetwork(*globalElevatorStates) == myID {
			globalQueue.RedistributeHallOrders(myID, *globalElevatorStates, AssignNewOrder)
		}
	case SupervisorHardwareFault:
		globalElevatorStates.Peers[myID].WorkingStatus = StatusHardwareFault
		if lowestIDOnNetwork(*globalElevatorStates) == myID {
			globalQueue.RedistributeHallOrders(myID, *globalElevatorStates, AssignNewOrder)
		}
	case SupervisorHardwareRecovered:
		globalElevatorStates.Peers[myID].WorkingStatus = StatusOK
	}
}

func handleReceivedMessage(
	myID int,
	receivedMessage Message,
	globalQueue *OrderQueue,
	globalElevatorStates *ElevatorStates) {

	oldPeer := globalElevatorStates.Peers[receivedMessage.Peer.ID]
	globalElevatorStates.UpdatePeer(receivedMessage.Peer, myID)

	globalQueue.UpdateOrderQueue(receivedMessage.HallOrders, receivedMessage.CabOrders, receivedMessage.ID)
	globalQueue.TransitionAllHallOrders(myID, *globalElevatorStates)
	globalQueue.TransitionAllCabOrders(myID, *globalElevatorStates)

	handleHallLights(myID, globalQueue)

	needRedistribute := fromOkToHardwareFault(receivedMessage.Peer, oldPeer)
	if needRedistribute && lowestIDOnNetwork(*globalElevatorStates) == myID {
		globalQueue.RedistributeHallOrders(myID, *globalElevatorStates, AssignNewOrder)
	}

}

func fromOkToHardwareFault(newPeer ElevatorPeer, oldPeer ElevatorPeer) bool {
	if oldPeer.WorkingStatus == StatusOK && newPeer.WorkingStatus == StatusHardwareFault {
		return true
	}
	return false
}

func lowestIDOnNetwork(globalElevatorStates ElevatorStates) int {
	for i := 0; i < N_ELEVATORS; i++ {
		if globalElevatorStates.Peers[i].WorkingStatus != StatusLostConnection {
			return i
		}
	}
	return -1 //return -1 if no elevator is StatusOK
}

func handleThisElevatorUpdate( // Return false if order could not complete, true otherwise
	myID int,
	thisElevator Elevator,
	globalQueue *OrderQueue,
	globalElevatorStates *ElevatorStates,
	prevMyElevatorQueue *[N_FLOORS][N_BUTTONS]bool) bool {

	completed := true
	*prevMyElevatorQueue = thisElevator.Requests
	currentFloor := thisElevator.Floor
	for btn := 0; btn < N_BUTTONS; btn++ {
		switch ButtonType(btn) {
		case BT_HallUp, BT_HallDown:
			if (globalQueue.Hall[myID][currentFloor][btn].State == Confirmed) && (globalQueue.Hall[myID][currentFloor][btn].AssignedTo == myID) && !thisElevator.Requests[currentFloor][btn] {
				completed = globalQueue.CompleteMyOrder(ButtonEvent{Floor: currentFloor, Button: ButtonType(btn)}, *globalElevatorStates, myID)
				if !completed {
					(*prevMyElevatorQueue)[currentFloor][btn] = true
				} else {
					(*prevMyElevatorQueue)[currentFloor][btn] = false
				}
			}

		case BT_Cab:
			if (globalQueue.Cab[myID][currentFloor][myID] == Confirmed) && !thisElevator.Requests[currentFloor][btn] {
				completed = globalQueue.CompleteMyOrder(ButtonEvent{Floor: currentFloor, Button: ButtonType(btn)}, *globalElevatorStates, myID)
				if !completed {
					(*prevMyElevatorQueue)[currentFloor][btn] = true
				} else {
					(*prevMyElevatorQueue)[currentFloor][btn] = false
				}
			}
		}
	}
	globalElevatorStates.UpdatePeer(ThisElevatorToElevatorPeer(thisElevator, myID), myID)
	return completed
}

func handleButtonEvent(myID int, buttonEvent ButtonEvent, globalQueue *OrderQueue, globalElevatorStates ElevatorStates) {
	assignTo := AssignNewOrder(buttonEvent, globalElevatorStates, globalQueue.Cab[myID], myID)
	globalQueue.AppendNewOrder(buttonEvent, myID, globalElevatorStates, assignTo)
}

func handleHallLights(myID int, globalQueue *OrderQueue) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS-1; btn++ {
			if globalQueue.Hall[myID][floor][btn].State == Confirmed {
				SetButtonLamp(ButtonType(btn), floor, true)
			} else {
				SetButtonLamp(ButtonType(btn), floor, false)
			}
		}
	}
}
