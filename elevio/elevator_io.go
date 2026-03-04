package elevio

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const N_FLOORS = 4
const N_BUTTONS = 3
const N_ELEVATORS = 3

const _pollRate = 20 * time.Millisecond

var _initialized bool = false
var _mtx sync.Mutex
var _conn net.Conn

type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down MotorDirection = -1
	MD_Stop MotorDirection = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown ButtonType = 1
	BT_Cab      ButtonType = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

func Init(addr string, numFloors int) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}

func SetMotorDirection(dir MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button ButtonType, floor int, value bool) {
	write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}

func PollButtons(buttonPressEvent chan<- ButtonEvent) {
	var prevButtonPress [N_FLOORS][N_BUTTONS]bool
	for {
		time.Sleep(_pollRate)
		for floor := 0; floor < N_FLOORS; floor++ {
			for btn := ButtonType(0); btn < N_BUTTONS; btn++ {
				newButtonPress := GetButton(btn, floor)
				if newButtonPress != prevButtonPress[floor][btn] && newButtonPress != false {
					buttonPressEvent <- ButtonEvent{floor, ButtonType(btn)}
				}
				prevButtonPress[floor][btn] = newButtonPress
			}
		}
	}
}

func PollFloorSensor(newFloorEvent chan<- int) {
	prevFloor := -1
	for {
		time.Sleep(_pollRate)
		currentFloor := GetFloor()
		if currentFloor != prevFloor && currentFloor != -1 {
			newFloorEvent <- currentFloor
		}
		prevFloor = currentFloor
	}
}

func PollStopButton(stopEvent chan<- bool) {
	prevStopState := false
	for {
		time.Sleep(_pollRate)
		currentStopState := GetStop()
		if currentStopState != prevStopState {
			stopEvent <- currentStopState
		}
		prevStopState = currentStopState
	}
}

func PollObstructionSwitch(obstrEvent chan<- bool) {
	prevObstrState := false
	for {
		time.Sleep(_pollRate)
		currentObstrState := GetObstruction()
		if currentObstrState != prevObstrState {
			obstrEvent <- currentObstrState
		}
		prevObstrState = currentObstrState
	}
}

func GetButton(button ButtonType, floor int) bool {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

func GetFloor() int {
	a := read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

func GetStop() bool {
	a := read([4]byte{8, 0, 0, 0})
	return toBool(a[1])
}

func GetObstruction() bool {
	a := read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}

func read(in [4]byte) [4]byte {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

func write(in [4]byte) {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
