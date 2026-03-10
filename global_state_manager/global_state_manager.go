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

const MyId = 1 // !!! Må endres til å være IDen til den heisen som kjører denne instansen av global state manager

func handleTimeEvent(timeEvent TimerEvent, globalElevatorStates *ElevatorStates, aliveElevatorsMap map[int]bool) {
	switch timeEvent.Type {
	case TimerElevatorTimeout:
		globalElevatorStates.Peers[timeEvent.ElevatorID].Alive = false
		UpdateAliveElevatorsMap(*globalElevatorStates, aliveElevatorsMap)
	case TimerMovementStuck:
		globalElevatorStates.Peers[MyId].Alive = false
		UpdateAliveElevatorsMap(*globalElevatorStates, aliveElevatorsMap)
		//handle movement stuck, update global queue and elevator states, and send messages if needed
		//case TimerAcceptancetest --- IGNORE ---
	}
}

func handleRecievedMessage(recievedMessage Message, globalQueue *OrderQueue, globalElevatorStates *ElevatorStates) {
	//handle network listener event
	globalElevatorStates.UpdateElevatorState(recievedMessage.Peer)
	//globalQueue.UpdateOrderQueue(recievedMessage.OrderQueue, recievedMessage.ID)
}

func handleThisElevatorUpdate(thisElevatorUpdate Elevator, globalQueue *OrderQueue, globalElevatorStates *ElevatorStates, prevMyElevatorQueue *[N_FLOORS][N_BUTTONS]bool, aliveElevatorsMap map[int]bool) {
	// kan det lages en funksjon for det under? Kanskje ligge i order queue?

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if (*prevMyElevatorQueue)[floor][btn] && thisElevatorUpdate.Requests[floor][btn] == false {
				globalQueue.CompleteMyOrder(ButtonEvent{Floor: floor, Button: ButtonType(btn)}, aliveElevatorsMap, MyId)
			}
		}
	}
	*prevMyElevatorQueue = thisElevatorUpdate.Requests

	globalElevatorStates.UpdateElevatorState(ThisElevatorToElevatorPeer(thisElevatorUpdate, MyId))
}

func handleButtonEvent(buttonEvent ButtonEvent, globalQueue *OrderQueue, globalElevatorStates ElevatorStates, aliveElevatorsMap map[int]bool) {
	assignTo := AssignNewOrder(buttonEvent, globalElevatorStates, globalQueue.Cab, MyId)
	globalQueue.AppendNewOrder(buttonEvent, MyId, aliveElevatorsMap, assignTo)
}

func Global_State_Manager(
	timerEventChan <-chan TimerEvent,
	recievedMessageChan <-chan Message,
	thisElevatorUpdateChan <-chan Elevator,
	buttonEventChan <-chan ButtonEvent,
	myOrderListChan chan<- [N_FLOORS][N_BUTTONS]bool,
	updateElevatorStateEvent chan<- ElevatorPeer,
	updateOrderQueueEvent chan<- OrderQueue) {

	// !!! er der her man skal ha backupPhase() og listen for other queuepahse()?

	//init forskjellige ting
	globalQueue := GenerateEmptyOrderQueue()
	globalElevatorStates := GenerateNewElevatorStates()
	aliveElevatorsMap := map[int]bool{1: false, 2: false, 3: false} //map for å holde styr på hvilke heiser som er alive, oppdateres i handleTimeEvent og handleRecievedMessage
	aliveElevatorsMap[MyId] = true
	prevMyElevatorQueue := [N_FLOORS][N_BUTTONS]bool{} //!!! egen init?

	for {
		select {
		case timerEvent := <-timerEventChan:
			handleTimeEvent(timerEvent, &globalElevatorStates, aliveElevatorsMap)
			updateElevatorStateEvent <- globalElevatorStates.Peers[MyId]

			// !!! Må oppdatere aliveElevatorsMap her, fordi det er kun her de dør

		case recievedMessage := <-recievedMessageChan:
			handleRecievedMessage(recievedMessage, &globalQueue, &globalElevatorStates)
			myOrderListChan <- globalQueue.RetrieveMyOrders(MyId)
			updateOrderQueueEvent <- globalQueue

		case thisElevatorUpdate := <-thisElevatorUpdateChan:
			handleThisElevatorUpdate(thisElevatorUpdate, &globalQueue, &globalElevatorStates, &prevMyElevatorQueue, aliveElevatorsMap)
			updateElevatorStateEvent <- globalElevatorStates.Peers[MyId]
			updateOrderQueueEvent <- globalQueue

		case buttonEvent := <-buttonEventChan:
			handleButtonEvent(buttonEvent, &globalQueue, globalElevatorStates, aliveElevatorsMap)
			myOrderListChan <- globalQueue.RetrieveMyOrders(MyId)
			updateOrderQueueEvent <- globalQueue
		}
	}
}
