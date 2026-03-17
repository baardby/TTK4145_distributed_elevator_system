package main

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
	. "distributed_elevator/global_state_manager"
	. "distributed_elevator/global_state_manager/elevator_states"
	. "distributed_elevator/global_state_manager/order_queue"
	. "distributed_elevator/network"
	. "distributed_elevator/network/message"
	. "distributed_elevator/supervisor"
	"flag"
	"fmt"
	"os"
)

func main() {

	// Initializing Elevator ID
	idFlag := flag.Int("id", -1, "elevator ID (0..N_ELEVATORS-1)")
	pidFlag := flag.Int("pid", -1, "PID of primary process")
	flag.Parse()

	if *idFlag < 0 || *idFlag > N_ELEVATORS-1 {
		fmt.Fprintf(os.Stderr, "error: --id must be in range 0..%d\n", N_ELEVATORS-1)
		os.Exit(2)
	}
	ID := *idFlag
	primaryPID := *pidFlag

	fmt.Println("Starting elevator with ID:", ID)

	Init("localhost:15656", N_FLOORS)

	BackupPhase(ID, primaryPID)

	// Creating communication channels
	newButtonEvent := make(chan ButtonEvent)
	newFloorEvent := make(chan int)
	stopEvent := make(chan bool)
	obstrEvent := make(chan bool)
	stateToGSM := make(chan Elevator, 1)
	stateToSupervisor := make(chan Elevator, 1)

	receivedFromPeerEvent := make(chan int)
	receivedMessageEvent := make(chan Message)
	newElevStateToSendEvent := make(chan ElevatorPeer)
	newOrderQueueToSendEvent := make(chan OrderQueue)
	heartBeatPing := make(chan int)

	updateQueueEvent := make(chan [N_FLOORS][N_BUTTONS]bool)

	supervisorEvent := make(chan SupervisorEvent)

	// Starting goroutines

	// IO goroutines
	go PollButtons(newButtonEvent)
	go PollFloorSensor(newFloorEvent)
	go PollObstructionSwitch(obstrEvent)
	go PollStopButton(stopEvent)

	// Elevator algorithm goroutines
	go Elevalgo_ElevatorControllerLoop(updateQueueEvent,
		newFloorEvent,
		stopEvent,
		obstrEvent,
		newButtonEvent,
		stateToGSM,
		stateToSupervisor)

	// Network goroutines
	go Network_ListenerLoop(ID,
		receivedFromPeerEvent,
		receivedMessageEvent)

	go Network_SenderLoop(ID,
		newElevStateToSendEvent,
		newOrderQueueToSendEvent,
		heartBeatPing)

	// GSM goroutines
	go Global_State_Manager(ID,
		supervisorEvent,
		receivedMessageEvent,
		stateToGSM,
		newButtonEvent,
		updateQueueEvent,
		newElevStateToSendEvent,
		newOrderQueueToSendEvent,
		heartBeatPing)

	// Supervisor goroutines
	go Supervisor(receivedFromPeerEvent,
		stateToSupervisor,
		supervisorEvent)

	// TEST ZONE
	//TestOrderQueue()
	//TestCostLogic()

	select {}
}
