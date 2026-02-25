package main

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/network"
	. "distributed_elevator/request_queue"
	"fmt"
)

func main() {
	// main given in driver-go
	/*
		numFloors := 4

		elevio.Init("localhost:15657", numFloors)

		var d elevio.MotorDirection = elevio.MD_Up
		//elevio.SetMotorDirection(d)

		drv_buttons := make(chan elevio.ButtonEvent)
		drv_floors := make(chan int)
		drv_obstr := make(chan bool)
		drv_stop := make(chan bool)

		go elevio.PollButtons(drv_buttons)
		go elevio.PollFloorSensor(drv_floors)
		go elevio.PollObstructionSwitch(drv_obstr)
		go elevio.PollStopButton(drv_stop)

		for {
			select {
			case a := <-drv_buttons:
				fmt.Printf("%+v\n", a)
				elevio.SetButtonLamp(a.Button, a.Floor, true)

			case a := <-drv_floors:
				fmt.Printf("%+v\n", a)
				if a == numFloors-1 {
					d = elevio.MD_Down
				} else if a == 0 {
					d = elevio.MD_Up
				}
				elevio.SetMotorDirection(d)

			case a := <-drv_obstr:
				fmt.Printf("%+v\n", a)
				if a {
					elevio.SetMotorDirection(elevio.MD_Stop)
				} else {
					elevio.SetMotorDirection(d)
				}

			case a := <-drv_stop:
				fmt.Printf("%+v\n", a)
				for f := 0; f < numFloors; f++ {
					for b := elevio.ButtonType(0); b < 3; b++ {
						elevio.SetButtonLamp(b, f, false)
					}
				}
			}
		}*/

	fmt.Println("Started!")

	// Creating communication channels
	newPeerCh := make(chan string)
	elevatorStateCh := make(chan Elevator)
	requestQueueCh := make(chan RequestQueue)

	updateQueueCh := make(chan [N_FLOORS][N_BUTTONS]bool)
	newButtonPress := make(chan [2]int)

	// Starting goroutines
	go Network_ListenerFSM(newPeerCh)
	go Network_SenderFSM(elevatorStateCh, requestQueueCh)

	go Elevalgo_ElevatorControllerLoop(updateQueueCh, newButtonPress)
}
