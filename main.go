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
	myID := *idFlag
	primaryPID := *pidFlag

	fmt.Println("Starting elevator with ID:", myID)

	Init("localhost:15656", N_FLOORS)

	BackupPhase(myID, primaryPID)

	// Creating communication channels
	newButtonEvent := make(chan ButtonEvent)
	newFloorEvent := make(chan int)
	stopEvent := make(chan bool)
	obstrEvent := make(chan bool)

	stateToGSM := make(chan Elevator, 1)
	stateToSupervisor := make(chan Elevator, 1)

	peerAlive := make(chan int)
	receivedMessage := make(chan Message)
	newElevStateToSend := make(chan ElevatorPeer)
	newOrderQueueToSend := make(chan OrderQueue)
	heartBeatPing := make(chan int)

	updateMyQueue := make(chan [N_FLOORS][N_BUTTONS]bool)

	supervisorEvent := make(chan SupervisorEvent)

	// Starting goroutines

	// IO goroutines
	go PollButtons(newButtonEvent)
	go PollFloorSensor(newFloorEvent)
	go PollObstructionSwitch(obstrEvent)
	go PollStopButton(stopEvent)

	// Elevator algorithm goroutines
	go ElevatorController(updateMyQueue,
		newFloorEvent,
		stopEvent,
		obstrEvent,
		newButtonEvent,
		stateToGSM,
		stateToSupervisor)

	// Network goroutines
	go NetworkListener(myID,
		peerAlive,
		receivedMessage)

	go NetworkSender(myID,
		newElevStateToSend,
		newOrderQueueToSend,
		heartBeatPing)

	// GSM goroutines
	go Global_State_Manager(myID,
		supervisorEvent,
		receivedMessage,
		stateToGSM,
		newButtonEvent,
		updateMyQueue,
		newElevStateToSend,
		newOrderQueueToSend,
		heartBeatPing)

	// Supervisor goroutines
	go Supervisor(peerAlive,
		stateToSupervisor,
		supervisorEvent)

	// TEST ZONE
	//TestOrderQueue()
	//TestCostLogic()

	select {}
}
