package elevio

import (
	"time"
)

const N_FLOORS = 4
const N_BUTTONS = 3

func Elevio_PollIO(drvButtons chan<- ButtonEvent,
	drvFloors chan<- int,
	drvObstr chan<- bool,
	drvStop chan<- bool) {

	Init("localhost:15657", N_FLOORS)

	prevFloor := -1
	prevStopBtnPress := false
	prevObstruction := false
	var prevBtnPress [N_FLOORS][N_BUTTONS]bool

	for {
		// Poll buttons
		time.Sleep(_pollRate)
		for floor := 0; floor < N_FLOORS; floor++ {
			for btn := 0; btn < N_BUTTONS; btn++ {
				buttonPressed := GetButton(ButtonType(btn), floor)
				if buttonPressed && buttonPressed != prevBtnPress[floor][btn] {
					drvButtons <- ButtonEvent{floor, ButtonType(btn)}
				}
				prevBtnPress[floor][btn] = buttonPressed
			}
		}

		// Poll floor sensor
		time.Sleep(_pollRate)
		currentFloor := GetFloor()
		if currentFloor != prevFloor && currentFloor != -1 {
			drvFloors <- currentFloor
		}
		prevFloor = currentFloor

		// Poll stopbutton
		time.Sleep(_pollRate)
		currStopBtnPress := GetStop()
		if currStopBtnPress != prevStopBtnPress {
			drvStop <- currStopBtnPress
		}
		prevStopBtnPress = currStopBtnPress

		// Poll obstruction sensor
		time.Sleep(_pollRate)
		currObstruction := GetObstruction()
		if currObstruction != prevObstruction {
			drvObstr <- currObstruction
		}
		prevObstruction = currObstruction
	}
}
