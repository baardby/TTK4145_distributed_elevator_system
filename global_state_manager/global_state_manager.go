package global_state_manager

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
	. "distributed_elevator/global_state_manager/cost_fns"
	. "distributed_elevator/global_state_manager/elevator_states"
	. "distributed_elevator/global_state_manager/order_queue"
	. "distributed_elevator/network/message"
	. "distributed_elevator/supervisor"
	"fmt"
	"time"
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

	handleHallLights(globalQueue, myId)

	// TODO: Remove this after testing
	for k, v := range globalQueue.Hall {
		fmt.Printf("%6v :  %+v\n", k, v)
	}
	for k, v := range globalQueue.Cab {
		fmt.Printf("%6v :  %+v\n", k, v)
	}
	// END OF TODO

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
	fmt.Println("Handling Update Elevator Event") // TODO: Remove after testing

	completed := true
	*prevMyElevatorQueue = thisElevator.Requests
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			/*
				if (*prevMyElevatorQueue)[floor][btn] && !thisElevator.Requests[floor][btn] { // TODO: Check if this can be built upon
					fmt.Println("Trying to complete")
					completed = globalQueue.CompleteMyOrder(ButtonEvent{Floor: floor, Button: ButtonType(btn)}, *globalElevatorStates, myId)
					// Added
					if !completed {
						(*prevMyElevatorQueue)[floor][btn] = true
					} else {
						(*prevMyElevatorQueue)[floor][btn] = false
					}
					// End of added
				}*/
			// New modification
			switch ButtonType(btn) {
			case BT_HallUp, BT_HallDown:
				if (globalQueue.Hall[myId][floor][btn].State == Confirmed) && (globalQueue.Hall[myId][floor][btn].AssignedTo == myId) && !thisElevator.Requests[floor][btn] && thisElevator.Floor == floor {
					fmt.Println("Trying to complete")
					completed = globalQueue.CompleteMyOrder(ButtonEvent{Floor: floor, Button: ButtonType(btn)}, *globalElevatorStates, myId)
					if !completed {
						(*prevMyElevatorQueue)[floor][btn] = true
					} else {
						(*prevMyElevatorQueue)[floor][btn] = false
					}
				}
			case BT_Cab:
				if (globalQueue.Cab[myId][floor][myId] == Confirmed) && !thisElevator.Requests[floor][btn] && thisElevator.Floor == floor {
					fmt.Println("Trying to complete")
					completed = globalQueue.CompleteMyOrder(ButtonEvent{Floor: floor, Button: ButtonType(btn)}, *globalElevatorStates, myId)
					// Added
					if !completed {
						(*prevMyElevatorQueue)[floor][btn] = true
					} else {
						(*prevMyElevatorQueue)[floor][btn] = false
						fmt.Println("Completed")
					}
					// End of added
				}
			} // END OF New modification
		}
	}
	//if completed { // TODO: Double check if this can be built upon
	//	*prevMyElevatorQueue = thisElevator.Requests
	//}
	globalElevatorStates.UpdatePeer(ThisElevatorToElevatorPeer(thisElevator, myId), myId)
	return completed
}

func handleButtonEvent(buttonEvent ButtonEvent, globalQueue *OrderQueue, globalElevatorStates ElevatorStates, myId int) {
	assignTo := AssignNewOrder(buttonEvent, globalElevatorStates, globalQueue.Cab[myId], myId)
	globalQueue.AppendNewOrder(buttonEvent, myId, globalElevatorStates, assignTo)
}

func handleHallLights(globalQueue *OrderQueue, myId int) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS-1; btn++ {
			if globalQueue.Hall[myId][floor][btn].State == Confirmed {
				SetButtonLamp(ButtonType(btn), floor, true)
			} else {
				SetButtonLamp(ButtonType(btn), floor, false)
			}
		}
	}
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

	updateOrderListTicker := time.NewTicker(100 * time.Millisecond) // TODO: Change to 10Hz
	defer updateOrderListTicker.Stop()

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

		case <-updateOrderListTicker.C:
			globalQueue.TransitionAllCabOrders(myId, globalElevatorStates)
			globalQueue.TransitionAllHallOrders(myId, globalElevatorStates)
			handleHallLights(&globalQueue, myId)
			myOrderListChan <- globalQueue.RetrieveMyOrders(myId)
		}
	}
}
