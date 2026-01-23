package elevalgo

import (
	"fmt"
)

type DirnBehaviourPair struct {
	Dirn      Dirn
	Behaviour ElevatorBehaviour
}

// Tidligere h funskjoner
func Requests_ChooseDirection(e Elevator) DirnBehaviourPair {
	switch e.Dirn {
	case D_Up:
		if requests_above(e) == 1 {
			return DirnBehaviourPair{Dirn: D_Up, Behaviour: EB_Moving}
		} else if requests_here(e) == 1 {
			return DirnBehaviourPair{Dirn: D_Down, Behaviour: EB_DoorOpen}
		} else if requests_below(e) == 1 {
			return DirnBehaviourPair{Dirn: D_Down, Behaviour: EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: D_Stop, Behaviour: EB_Idle}
		}
	case D_Down:
		if requests_below(e) == 1 {
			return DirnBehaviourPair{Dirn: D_Down, Behaviour: EB_Moving}
		} else if requests_here(e) == 1 {
			return DirnBehaviourPair{Dirn: D_Up, Behaviour: EB_DoorOpen}
		} else if requests_above(e) == 1 {
			return DirnBehaviourPair{Dirn: D_Up, Behaviour: EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: D_Stop, Behaviour: EB_Idle}
		}

	case D_Stop:
		if requests_here(e) == 1 {
			return DirnBehaviourPair{Dirn: D_Stop, Behaviour: EB_DoorOpen}
		} else if requests_above(e) == 1 {
			return DirnBehaviourPair{Dirn: D_Up, Behaviour: EB_Moving}
		} else if requests_below(e) == 1 {
			return DirnBehaviourPair{Dirn: D_Down, Behaviour: EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: D_Stop, Behaviour: EB_Idle}
		}
	default:
		return DirnBehaviourPair{Dirn: D_Stop, Behaviour: EB_Idle}
	}

}

func Requests_ShouldStop(e Elevator) bool {
	switch e.Dirn {
	case D_Down:
		return e.Requests[e.Floor][B_HallDown] != 0 ||
			e.Requests[e.Floor][B_Cab] != 0 ||
			requests_below(e) == 0

	case D_Up:
		return e.Requests[e.Floor][B_HallUp] != 0 ||
			e.Requests[e.Floor][B_Cab] != 0 ||
			requests_above(e) == 0

	case D_Stop:
	default:
		return true
	}
	return false
}

func Request_ShouldClearImmediately(e Elevator, btn_floor int, btn_type Button) bool {
	return e.Floor == btn_floor && ((e.Dirn == D_Up && btn_type == B_HallUp) ||
		(e.Dirn == D_Down && btn_type == B_HallDown) ||
		e.Dirn == D_Stop ||
		btn_type == B_Cab)
}

func Requests_ClearAtCurrentFloor(e Elevator) Elevator {
	e.Requests[e.Floor][B_Cab] = 0
	switch e.Dirn {
	case D_Up:
		if requests_above(e) == 0 && e.Requests[e.Floor][B_HallUp] == 0 {
			e.Requests[e.Floor][B_HallDown] = 0
		}
		e.Requests[e.Floor][B_HallUp] = 0
	case D_Down:
		if requests_below(e) == 0 && e.Requests[e.Floor][B_HallDown] == 0 {
			e.Requests[e.Floor][B_HallUp] = 0
		}
		e.Requests[e.Floor][B_HallDown] = 0
	case D_Stop:
	default:
		e.Requests[e.Floor][B_HallUp] = 0
		e.Requests[e.Floor][B_HallDown] = 0
	}
	return e
}

//Filer som var i .c fil

func requests_above(e Elevator) int {
	for f := e.Floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Requests[f][btn] != 0 {
				return 1
			}
		}
	}
	return 0
}

func requests_below(e Elevator) int {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.Requests[f][btn] == 1 {
				return 1
			}
		}
	}
	return 0
}

func requests_here(e Elevator) int {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] != 0 {
			return 1
		}
	}
	return 0
}
