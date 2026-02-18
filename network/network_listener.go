package network

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
	. "distributed_elevator/network/localip"
	"log"
	"net"
)

type NetworkListener struct {
	MyPort string
	MyIP   string
	MyConn *net.UDPConn // Remember to add defer myConn.Close() in the loop the listener is run
}

type ElevatorPeer struct { // MOVE TO ELEVATOR STATES
	Floor     int
	Direction MotorDirection
	Behaviour ElevatorBehaviour
}

// MAYBE ADD A TABLE OF PEERS AND THEIR IPs AND IDs

func (listener *NetworkListener) NetworkListenerInit() {
	var err error
	var myAddr *net.UDPAddr

	listener.MyPort = "20003"
	listener.MyIP, err = LocalIP()

	myAddr, err = net.ResolveUDPAddr("udp4", listener.MyIP+":"+listener.MyPort)
	if err != nil { // ADD ERROR HANDLING
		log.Fatalf("Failed to bind UDP socket %v", err)
	}

	listener.MyConn, err = net.ListenUDP("udp4", myAddr)
	// ADD ERROR HANDLING
}

func (listener *NetworkListener) ReadFromNetwork() (peer ElevatorPeer, requestStates [N_FLOORS][N_BUTTONS]byte, requestAssignedTo [N_FLOORS][N_BUTTONS]byte) { // MAYBE RETURN ID OF ELEVATOR PEER ALSO
	var recvAddr *net.UDPAddr
	var err error
	var msgSize int
	msgBuffer := make([]byte, 1024)

	msgSize, recvAddr, err = listener.MyConn.ReadFromUDP(msgBuffer)
	if err != nil { // ADD ERROR HANDLING
		log.Fatalf("Message error: %v", err)
	}

	// LOGIC THAT REGISTERS WHO SENT IT WITH recvAddr

	return reconstructMessageFromSlice(msgBuffer, msgSize)
}

func reconstructMessageFromSlice(msgBuffer []byte, msgSize int) (peer ElevatorPeer, requestStates [N_FLOORS][N_BUTTONS]byte, requestAssignedTo [N_FLOORS][N_BUTTONS]byte) {
	// DO STUFF
	return
}

//func setUpSocketListnere			//Skal være en go routine
//		Input: none
//		set up socket
//		return: Socket?

//func sendInfoToModule 			//nødvendig?
//		Input: buffer/message
//		Forward info
//		return: none

//func informNoConnection
//		input: none
//		send info to supervisor
//		return: none
