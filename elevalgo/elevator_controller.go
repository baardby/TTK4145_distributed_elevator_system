package elevalgo //Må endres hvis det puttes inn i en mappe

import (
	. "distributed_elevator/elevio"
)

func Elevalgo_ElevatorControllerLoop(updateQueueEvent <-chan [N_FLOORS][N_BUTTONS]bool,
	newFloorEvent <-chan int,
	stopEvent <-chan bool,
	obstrEvent <-chan bool,
	buttonPressEvent <-chan ButtonEvent) {

	var elevator Elevator = Elevator_Uninitialized()

	//startFloor := <- newFloorEvent
	//if startFloor == -1 {
	//	Fsm_OnInitBetweenFloors(&elevator)
	//}

	for {
		select {
		case newRequests := <-updateQueueEvent:
			for floor := 0; floor < N_FLOORS; floor++ {
				for btn := 0; btn < N_BUTTONS; btn++ {
					if elevator.Requests[floor][btn] != newRequests[floor][btn] && newRequests[floor][btn] {
						Fsm_OnRequestButtonPress(&elevator, floor, ButtonType(btn))
					}
					elevator.Requests[floor][btn] = newRequests[floor][btn]
				}
			}
			// Set lights accordingly to this new queue
		case newFloor := <-newFloorEvent:
			Fsm_OnFloorArrival(&elevator, newFloor)
		case newButton := <-buttonPressEvent:
			Fsm_OnRequestButtonPress(&elevator, newButton.Floor, newButton.Button)
		case <-stopEvent:
			// Maybe occupy the elevator completely
		case currentObstrState := <-obstrEvent:
			elevator.SetObstr(currentObstrState)
		default:
			if Timer_TimedOut() {
				Timer_Stop()
				Fsm_OnDoorTimeout(&elevator)
			}
		}
	}
}

// Elevio and Elevalgo as one goroutine REMOVE ONLY IF WE KNOW ABOVE WORKS
/*
func Elevalgo_ElevatorControllerLoop(updateQueueCh <-chan [N_FLOORS][N_BUTTONS]bool, newButtonPress chan<- [2]int) {
	var elevator Elevator = Elevator_Uninitialized()
	var inputPollRate_ms int = 25

	// con_load

	if Elevator_FloorSensor() == -1 {
		Fsm_OnInitBetweenFloors(&elevator)
	}
	prevFloor := -1
	var prevOrder [N_FLOORS][N_BUTTONS]bool

	for {
		select {
		case newRequests := <-updateQueueCh:
			for floor := 0; floor < N_FLOORS; floor++ {
				for btn := 0; btn < N_BUTTONS; btn++ {
					elevator.Requests[floor][btn] = newRequests[floor][btn]
				}
			}
		default:
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

			time.Sleep(time.Duration(inputPollRate_ms) * time.Millisecond)
		}
	}
}
*/
