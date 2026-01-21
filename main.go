package main

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
	"fmt"
	"time"
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

	var elevator Elevator = Elevator_Uninitialized()
	var inputPollRate_ms int = 25

	// con_load

	if Elevator_FloorSensor() == -1 {
		Fsm_OnInitBetweenFloors(&elevator)
	}
	prevFloor := -1
	var prevOrder [N_FLOORS][N_BUTTONS]bool

	for {
		// Request button
		for floor := 0; floor < N_FLOORS; floor++ {
			for btn := 0; btn < N_BUTTONS; btn++ {
				value := Elevator_RequestButton(floor, ButtonType(btn))
				if value && value != prevOrder[floor][btn] {
					Fsm_OnRequestButtonPress(&elevator, floor, ButtonType(btn))
				}
				prevOrder[floor][btn] = value
			}
		}

		// Floor sensor
		floor := Elevator_FloorSensor()
		if floor != -1 && floor != prevFloor {
			Fsm_OnFloorArrival(&elevator, floor)
		}
		prevFloor = floor

		// Timer
		if Timer_TimedOut() {
			Timer_Stop()
			Fsm_OnDoorTimeout(&elevator)
		}

		time.Sleep(time.Duration(inputPollRate_ms) * time.Millisecond)
	}
}
