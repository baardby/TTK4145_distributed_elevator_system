package network

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
	. "distributed_elevator/global_state_manager/elevator_states"
	. "distributed_elevator/global_state_manager/order_queue"
	. "distributed_elevator/network/message"
	"log"
	"net"
	"time"
)

type NetworkSender struct {
	DestID     int // DO WE NEED THIS?
	DestIP     string
	DestPort   string
	DestAddr   *net.UDPAddr
	MyConn     *net.UDPConn // Remember to add defer myConn.Close() in the loop the sender is run
	Myself     Elevator
	Orderqueue OrderQueue
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
	_, err := sender.MyConn.WriteToUDP(ConstructMessageToSlice(myself, msg), sender.DestAddr)
	if err != nil { // ADD ERROR HANDLING
		log.Fatalf("Sending message error: %v", err)
	}

	return err
}

func Network_SenderFSM(updateElevatorStateEvent <-chan Elevator, updateRequestQueueEvent <-chan OrderQueue) {
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
		case <-updateElevatorStateEvent:
			// Update our elevator object which is to be sent
		case <-updateRequestQueueEvent:
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
