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

type networkListener struct {
	myPort      string
	myIP        string // FOR TESTING. TODO: Remove after testing
	myConn      *net.UDPConn
	listOfPeers map[string]int // FOR TESTING. TODO: Remove after testing
}

func NetworkListener(myID int,
	peerAlive chan<- int,
	receivedMessage chan<- Message) {

	var listener networkListener
	listener.networkListenerInit()
	defer listener.myConn.Close()

	var recvMsg Message
	var reconstructErr error

	var recvAddr *net.UDPAddr
	var msgSize int
	recvDecodedMsg := make([]byte, 1024)

	for {
		recvAddr, recvDecodedMsg, msgSize = listener.readFromNetwork()
		recvMsg, reconstructErr = ReconstructMessageFromSlice(recvDecodedMsg, msgSize)

		// Filters out messages which doesn't follow the correct format of the network and messages broadcasted to ourselves
		if (reconstructErr == nil) && (recvMsg.NetworkCode == NETWORK_CODE) && (recvMsg.ID != myID) {

			// FOR TESTING. TODO: Remove after testing
			testPrintRecvMsg(&recvMsg)
			_, isInPeerList := listener.listOfPeers[recvAddr.IP.String()]
			if !isInPeerList {
				listener.listOfPeers[recvAddr.IP.String()] = recvMsg.ID
			}
			listener.testPrintPeerList()
			// END OF TESTING

			// Notify Supervisor of new msg from peer
			peerAlive <- recvMsg.ID

			// Send message to global state manager
			receivedMessage <- recvMsg
		}
	}
}

func (listener *networkListener) networkListenerInit() {
	var err error
	var myAddr *net.UDPAddr

	// FOR TESTING CONNECTIONS. TODO: Remove after testing
	listener.listOfPeers = make(map[string]int)
	listener.myIP, err = LocalIP()
	// END OF TESTING

	listener.myPort = "20003"

	// We have to bind to 0.0.0.0 to be able to pickup broadcasts
	myAddr, err = net.ResolveUDPAddr("udp4", "0.0.0.0"+":"+listener.myPort)
	if err != nil {
		log.Fatalf("Failed to bind UDP socket %v", err)
	}

	listener.myConn, err = net.ListenUDP("udp4", myAddr)
	if err != nil {
		log.Fatalf("Failed to bind UDP socket %v", err)
	}
}

func (listener *networkListener) readFromNetwork() (*net.UDPAddr, []byte, int) {
	decodedMsg := make([]byte, 1024)

	msgSize, recvAddr, readErr := listener.myConn.ReadFromUDP(decodedMsg)
	if readErr != nil {
		fmt.Println("Message error:", readErr)
	}

	return recvAddr, decodedMsg, msgSize
}

// TODO: Remove testing functions after testing is done
func testPrintRecvMsg(recvMsg *Message) {
	fmt.Println("--Received message--")
	fmt.Println(recvMsg.Peer.WorkingStatus)
	fmt.Println(recvMsg.Peer.Floor)
	fmt.Println(Elevator_BehaviourToString(recvMsg.Peer.Behaviour))
	fmt.Println(Elevator_MotorDirectionToString(recvMsg.Peer.Direction))
	fmt.Println("--------------------")
}

func (listener *networkListener) testPrintPeerList() {
	fmt.Println("----Alive peers----")
	for key, value := range listener.listOfPeers {
		fmt.Println(key, value)
	}
	fmt.Println("-------------------")
}
