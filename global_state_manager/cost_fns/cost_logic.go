package cost_fns //Må endres hvis det puttes inn i en mappe

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
	. "distributed_elevator/global_state_manager/elevator_states"
	. "distributed_elevator/global_state_manager/order_queue"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
)

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func AssignNewOrder(newOrder ButtonEvent, elevatorStates ElevatorStates, cabOrders map[int]AllCabOrders, myID int) (IDAssigned int) { // Needs to take in elevator states
	switch newOrder.Button {
	// If it is a CAB order, this elevator should do it.
	case BT_Cab:
		IDAssigned = myID

	case BT_HallDown:
		hraExecutable := ""
		switch runtime.GOOS {
		case "linux":
			hraExecutable = "hall_request_assigner"
		case "windows":
			hraExecutable = "hall_request_assigner.exe"
		default:
			panic("OS not supported")
		}

		input := makeHRAIInput()

		input.HallRequests[newOrder.Floor][int(newOrder.Button)] = true

		for elevatorPeer := 0; elevatorPeer < N_ELEVATORS; elevatorPeer++ {
			currentPeerID := elevatorPeer + 1
			if elevatorStates.Peers[elevatorPeer].Alive {
				input.States[IDToString(currentPeerID)] = HRAElevState{
					Behavior:    Elevator_BehaviourToString(elevatorStates.Peers[elevatorPeer].Behaviour),
					Floor:       elevatorStates.Peers[elevatorPeer].Floor,
					Direction:   Elevator_MotorDirectionToString(elevatorStates.Peers[elevatorPeer].Direction),
					CabRequests: extractCabOrder(currentPeerID, cabOrders),
				}
			}
		}

		// Encode the input to json to be sent to executable
		jsonBytes, err := json.Marshal(input)
		if err != nil {
			fmt.Println("json.Marshal error: ", err)
			return
		}
		// Start the hall_request_assigner executable
		ret, err := exec.Command("hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
		if err != nil {
			fmt.Println("exec.Command error: ", err)
			fmt.Println(string(ret))
			return
		}

		// Decode output from executable
		output := new(map[string][][2]bool)
		err = json.Unmarshal(ret, &output)
		if err != nil {
			fmt.Println("json.Unmarshal error: ", err)
			return
		}

		// Find which elevator that was assigned the order
		for string_ID, assignedHallRequests := range *output {
			if assignedHallRequests[newOrder.Floor][int(newOrder.Button)] {
				IDAssigned = IDToInt(string_ID)
			}
		}
		return

	case BT_HallUp:
		hraExecutable := ""
		switch runtime.GOOS {
		case "linux":
			hraExecutable = "hall_request_assigner"
		case "windows":
			hraExecutable = "hall_request_assigner.exe"
		default:
			panic("OS not supported")
		}

		input := makeHRAIInput()

		input.HallRequests[newOrder.Floor][int(newOrder.Button)] = true

		for elevatorPeer := 0; elevatorPeer < N_ELEVATORS; elevatorPeer++ {
			currentPeerID := elevatorPeer + 1
			if elevatorStates.Peers[elevatorPeer].Alive {
				input.States[IDToString(currentPeerID)] = HRAElevState{
					Behavior:    Elevator_BehaviourToString(elevatorStates.Peers[elevatorPeer].Behaviour),
					Floor:       elevatorStates.Peers[elevatorPeer].Floor,
					Direction:   Elevator_MotorDirectionToString(elevatorStates.Peers[elevatorPeer].Direction),
					CabRequests: extractCabOrder(currentPeerID, cabOrders),
				}
			}
		}

		// Encode the input to json to be sent to executable
		jsonBytes, err := json.Marshal(input)
		if err != nil {
			fmt.Println("json.Marshal error: ", err)
			return
		}
		// Start the hall_request_assigner executable
		ret, err := exec.Command("hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
		if err != nil {
			fmt.Println("exec.Command error: ", err)
			fmt.Println(string(ret))
			return
		}

		// Decode output from executable
		output := new(map[string][][2]bool)
		err = json.Unmarshal(ret, &output)
		if err != nil {
			fmt.Println("json.Unmarshal error: ", err)
			return
		}

		// Find which elevator that was assigned the order
		for string_ID, assignedHallRequests := range *output {
			if assignedHallRequests[newOrder.Floor][int(newOrder.Button)] {
				IDAssigned = IDToInt(string_ID)
			}
		}
		return

	default:
		fmt.Println("Not a valid button")
	}
	return
}

func IDToString(ID int) string {
	switch ID {
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	case 4:
		return "four"
	case 5:
		return "five"
	default:
		return "none"
	}
}

func IDToInt(ID string) int {
	switch ID {
	case "one":
		return 1
	case "two":
		return 2
	case "three":
		return 3
	case "four":
		return 4
	case "five":
		return 5
	default:
		return 0
	}
}

func makeHRAIInput() HRAInput {
	return HRAInput{
		HallRequests: [][2]bool{{false, false}, {false, false}, {false, false}, {false, false}},
		// {{BT_HallUp, BT_HallDown}, {BT_HallUp, BT_HallDown}, ...}
		States: map[string]HRAElevState{},
	}
}

func extractCabOrder(elevatorID int, cabOrders map[int]AllCabOrders) []bool {
	cabRequests := make([]bool, N_FLOORS)
	for floor := 0; floor < N_FLOORS; floor++ {
		cabRequests[floor] = (cabOrders[elevatorID][floor].State == Confirmed)
	}
	return cabRequests
}

func testCostLogic() {
	elevatorStates := ElevatorStates{
		Peers: [N_ELEVATORS]ElevatorPeer{
			{Alive: true, Floor: 1, Behaviour: ElevatorBehaviour(0), Direction: MotorDirection(0)},
			{Alive: false, Floor: 2, Behaviour: ElevatorBehaviour(1), Direction: MotorDirection(0)},
			{Alive: true, Floor: 0, Behaviour: ElevatorBehaviour(1), Direction: MotorDirection(0)},
		},
	}

	cabOrders := map[int]AllCabOrders{
		1: {{State: None, AssignedTo: 1}, {State: None, AssignedTo: 1}, {State: None, AssignedTo: 1}, {State: Unconfirmed, AssignedTo: 1}},
		2: {{State: None, AssignedTo: 1}, {State: None, AssignedTo: 1}, {State: Completed, AssignedTo: 1}, {State: None, AssignedTo: 1}},
		3: {{State: None, AssignedTo: 1}, {State: Confirmed, AssignedTo: 1}, {State: None, AssignedTo: 1}, {State: None, AssignedTo: 1}},
	}

	myId := 1

	newButtonEvent := ButtonEvent{
		Floor:  1,
		Button: ButtonType(1),
	}

	fmt.Println(AssignNewOrder(newButtonEvent, elevatorStates, cabOrders, myId))

	newButtonEvent = ButtonEvent{
		Floor:  3,
		Button: ButtonType(2),
	}
	fmt.Println(AssignNewOrder(newButtonEvent, elevatorStates, cabOrders, myId))
}
