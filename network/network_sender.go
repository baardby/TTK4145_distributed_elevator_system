package network

import (
	. "distributed_elevator/elevalgo"
	"log"
	"net"
)

type NetworkSender struct {
	DestID   int // DO WE NEED THIS?
	DestIP   string
	DestPort string
	DestAddr *net.UDPAddr
	MyConn   *net.UDPConn // Remember to add defer myConn.Close() in the loop the sender is run
}

func (sender *NetworkSender) Network_SenderInit() {
	var err error

	sender.DestPort = "20003"
	sender.DestIP = "255.255.255.255"

	sender.DestAddr, err = net.ResolveUDPAddr("udp4", sender.DestIP+":"+sender.DestPort)
	if err != nil { // ADD ERROR HANDLING
		log.Fatalf("Could not resolve address: %v", err)
	}

	sender.MyConn, err = net.ListenUDP("udp4", nil)
	if err != nil { // ADD ERROR HANDLING
		log.Fatalf("Error dialing: %v", err)
	}
}

func (sender *NetworkSender) BroadcastOnNetwork(me Elevator, requestStates [N_FLOORS][N_BUTTONS]byte, requestAssignedTo [N_FLOORS][N_BUTTONS]byte) error {
	_, err := sender.MyConn.WriteToUDP(constructMessageToSlice(me, requestStates, requestAssignedTo), sender.DestAddr)
	if err != nil { // ADD ERROR HANDLING
		log.Fatalf("Sending message error: %v", err)
	}

	return err
}

func constructMessageToSlice(me Elevator, requestStates [N_FLOORS][N_BUTTONS]byte, requestAssignedTo [N_FLOORS][N_BUTTONS]byte) []byte {
	buffer := make([]byte, 1024)
	// DO STUFF
	return buffer
}

//func setUpSocketSender			//skal være go-routine
//		input: none
//		set up socket
//		return: none

//func broadcastInfo				//go-routine?
//		input: message
//		Use socket to broadcast info
//		return: none
