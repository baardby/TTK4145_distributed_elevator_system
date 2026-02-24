package network

import (
	"bytes"
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevator_states"
	. "distributed_elevator/elevio"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

type NetworkSender struct {
	DestID        int // DO WE NEED THIS?
	DestIP        string
	DestPort      string
	DestAddr      *net.UDPAddr
	MyConn        *net.UDPConn // Remember to add defer myConn.Close() in the loop the sender is run
	NumberOfPeers int
	ListOfPeers   map[string]int
}

func (sender *NetworkSender) networkSenderInit() {
	var err error

	sender.NumberOfPeers = 0
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
	_, err := sender.MyConn.WriteToUDP(constructMessageToSlice(myself, msg), sender.DestAddr) //UNCOMMENT AFTER TEST
	//_, err := sender.MyConn.WriteToUDP([]byte("Hello from fedora!"), sender.DestAddr) //COMMENT OUT AFTER TEST
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

	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, msg)

	return buf.Bytes()
}

func Network_SenderFSM(newPeerCh <-chan string) {
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
		case newPeer := <-newPeerCh:
			_, isInPeerList := sender.ListOfPeers[newPeer]
			if !isInPeerList {
				sender.NumberOfPeers++
				sender.ListOfPeers[newPeer] = sender.NumberOfPeers
			}
		case <-ticker.C:
			sender.broadcastOnNetwork(elevator, msgToSend)
			sender.testPrintPeerList()
		}
	}
}

func (sender *NetworkSender) testPrintPeerList() {
	fmt.Println("----Alive peers----")
	for key, value := range sender.ListOfPeers {
		fmt.Println(key, value)
	}
	fmt.Println("-------------------")
}

//func setUpSocketSender			//skal være go-routine
//		input: none
//		set up socket
//		return: none

//func broadcastInfo				//go-routine?
//		input: message
//		Use socket to broadcast info
//		return: none
