package network

import (
	. "distributed_elevator/network/message"
	"fmt"
	"log"
	"net"
)

const NETWORK_CODE = "gruppe2"

type networkListener struct {
	myPort string
	myConn *net.UDPConn
}

func NetworkListener(myID int,
	peerAlive chan<- int,
	receivedMessage chan<- Message) {

	var listener networkListener
	listener.networkListenerInit()
	defer listener.myConn.Close()

	var recvMsg Message
	var reconstructErr error

	var msgSize int
	recvDecodedMsg := make([]byte, 1024)

	for {
		_, recvDecodedMsg, msgSize = listener.readFromNetwork()
		recvMsg, reconstructErr = ReconstructMessageFromSlice(recvDecodedMsg, msgSize)

		// Filters out messages which doesn't follow the correct format of the network and messages broadcasted to ourselves
		if (reconstructErr == nil) && (recvMsg.NetworkCode == NETWORK_CODE) && (recvMsg.ID != myID) {

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
