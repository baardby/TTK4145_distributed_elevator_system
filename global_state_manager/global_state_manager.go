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

func handleSupervisorEvent(supervisorEvent SupervisorEvent, globalElevatorStates *ElevatorStates, myId int) {
	switch supervisorEvent.Type {
	case TimerElevatorTimeout:
		globalElevatorStates.Peers[supervisorEvent.ElevatorID].WorkingStatus = StatusLostConnection
	case SupervisorHardwareFault:
		globalElevatorStates.Peers[myId].WorkingStatus = StatusHardwareFault
	case SupervisorHardwareRecovered:
		globalElevatorStates.Peers[myId].WorkingStatus = StatusOK
	}
}

func handleReceivedMessage(receivedMessage Message, globalQueue *OrderQueue, globalElevatorStates *ElevatorStates, myId int) {
	globalElevatorStates.UpdatePeer(receivedMessage.Peer, myId)

	AddElevatorToQueue(globalQueue, receivedMessage.ID)

	globalQueue.UpdateOrderQueue(receivedMessage.HallOrders, receivedMessage.CabOrders, receivedMessage.ID)
	globalQueue.TransitionAllHallOrders(myId, *globalElevatorStates)
	globalQueue.TransitionAllCabOrders(myId, *globalElevatorStates)
}

func handleThisElevatorUpdate(thisElevator Elevator, globalQueue *OrderQueue, globalElevatorStates *ElevatorStates, prevMyElevatorQueue *[N_FLOORS][N_BUTTONS]bool, myId int) {

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if (*prevMyElevatorQueue)[floor][btn] && !thisElevator.Requests[floor][btn] {
				globalQueue.CompleteMyOrder(ButtonEvent{Floor: floor, Button: ButtonType(btn)}, *globalElevatorStates, myId)
			}
		}
	}
	*prevMyElevatorQueue = thisElevator.Requests

	globalElevatorStates.UpdatePeer(ThisElevatorToElevatorPeer(thisElevator, myId), myId)
}

func handleButtonEvent(buttonEvent ButtonEvent, globalQueue *OrderQueue, globalElevatorStates ElevatorStates, myId int) {
	assignTo := AssignNewOrder(buttonEvent, globalElevatorStates, globalQueue.Cab[myId], myId)
	globalQueue.AppendNewOrder(buttonEvent, myId, globalElevatorStates, assignTo)
}

func Global_State_Manager(
	myId int,
	supervisorEventChan <-chan SupervisorEvent,
	recievedMessageChan <-chan Message,
	thisElevatorUpdateChan <-chan Elevator,
	buttonEventChan <-chan ButtonEvent,
	myOrderListChan chan<- [N_FLOORS][N_BUTTONS]bool,
	updateElevatorStateEvent chan<- ElevatorPeer,
	updateOrderQueueEvent chan<- OrderQueue) {

	// !!! er der her man skal ha backupPhase() og listen for other queuepahse()?

	//init forskjellige ting
	globalQueue := GenerateEmptyOrderQueue()
	globalElevatorStates := GenerateNewElevatorStates(myId)
	prevMyElevatorQueue := [N_FLOORS][N_BUTTONS]bool{}

	for {
		select {
		case supervisorEvent := <-supervisorEventChan:
			handleSupervisorEvent(supervisorEvent, &globalElevatorStates, myId)
			updateElevatorStateEvent <- globalElevatorStates.Peers[myId-1]

		case recievedMessage := <-recievedMessageChan:
			handleReceivedMessage(recievedMessage, &globalQueue, &globalElevatorStates, myId)
			myOrderListChan <- globalQueue.RetrieveMyOrders(myId)
			updateOrderQueueEvent <- globalQueue

		case thisElevatorUpdate := <-thisElevatorUpdateChan:
			handleThisElevatorUpdate(thisElevatorUpdate, &globalQueue, &globalElevatorStates, &prevMyElevatorQueue, myId)
			updateElevatorStateEvent <- globalElevatorStates.Peers[myId-1]
			updateOrderQueueEvent <- globalQueue

		case buttonEvent := <-buttonEventChan:
			handleButtonEvent(buttonEvent, &globalQueue, globalElevatorStates, myId)
			myOrderListChan <- globalQueue.RetrieveMyOrders(myId)
			updateOrderQueueEvent <- globalQueue
		}
	}
}
