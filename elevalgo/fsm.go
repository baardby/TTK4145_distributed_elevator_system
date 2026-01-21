package elevalgo

import (
	"fmt"
	."distributed_elevator/elevio"
)

func fsm_setAllLights(es Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			Elevator_RequestButtonLight(floor, ButtonType(btn), es.requests[floor][btn])
		}
	}
}

func Fsm_OnInitBetweenFloors(e *Elevator) {
	Elevator_MotorDirection(MD_Down)
	e.direction = MD_Down
	e.behaviour = EB_Moving
}

func Fsm_OnRequestButtonPress(e *Elevator, btn_floor int, btn_type ButtonType) {
	fmt.Printf("\n\nFsm_OnRequestButtonPress(%d, %s)\n", btn_floor, Elevator_ButtonToString(btn_type))
	e.PrintState()

	switch e.behaviour {
	case EB_DoorOpen:
		if Requests_ShouldClearImmediately(*e, btn_floor, btn_type) {
			Timer_Start(e.config.doorOpenDuration_s)
		} else {
			e.requests[btn_floor][btn_type] = true
		}
		break

	case EB_Moving:
		e.requests[btn_floor][btn_type] = true
		break

	case EB_Idle:
		e.requests[btn_floor][btn_type] = true
		pair := Requests_ChooseDirection(*e)
		e.direction = pair.dirn
		e.behaviour = pair.behaviour
		switch pair.behaviour {
		case EB_DoorOpen:
			Elevator_DoorLight(true)
			Timer_Start(e.config.doorOpenDuration_s)
			*e = Requests_ClearAtCurrentFloor(*e)
			break

		case EB_Moving:
			Elevator_MotorDirection(e.direction)
			break

		case EB_Idle:
			break
		}
		break
	}

	fsm_setAllLights(*e)

	fmt.Println("\nNew state:")
	e.PrintState()
}

func Fsm_OnFloorArrival(e *Elevator, newFloor int) {
	fmt.Printf("\n\nFsm_OnFloorArrival(%d)\n", newFloor)
	e.PrintState()

	e.floor = newFloor

	Elevator_FloorIndicator(e.floor)

	switch e.behaviour {
	case EB_Moving:
		if Requests_ShouldStop(*e) {
			Elevator_MotorDirection(MD_Stop)
			Elevator_DoorLight(1)
			*e = Requests_ClearAtCurrentFloor(*e)
			Timer_Start(e.config.doorOpenDuration_s)
			fsm_setAllLights(*e)
			e.behaviour = EB_DoorOpen
		}
		break
	default:
		break
	}

	fmt.Println("\nNew state:")
	e.PrintState()
}

func Fsm_OnDoorTimeout(e *Elevator) {
	fmt.Printf("\n\nFsm_OnDoorTimeout()\n")
	e.PrintState()

	switch e.behaviour {
	case EB_DoorOpen:
		pair := Requests_ChooseDirection(*e)
		e.direction = pair.dirn
		e.behaviour = pair.behaviour

		switch e.behaviour {
		case EB_DoorOpen:
			Timer_Start(e.config.doorOpenDuration_s)
			*e = Requests_ClearAtCurrentFloor(*e)
			fsm_setAllLights(*e)
			break
		case EB_Moving:
		case EB_Idle:
			Elevator_DoorLight(false)
			Elevator_MotorDirection(e.direction)
			break
		}

		break
	default:
		break
	}

	fmt.Println("\nNew state:")
	e.PrintState()
}