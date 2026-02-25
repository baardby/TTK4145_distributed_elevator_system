package main //Må endres hvis det puttes inn i en mappe

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/supervisor"
)

type RequestState int

const (
	None        RequestState = iota // Request completed by all
	Unconfirmed                     // Request confirmed by at least one
	Confirmed                       // Request confirmed by all
	Completed                       // Request completed by at least one
)

type RequestQueue struct {
	HallRequestStates      [N_FLOORS][2]RequestState
	CabRequestStates       [N_FLOORS][N_ELEVATORS]RequestState
	HallRequestsAssignedTo [N_FLOORS][2]string // Some kind of ID (the IP) for each elevator
	CabRequestsAssignedTo  [N_FLOORS][N_ELEVATORS]string
}

func initRequestQueue() (rq RequestQueue) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < 2; btn++ {
			rq.HallRequestsAssignedTo[floor][btn] = ""
		}
		for elevator := 0; elevator < N_ELEVATORS; elevator++ {
			rq.CabRequestsAssignedTo[floor][elevator] = ""
		}
	}

	return
}

//func restoreQueueAfterDisconnect 		//For å fikse når man kommer tilbake på nettet
//		input: None
//		Listen for other queue
//		Figure out if two different queues are recieved.
//		if riecieved:
//			Adopt
//		else:
//			init_queue()
//		return: None

//func broadcastQueue					//Fortelle andre om køen min
//		input: none
//		call broadcastInfo with queue as message
//		return: none

//func QueueUnion						//Motatt kø fra andre, må fikse unionen
//		input:QueueFromElevator#
//		figure out logic with Union
//		Barriers?
//		Update requests_queue
//		return: none

//func newRequest
//		input: request
//		legg til å køen
//		return: none
// NEED CHANNEL FOR NEW ORDERS

// Elevator controller:

// Sjekk elevator_controller for info om struktur av channels
// Skal sende cab og hall som nye bool matriser på updateQueueCh
// Skal bare sende cab og hall orders som er assignedTo denne heisen

// Skal også motta nye knappetrykk fra elevator controller

// Sender and listener:

// Select example in network sender
// Use ticker to periodically send to to network sender

// Standard case with channel to read from listener
