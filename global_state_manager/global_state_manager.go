package global_state_manager

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
	. "distributed_elevator/global_state_manager/cost_fns"
	. "distributed_elevator/global_state_manager/elevator_states"
	. "distributed_elevator/global_state_manager/order_queue"
	. "distributed_elevator/network/message"
	. "distributed_elevator/supervisor"
)

func handleSupervisorEvent(
	supervisorEvent SupervisorEvent,
	globalQueue *OrderQueue,
	globalElevatorStates *ElevatorStates,
	myId int) {

	switch supervisorEvent.Type {
	case TimerElevatorTimeout:
		globalElevatorStates.Peers[supervisorEvent.ElevatorID].WorkingStatus = StatusLostConnection
		if lowestIDOnNetwork(*globalElevatorStates) == myId {
			globalQueue.RedistributeHallOrders(myId, *globalElevatorStates, AssignNewOrder)
		}
	case SupervisorHardwareFault:
		globalElevatorStates.Peers[myId].WorkingStatus = StatusHardwareFault
		if lowestIDOnNetwork(*globalElevatorStates) == myId {
			globalQueue.RedistributeHallOrders(myId, *globalElevatorStates, AssignNewOrder)
		}
	case SupervisorHardwareRecovered:
		globalElevatorStates.Peers[myId].WorkingStatus = StatusOK
	}
}

func handleReceivedMessage(
	receivedMessage Message,
	globalQueue *OrderQueue,
	globalElevatorStates *ElevatorStates,
	myId int) {

	oldPeer := globalElevatorStates.Peers[receivedMessage.Peer.ID]
	globalElevatorStates.UpdatePeer(receivedMessage.Peer, myId)

	globalQueue.UpdateOrderQueue(receivedMessage.HallOrders, receivedMessage.CabOrders, receivedMessage.ID)
	globalQueue.TransitionAllHallOrders(myId, *globalElevatorStates)
	globalQueue.TransitionAllCabOrders(myId, *globalElevatorStates)

	needRedistribute := fromOkToHardwareFault(receivedMessage.Peer, oldPeer)
	if needRedistribute && lowestIDOnNetwork(*globalElevatorStates) == myId {
		globalQueue.RedistributeHallOrders(myId, *globalElevatorStates, AssignNewOrder)
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
	thisElevator Elevator,
	globalQueue *OrderQueue,
	globalElevatorStates *ElevatorStates,
	prevMyElevatorQueue *[N_FLOORS][N_BUTTONS]bool,
	myId int) bool {

	completed := true
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if (*prevMyElevatorQueue)[floor][btn] && !thisElevator.Requests[floor][btn] {
				completed = globalQueue.CompleteMyOrder(ButtonEvent{Floor: floor, Button: ButtonType(btn)}, *globalElevatorStates, myId)
			}
		}
	}
	if completed {
		*prevMyElevatorQueue = thisElevator.Requests
	}
	globalElevatorStates.UpdatePeer(ThisElevatorToElevatorPeer(thisElevator, myId), myId)
	return completed
}

func handleButtonEvent(buttonEvent ButtonEvent, globalQueue *OrderQueue, globalElevatorStates ElevatorStates, myId int) {
	assignTo := AssignNewOrder(buttonEvent, globalElevatorStates, globalQueue.Cab[myId], myId)
	globalQueue.AppendNewOrder(buttonEvent, myId, globalElevatorStates, assignTo)
}

func Global_State_Manager(
	myId int,
	supervisorEventChan <-chan SupervisorEvent,
	receivedMessageChan <-chan Message,
	thisElevatorUpdateChan <-chan Elevator,
	buttonEventChan <-chan ButtonEvent,
	myOrderListChan chan<- [N_FLOORS][N_BUTTONS]bool,
	updateElevatorStateEvent chan<- ElevatorPeer,
	updateOrderQueueEvent chan<- OrderQueue) {

	// !!! er der her man skal ha backupPhase() og listen for other queuepahse()?

	//init forskjellige ting
	globalQueue := GenerateNewOrderQueue()
	globalElevatorStates := GenerateNewElevatorStates(myId)
	prevMyElevatorQueue := [N_FLOORS][N_BUTTONS]bool{}

	for {
		select {
		case supervisorEvent := <-supervisorEventChan:
			handleSupervisorEvent(supervisorEvent, &globalQueue, &globalElevatorStates, myId)
			updateElevatorStateEvent <- globalElevatorStates.Peers[myId]

		case receivedMessage := <-receivedMessageChan:
			handleReceivedMessage(receivedMessage, &globalQueue, &globalElevatorStates, myId)
			myOrderListChan <- globalQueue.RetrieveMyOrders(myId)
			updateOrderQueueEvent <- globalQueue

		case thisElevatorUpdate := <-thisElevatorUpdateChan:
			couldCompleteOrder := handleThisElevatorUpdate(thisElevatorUpdate, &globalQueue, &globalElevatorStates, &prevMyElevatorQueue, myId)
			if !couldCompleteOrder {
				myOrderListChan <- prevMyElevatorQueue
			}
			updateElevatorStateEvent <- globalElevatorStates.Peers[myId]
			updateOrderQueueEvent <- globalQueue

		case buttonEvent := <-buttonEventChan:
			handleButtonEvent(buttonEvent, &globalQueue, globalElevatorStates, myId)
			myOrderListChan <- globalQueue.RetrieveMyOrders(myId)
			updateOrderQueueEvent <- globalQueue
		}
	}
}
