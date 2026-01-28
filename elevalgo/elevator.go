package elevalgo

import (
	. "distributed_elevator/elevio"
	"fmt"
)

const N_FLOORS = 4
const N_BUTTONS = 3

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type Config struct {
	DoorOpenDuration_s float64
}

type Elevator struct {
	Floor     int
	Direction MotorDirection
	Requests  [N_FLOORS][N_BUTTONS]bool
	Behaviour ElevatorBehaviour
	Config    Config
}

func Elevator_BehaviourToString(eb ElevatorBehaviour) string {
	switch eb {
	case EB_Idle:
		return "EB_Idle"
	case EB_DoorOpen:
		return "EB_DoorOpen"
	case EB_Moving:
		return "EB_Moving"
	default:
		return "Udefined elevator behaviour"
	}
}

func Elevator_MotorDirectionToString(dirn MotorDirection) string {
	switch dirn {
	case MD_Down:
		return "MD_Down"
	case MD_Stop:
		return "MD_Stop"
	case MD_Up:
		return "MD_Up"
	default:
		return "Undefined direction"
	}
}

func Elevator_ButtonToString(btn ButtonType) string {
	switch btn {
	case BT_HallUp:
		return "B_HallUp"
	case BT_HallDown:
		return "B_HallDown"
	case BT_Cab:
		return "B_Cab"
	default:
		return "Udefined button"
	}
}

func (elevator *Elevator) PrintState() {
	fmt.Printf(" +--------------------+\n")
	fmt.Printf("  |floor = %-2d          |\n", elevator.Floor)
	fmt.Printf("  |dirn  = %-12.12s|\n", Elevator_MotorDirectionToString(elevator.Direction))
	fmt.Printf("  |behav = %-12.12s|\n", Elevator_BehaviourToString(elevator.Behaviour))
	fmt.Printf(" +--------------------+\n")
	fmt.Printf("  |  | up  | dn  | cab |\n")
	for floor := N_FLOORS - 1; floor >= 0; floor-- {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if floor == N_FLOORS-1 && btn == int(BT_HallUp) || floor == 0 && btn == int(BT_HallDown) {
				fmt.Print("|	")
			} else {
				if elevator.Requests[floor][btn] {
					fmt.Print("| # ")
				} else {
					fmt.Print("| - ")
				}
			}
		}
		fmt.Println("|")
	}
	fmt.Println("  +--------------------+")
}

func Elevator_Uninitialized() Elevator {
	Init("localhost:15657", N_FLOORS)
	return Elevator{
		Floor:     -1,
		Direction: MD_Stop,
		Behaviour: EB_Idle,
		Config: Config{
			DoorOpenDuration_s: 3.0,
		},
	}
}

func Elevator_FloorSensor() int {
	return GetFloor()
}

func Elevator_RequestButton(floor int, btn ButtonType) bool {
	return GetButton(btn, floor)
}

func Elevator_StopButton() bool {
	return GetStop()
}

func Elevator_Obstruction() bool {
	return GetObstruction()
}

func Elevator_FloorIndicator(floor int) {
	SetFloorIndicator(floor)
}

func Elevator_DoorLight(value bool) {
	SetDoorOpenLamp(value)
}

func Elevator_StopButtonLight(value bool) {
	SetStopLamp(value)
}

func Elevator_MotorDirection(dirn MotorDirection) {
	SetMotorDirection(dirn)
}

func Elevator_RequestButtonLight(floor int, btn ButtonType, value bool) {
	SetButtonLamp(btn, floor, value)
}
