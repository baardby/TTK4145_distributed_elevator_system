package elevalgo //Må endres hvis det puttes inn i en mappe

import (
	. "distributed_elevator/elevio"
	"time"
)

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
			for i := 0; i < N_FLOORS; i++ {
				for j := 0; j < N_BUTTONS; j++ {
					elevator.Requests[i][j] = newRequests[i][j]
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