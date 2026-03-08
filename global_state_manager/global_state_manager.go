package global_state_manager

import (
	. "distributed_elevator/HealthTimers"
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
	. "distributed_elevator/network"
	. "distributed_elevator/supervisor"
)

func handleTimeEvent(timeEvent TimerEvent, globalQueue *OrderQueue, globalElevatorStates *ElevatorStates) {
	//handle time event
}

func handleSingleElevatorUpdate(singleElevatorUpdate SingleElevatorUpdate) {
	//handle hardware event
}

func handleNetwork_ListenerEvent(network_listenerEvent Message, globalQueue *OrderQueue, globalElevatorStates *ElevatorStates) {
	//handle network listener even
	//globalElevatorStates.UpdateElevatorStates()
	//globalQueue.UpdateOrderQueue(message.OrderQueue, message.ID)
}

func Global_State_Manager(
	recievedMessageChan <-chan Message,
	timerEvent <-chan TimerEvent,
	singleElevatorUpdate <-chan SingleElevatorUpdate,
	singleElevatorCommand chan<- SingleElevatorCommand,
	sendMessageChan chan<- NetworkMessage) {

	//er der her man skal ha backupPhase() og listen for other queuepahse()?

	//init forskjellige ting
	globalQueue := generateNewOrderQueue()
	//globalElevatorStates := ElevatorStates{} //egen init funksjon?

	for {
		select {
		case timerEvent := <-timerEvent:
			handleTimeEvent(timerEvent, &globalQueue /*&globalElevatorStates*/)
		case recievedMessage := <-recievedMessageChan:
			handleNetwork_ListenerEvent(recievedMessage, &globalQueue /*&globalElevatorStates*/)
		case singleElevatorUpdate := <-singleElevatorUpdate:
			handleSingleElevatorUpdate(singleElevatorUpdate)
			//update global queue and elevator states based on update from single elevator, and send messages if needed

		}
	}
}
