package main

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
	. "distributed_elevator/global_state_manager/elevator_states"
	. "distributed_elevator/global_state_manager/order_queue"
	. "distributed_elevator/network"
	. "distributed_elevator/network/message"
	"flag"
	"fmt"
	"os"
)

func main() {

	// Initializing Elevator ID
	idFlag := flag.Int("id", -1, "elevator ID (1..N_ELEVATORS)")
	flag.Parse()

	if *idFlag < 1 || *idFlag > N_ELEVATORS {
		fmt.Fprintf(os.Stderr, "error: --id must be in range 0..%d\n", N_ELEVATORS)
		os.Exit(2)
	}
	ID := *idFlag

	fmt.Println("Starting elevator with ID:", ID)

	Init("localhost:15657", N_FLOORS)

	// Creating communication channels
	newButtonEvent := make(chan ButtonEvent)
	newFloorEvent := make(chan int)
	stopEvent := make(chan bool)
	obstrEvent := make(chan bool)
	updateElevatorEvent := make(chan Elevator)

	receivedFromPeerEvent := make(chan int)
	receivedMessageEvent := make(chan Message)
	newElevStateToSendEvent := make(chan ElevatorPeer)
	newRequestQueueToSendEvent := make(chan OrderQueue)

	updateQueueEvent := make(chan [N_FLOORS][N_BUTTONS]bool)

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
		updateElevatorEvent)

	// Network goroutines
	go Network_ListenerLoop(ID,
		receivedFromPeerEvent,
		receivedMessageEvent)
	go Network_SenderLoop(ID,
		newElevStateToSendEvent,
		newRequestQueueToSendEvent)

	select {}
}
