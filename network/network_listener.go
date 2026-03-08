package network

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/network/localip"
	. "distributed_elevator/network/message"
	"fmt"
	"log"
	"net"
)

type NetworkListener struct {
	MyPort        string
	MyIP          string
	MyConn        *net.UDPConn // Remember to add defer myConn.Close() in the loop the listener is run
	NumberOfPeers int
	ListOfPeers   map[string]int
}

func (listener *NetworkListener) networkListenerInit() {
	var err error
	var myAddr *net.UDPAddr

	listener.NumberOfPeers = 0
	listener.ListOfPeers = make(map[string]int)

	listener.MyPort = "20003"
	listener.MyIP, err = LocalIP() // Save our local IP to be able to filter out these messages afterwards

	myAddr, err = net.ResolveUDPAddr("udp4", "0.0.0.0"+":"+listener.MyPort) // We have to bind to 0.0.0.0 to be able to pickup broadcasts
	if err != nil {                                                         // ADD ERROR HANDLING
		log.Fatalf("Failed to bind UDP socket %v", err)
	}

	listener.MyConn, err = net.ListenUDP("udp4", myAddr)
	// ADD ERROR HANDLING
}

func (listener *NetworkListener) readFromNetwork() (recvAddr *net.UDPAddr, recvMsg Message) {
	var err error
	var msgSize int
	msgBuffer := make([]byte, 1024)

	msgSize, recvAddr, err = listener.MyConn.ReadFromUDP(msgBuffer)
	if err != nil { // ADD ERROR HANDLING
		log.Fatalf("Message error: %v", err)
	}

	if recvAddr.IP.String() == listener.MyIP { // Eliminating broadcasts to myself
		return
	}

	recvMsg = ReconstructMessageFromSlice(msgBuffer, msgSize)

	return
}

func Network_ListenerLoop(receivedFromPeerEvent chan<- int,
	receivedMessageEvent chan<- Message) {
	var listener NetworkListener
	listener.networkListenerInit()
	defer listener.MyConn.Close()

	for {
		recvAddr, recvMsg := listener.readFromNetwork()

		// Filter out our own messages again
		if !(recvAddr.IP.String() == listener.MyIP) {
			//testPrintRecvMsg(&recvMsg) // For testing

			// Adding a new peer to the list
			_, isInPeerList := listener.ListOfPeers[recvAddr.IP.String()]
			if !isInPeerList {
				listener.NumberOfPeers++
				listener.ListOfPeers[recvAddr.IP.String()] = recvMsg.ID
			}
			//listener.testPrintPeerList()

			// Notify Supervisor of new msg from peer
			receivedFromPeerEvent <- recvMsg.ID

			// Send message to global state manager
			receivedMessageEvent <- recvMsg
		}
	}
}

func testPrintRecvMsg(recvMsg *Message) {
	fmt.Println("------")
	fmt.Println(recvMsg.Peer.Floor)
	fmt.Println(Elevator_BehaviourToString(recvMsg.Peer.Behaviour))
	fmt.Println(Elevator_MotorDirectionToString(recvMsg.Peer.Direction))
}

func (listener *NetworkListener) testPrintPeerList() {
	fmt.Println("----Alive peers----")
	for key, value := range listener.ListOfPeers {
		fmt.Println(key, value)
	}
	fmt.Println("-------------------")
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
