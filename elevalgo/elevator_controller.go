package elevalgo //Må endres hvis det puttes inn i en mappe

import (
	. "distributed_elevator/elevio"
	"fmt"
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
			fmt.Println(newButton.Floor)
			fmt.Println(int(newButton.Button))
			Fsm_OnRequestButtonPress(&elevator, newButton.Floor, newButton.Button)
		case <-stopEvent:
			// Do stuff
		case <-obstrEvent:
			// Do stuff
		default:
			if Timer_TimedOut() {
				Timer_Stop()
				Fsm_OnDoorTimeout(&elevator)
			}
		}
	}
}
