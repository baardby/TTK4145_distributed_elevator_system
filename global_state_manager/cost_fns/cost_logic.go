package costfns //Må endres hvis det puttes inn i en mappe

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
	. "distributed_elevator/global_state_manager"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
)

// func costFunction				//finn den beste heisen, kalles av elevator driver
//		input: order		//floor and btntype
//		regner ut beste heis
//		return: elevator with best cost

//func newButtonPress
//		input: order 		//floor and btntype
//		kaller costfunction med order
//		lager nytt element av type request
//		kaller newRequest fra modul request_queue
//		return: none

// func redistrubuteRequests 		//iterer alle request som ikke lenger har en heis, og tildel en heis
//		input: request_queue?
//		for every request with defect elevator
//			call calculate cost for
//		return: none

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

func assignNewOrder(newOrder ButtonEvent, elevatorStates ElevatorStates, myID int) (IDAssigned int) { // Needs to take in elevator states
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

		// Lag input til assigner
		input := makeHRAIInput()
		// Legg inn den ene ordren
		input.HallRequests[newOrder.Floor][int(newOrder.Button)] = true
		// Legg inn heisene som er i live
		// Legg inn de aktive CabOrdrene til hver, eller la det blir tilsendt
		for elevatorPeer := 0; elevatorPeer < N_ELEVATORS; elevatorPeer++ {
			if (elevatorPeer != myID) && (elevatorStates.Peers[elevatorPeer].Alive) {
				input.States
			}
		}

		// Send inn input
		jsonBytes, err := json.Marshal(input)
		if err != nil {
			fmt.Println("json.Marshal error: ", err)
			return
		}

		ret, err := exec.Command("hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
		if err != nil {
			fmt.Println("exec.Command error: ", err)
			fmt.Println(string(ret))
			return
		}

		// Hent ut output
		// Finn heisen som fikk ordren, altså hvem som har true
		output := new(map[string][][2]bool)
		err = json.Unmarshal(ret, &output)
		if err != nil {
			fmt.Println("json.Unmarshal error: ", err)
			return
		}

		fmt.Printf("output: \n")
		for k, v := range *output {
			fmt.Printf("%6v :  %+v\n", k, v)
		}

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

		// Lag input til assigner
		input := makeHRAIInput()
		// Legg inn den ene ordren
		input.HallRequests[newOrder.Floor][int(newOrder.Button)] = true
		// Legg inn heisene som er i live
		// Legg inn de aktive CabOrdrene til hver, eller la det blir tilsendt
		for elevatorPeer := 0; elevatorPeer < N_ELEVATORS; elevatorPeer++ {
			if (elevatorPeer != myID) && (elevatorStates.Peers[elevatorPeer].Alive) {
				input.States
			}
		}

		// Send inn input
		jsonBytes, err := json.Marshal(input)
		if err != nil {
			fmt.Println("json.Marshal error: ", err)
			return
		}

		ret, err := exec.Command("hall_request_assigner/"+hraExecutable, "-i", string(jsonBytes)).CombinedOutput()
		if err != nil {
			fmt.Println("exec.Command error: ", err)
			fmt.Println(string(ret))
			return
		}

		// Hent ut output
		// Finn heisen som fikk ordren, altså hvem som har true
		output := new(map[string][][2]bool)
		err = json.Unmarshal(ret, &output)
		if err != nil {
			fmt.Println("json.Unmarshal error: ", err)
			return
		}

		fmt.Printf("output: \n")
		for k, v := range *output {
			fmt.Printf("%6v :  %+v\n", k, v)
		}

	default:
		fmt.Println("Not a valid button")
	}
	return
}

func makeHRAIInput() HRAInput {
	return HRAInput{
		HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
		States: map[string]HRAElevState{
			"one": HRAElevState{
				Behavior:    "idle",
				Floor:       0,
				Direction:   "stop",
				CabRequests: []bool{false, false, false, false},
			},
		},
	}
}
