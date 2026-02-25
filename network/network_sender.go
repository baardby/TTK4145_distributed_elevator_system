package network

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
	. "distributed_elevator/supervisor"
	. "distributed_elevator/request_queue"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

type NetworkSender struct {
	DestID       int // DO WE NEED THIS?
	DestIP       string
	DestPort     string
	DestAddr     *net.UDPAddr
	MyConn       *net.UDPConn // Remember to add defer myConn.Close() in the loop the sender is run
	Myself       Elevator
	Requestqueue RequestQueue
}

func (sender *NetworkSender) networkSenderInit() {
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

func (sender *NetworkSender) broadcastOnNetwork(myself Elevator, msg Message) error {
	_, err := sender.MyConn.WriteToUDP(constructMessageToSlice(myself, msg), sender.DestAddr)
	if err != nil { // ADD ERROR HANDLING
		log.Fatalf("Sending message error: %v", err)
	}

	return err
}

func constructMessageToSlice(myself Elevator, msg Message) []byte {
	msg.Peer.Floor = myself.Floor
	msg.Peer.Behaviour = myself.Behaviour
	msg.Peer.Direction = myself.Direction
	msg.Peer.Alive = true

	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("marshal error:", err)
	}

	return data
}

func Network_SenderFSM(elevatorStateCh <-chan Elevator, requestQueueCh <-chan RequestQueue) {
	var sender NetworkSender
	sender.networkSenderInit()
	defer sender.MyConn.Close()

	// Setting up periodic sending
	ticker := time.NewTicker(1000 * time.Millisecond) // CHANGE TO CORRECT TIME 50Hz?
	defer ticker.Stop()

	var msgToSend Message = Message{
		Peer: ElevatorPeer{
			Floor:     -1,
			Direction: MD_Stop,
			Behaviour: EB_Idle,
			Alive:     false,
		},
	}
	var elevator Elevator = Elevator{
		Floor:     2,
		Direction: MD_Up,
		Behaviour: EB_Moving,
		Config: Config{
			DoorOpenDuration_s: 3.0,
		},
	} //COMMENT OUT AFTER TEST

	for {
		select {
		case <-elevatorStateCh:
			// Update our elevator object which is to be sent
		case <-requestQueueCh:
			// Update our requestqueue which is to be sent
		case <-ticker.C:
			sender.broadcastOnNetwork(elevator, msgToSend)
		}
	}
}

//func setUpSocketSender			//skal være go-routine
//		input: none
//		set up socket
//		return: none

//func broadcastInfo				//go-routine?
//		input: message
//		Use socket to broadcast info
//		return: none
