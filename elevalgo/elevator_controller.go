package elevalgo //Må endres hvis det puttes inn i en mappe

import (
	. "distributed_elevator/elevio"
)

func Elevalgo_ElevatorControllerLoop(updateQueue <-chan [N_FLOORS][N_BUTTONS]bool,
	drvFloors <-chan int,
	drvObstr <-chan bool,
	drvStop <-chan bool) {

	var elevator Elevator = Elevator_Uninitialized()

	for {
		select {
		case newRequests := <-updateQueue:
			for floor := 0; floor < N_FLOORS; floor++ {
				for btn := 0; btn < N_BUTTONS; btn++ {
					elevator.Requests[floor][btn] = newRequests[floor][btn]
				}
			}
			// Set lights accordingly to this new queue
		case newFloor := <-drvFloors:
			Fsm_OnFloorArrival(&elevator, newFloor)
		case <-drvObstr:
			// Do stuff
		case <-drvStop:
			// Do stuff

			/*default:
			// Request button
			for floor := 0; floor < N_FLOORS; floor++ {
				for btn := 0; btn < N_BUTTONS; btn++ {
					value := Elevator_RequestButton(floor, ButtonType(btn))
					if value && value != prevOrder[floor][btn] {
						Fsm_OnRequestButtonPress(&elevator, floor, ButtonType(btn))
						// ADD A CHANNEL THAT SENDS NEW ORDER TO REQUEST QUEUE. MAYBE INSIDE Fsm_OnRequestButtonPress(&elevator, floor, ButtonType(btn))
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

			time.Sleep(time.Duration(inputPollRate_ms) * time.Millisecond)*/
		}
	}
}
