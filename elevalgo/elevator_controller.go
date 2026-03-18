package elevalgo //Må endres hvis det puttes inn i en mappe

import (
	. "distributed_elevator/elevio"
	"fmt"
	"time"
)

func ElevatorController(updateMyQueue <-chan [N_FLOORS][N_BUTTONS]bool,
	newFloorEvent <-chan int,
	stopEvent <-chan bool,
	obstrEvent <-chan bool,
	buttonPressEvent <-chan ButtonEvent, // FOR TESTING WITH SINGLE ELEVATOR. TODO: Remove after testing
	stateToGSM chan Elevator,
	stateToSupervisor chan Elevator) {

	var elevator Elevator = Elevator_Uninitialized()

	select {
	// If a new floor can be received, the elevator is not between floors. Do nothing
	case startFloor := <-newFloorEvent:
		Fsm_OnFloorArrival(&elevator, startFloor)

	// If a new floor can't be received, the elevator is between floors. Go down
	default:
		Fsm_OnInitBetweenFloors(&elevator)
		startFloor := <-newFloorEvent
		Fsm_OnFloorArrival(&elevator, startFloor)
	}

	updateElevatorTicker := time.NewTicker(100 * time.Millisecond)
	defer updateElevatorTicker.Stop()

	for {
		select {
		case newRequests := <-updateMyQueue:
			for floor := 0; floor < N_FLOORS; floor++ {
				for btn := 0; btn < N_BUTTONS; btn++ {
					// Treat new request as button press if it is a new request.
					// Otherwise, just update the request state in the elevator struct
					if (elevator.Requests[floor][btn] != newRequests[floor][btn]) && newRequests[floor][btn] {
						Fsm_OnRequestButtonPress(&elevator, floor, ButtonType(btn))
					} else {
						elevator.Requests[floor][btn] = newRequests[floor][btn]
					}
				}
			}

		case newFloor := <-newFloorEvent:
			Fsm_OnFloorArrival(&elevator, newFloor)

		//case newButton := <-buttonPressEvent: // FOR TESTING WITH SINGLE ELEVATOR. TODO: Remove after testing
		//	Fsm_OnRequestButtonPress(&elevator, newButton.Floor, newButton.Button)

		case stopButtonState := <-stopEvent: // TODO: Double check if we need to implement something more complex for the stop button
			SetStopLamp(stopButtonState)

		case currentObstrState := <-obstrEvent:
			elevator.SetObstr(currentObstrState)

			// When obstruction disappears, restart doorOpenTimer
			if !currentObstrState {
				Timer_Start(elevator.Config.DoorOpenDuration_s)
			}

		case <-updateElevatorTicker.C:
			select {
			// Try send new update
			case stateToGSM <- elevator:

			// Empty the channel if the old message wasn't received and send new update
			default:
				<-stateToGSM
				stateToGSM <- elevator
			}

			select {
			// Try send new update
			case stateToSupervisor <- elevator:

			// Empty the channel if the old message wasn't received and send new update
			default:
				<-stateToSupervisor
				stateToSupervisor <- elevator
			}

			// TODO: Remove this after testing
			for k, v := range elevator.Requests {
				fmt.Printf("%6v :  %+v\n", k, v)
			}
			// END OF TODO

		default:
			if Timer_TimedOut() {
				Timer_Stop()
				Fsm_OnDoorTimeout(&elevator)
			}
		}
	}
}
