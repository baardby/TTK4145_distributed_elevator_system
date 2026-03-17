package network

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/network/localip"
	. "distributed_elevator/network/message"
	"fmt"
	"log"
	"net"
)

const NETWORK_CODE = "gruppe2"

type NetworkListener struct {
	MyPort      string
	MyIP        string // FOR TESTING. TODO: Remove after testing
	MyConn      *net.UDPConn
	ListOfPeers map[string]int // FOR TESTING. TODO: Remove after testing
}

func Network_ListenerLoop(myID int,
	receivedFromPeerEvent chan<- int,
	receivedMessageEvent chan<- Message) {
	
	var listener NetworkListener
	listener.networkListenerInit()
	defer listener.MyConn.Close()

	var recvMsg Message
	var deconstructErr error

	var recvAddr *net.UDPAddr
	var msgSize int
	recvDecodedMsg := make([]byte, 1024)

	for {
		recvAddr, recvDecodedMsg, msgSize = listener.readFromNetwork()
		recvMsg, deconstructErr = ReconstructMessageFromSlice(recvDecodedMsg, msgSize)

		// Filters out messages which doesn't follow the correct format of the network and messages broadcasted to ourselves
		if (deconstructErr == nil) && (recvMsg.NetworkCode == NETWORK_CODE) && (recvMsg.ID != myID) {
			
			// FOR TESTING. TODO: Remove after testing
			testPrintRecvMsg(&recvMsg)
			_, isInPeerList := listener.ListOfPeers[recvAddr.IP.String()]
			if !isInPeerList {
				listener.ListOfPeers[recvAddr.IP.String()] = recvMsg.ID
			}
			listener.testPrintPeerList()
			// END OF TESTING

			// Notify Supervisor of new msg from peer
			receivedFromPeerEvent <- recvMsg.ID

			// Send message to global state manager
			receivedMessageEvent <- recvMsg
		}
	}
}

func (listener *NetworkListener) networkListenerInit() {
	var err error
	var myAddr *net.UDPAddr

	listener.ListOfPeers = make(map[string]int)

	listener.MyPort = "20003"
	// FOR TESTING CONNECTIONS. TODO: Remove after testing
	listener.MyIP, err = LocalIP()
	// END OF TESTING

	// We have to bind to 0.0.0.0 to be able to pickup broadcasts
	myAddr, err = net.ResolveUDPAddr("udp4", "0.0.0.0"+":"+listener.MyPort)
	if err != nil { // ADD ERROR HANDLING
		log.Fatalf("Failed to bind UDP socket %v", err)
	}

	listener.MyConn, err = net.ListenUDP("udp4", myAddr)
	// ADD ERROR HANDLING
}

func (listener *NetworkListener) readFromNetwork() (*net.UDPAddr, []byte, int) {
	decodedMsg := make([]byte, 1024)

	msgSize, recvAddr, readErr := listener.MyConn.ReadFromUDP(decodedMsg)
	if readErr != nil { // ADD ERROR HANDLING
		fmt.Println("Message error:", readErr)
	}

	return recvAddr, decodedMsg, msgSize
}

func testPrintRecvMsg(recvMsg *Message) {
	fmt.Println("--Received message--")
	fmt.Println(recvMsg.Peer.WorkingStatus)
	fmt.Println(recvMsg.Peer.Floor)
	fmt.Println(Elevator_BehaviourToString(recvMsg.Peer.Behaviour))
	fmt.Println(Elevator_MotorDirectionToString(recvMsg.Peer.Direction))
	fmt.Println("--------------------")
}

func (listener *NetworkListener) testPrintPeerList() {
	fmt.Println("----Alive peers----")
	for key, value := range listener.ListOfPeers {
		fmt.Println(key, value)
	}
	fmt.Println("-------------------")
}
