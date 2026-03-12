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
	MyIP        string
	MyConn      *net.UDPConn // Remember to add defer myConn.Close() in the loop the listener is run
	ListOfPeers map[string]int
}

func (listener *NetworkListener) networkListenerInit() {
	var err error
	var myAddr *net.UDPAddr

	listener.ListOfPeers = make(map[string]int)

	listener.MyPort = "20003"
	// Save our local IP to be able to filter out broadcasts to ourselves using IPs
	listener.MyIP, err = LocalIP()

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
		// TODO: Choose between filtering out messages based on IP or ID

		/*// Filter out broadcasts to ourselves
		if !(recvAddr.IP.String() == listener.MyIP) {
			recvMsg, deconstructErr = ReconstructMessageFromSlice(recvDecodedMsg, msgSize)

			// Filters out messages which doesn't follow the correct format of the network
			// This applies to messages that can't be reconstructed or doesn't have the correct network code
			if (deconstructErr == nil) && (recvMsg.NetworkCode == NETWORK_CODE) {
				//testPrintRecvMsg(&recvMsg) // FOR TESTING

				// Adding a new peer to the list MIGHT REMOVE
				_, isInPeerList := listener.ListOfPeers[recvAddr.IP.String()]
				if !isInPeerList {
					listener.ListOfPeers[recvAddr.IP.String()] = recvMsg.ID
				}
				//listener.testPrintPeerList() // FOR TESTING

				// Notify Supervisor of new msg from peer
				receivedFromPeerEvent <- recvMsg.ID

				// Send message to global state manager
				receivedMessageEvent <- recvMsg
			}
		}*/
		recvMsg, deconstructErr = ReconstructMessageFromSlice(recvDecodedMsg, msgSize)

		// Filters out messages which doesn't follow the correct format of the network and messages broadcasted to ourselves
		if (deconstructErr == nil) && (recvMsg.NetworkCode == NETWORK_CODE) && (recvMsg.ID != myID) {
			testPrintRecvMsg(&recvMsg) // FOR TESTING

			// Adding a new peer to the list MIGHT REMOVE
			_, isInPeerList := listener.ListOfPeers[recvAddr.IP.String()]
			if !isInPeerList {
				listener.ListOfPeers[recvAddr.IP.String()] = recvMsg.ID
			}
			listener.testPrintPeerList() // FOR TESTING

			// Notify Supervisor of new msg from peer
			//receivedFromPeerEvent <- recvMsg.ID

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
