package elevalgo

import (
	. "distributed_elevator/elevio"
)

type DirnBehaviourPair struct {
	Dirn      MotorDirection
	Behaviour ElevatorBehaviour
}

// Tidligere h funskjoner
func Requests_ChooseDirection(e Elevator) DirnBehaviourPair {
	switch e.direction {
	case MD_Up:
		if requests_above(e) == true {
			return DirnBehaviourPair{Dirn: MD_Up, Behaviour: EB_Moving}
		} else if requests_here(e) == true {
			return DirnBehaviourPair{Dirn: MD_Down, Behaviour: EB_DoorOpen}
		} else if requests_below(e) == true {
			return DirnBehaviourPair{Dirn: MD_Down, Behaviour: EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: MD_Stop, Behaviour: EB_Idle}
		}
	case MD_Down:
		if requests_below(e) == true {
			return DirnBehaviourPair{Dirn: MD_Down, Behaviour: EB_Moving}
		} else if requests_here(e) == true {
			return DirnBehaviourPair{Dirn: MD_Up, Behaviour: EB_DoorOpen}
		} else if requests_above(e) == true {
			return DirnBehaviourPair{Dirn: MD_Up, Behaviour: EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: MD_Stop, Behaviour: EB_Idle}
		}

	case MD_Stop:
		if requests_here(e) == true {
			return DirnBehaviourPair{Dirn: MD_Stop, Behaviour: EB_DoorOpen}
		} else if requests_above(e) == true {
			return DirnBehaviourPair{Dirn: MD_Up, Behaviour: EB_Moving}
		} else if requests_below(e) == true {
			return DirnBehaviourPair{Dirn: MD_Down, Behaviour: EB_Moving}
		} else {
			return DirnBehaviourPair{Dirn: MD_Stop, Behaviour: EB_Idle}
		}
	default:
		return DirnBehaviourPair{Dirn: MD_Stop, Behaviour: EB_Idle}
	}

}

func Requests_ShouldStop(e Elevator) bool {
	switch e.direction {
	case MD_Down:
		return e.requests[e.floor][BT_HallDown] != false ||
			e.requests[e.floor][BT_Cab] != false ||
			requests_below(e) == false

	case MD_Up:
		return e.requests[e.floor][BT_HallUp] != false ||
			e.requests[e.floor][BT_Cab] != false ||
			requests_above(e) == false

	case MD_Stop:
	default:
		return true
	}
	return false
}

func Request_ShouldClearImmediately(e Elevator, btn_floor int, btn_type ButtonType) bool {
	return e.floor == btn_floor && ((e.direction == MD_Up && btn_type == BT_HallUp) ||
		(e.direction == MD_Down && btn_type == BT_HallDown) ||
		e.direction == MD_Stop ||
		btn_type == BT_Cab)
}

func Requests_ClearAtCurrentFloor(e Elevator) Elevator {
	e.requests[e.floor][BT_Cab] = false
	switch e.direction {
	case MD_Up:
		if requests_above(e) == false && e.requests[e.floor][BT_HallUp] == false {
			e.requests[e.floor][BT_HallDown] = false
		}
		e.requests[e.floor][BT_HallUp] = false
	case MD_Down:
		if requests_below(e) == false && e.requests[e.floor][BT_HallDown] == false {
			e.requests[e.floor][BT_HallUp] = false
		}
		e.requests[e.floor][BT_HallDown] = false
	case MD_Stop:
	default:
		e.requests[e.floor][BT_HallUp] = false
		e.requests[e.floor][BT_HallDown] = false
	}
	return e
}

//Filer som var i .c fil

func requests_above(e Elevator) bool {
	for f := e.floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.requests[f][btn] != false {
				return true
			}
		}
	}
	return false
}

func requests_below(e Elevator) bool {
	for f := 0; f < e.floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.requests[f][btn] == true {
				return true
			}
		}
	}
	return false
}

func requests_here(e Elevator) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.requests[e.floor][btn] != false {
			return true
		}
	}
	return false
}
