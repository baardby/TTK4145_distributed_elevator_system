package network

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/network/localip"
	. "distributed_elevator/supervisor"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type Message struct {
	Peer              ElevatorPeer
	RequestStates     [N_FLOORS][N_BUTTONS]byte
	RequestAssignedTo [N_FLOORS][N_BUTTONS]byte
}

type NetworkListener struct {
	MyPort string
	MyIP   string
	MyConn *net.UDPConn // Remember to add defer myConn.Close() in the loop the listener is run
}

func (listener *NetworkListener) networkListenerInit() {
	var err error
	var myAddr *net.UDPAddr

	listener.MyPort = "20003"
	listener.MyIP, err = LocalIP() // Save our local IP to be able to filter out these messages afterwards

	myAddr, err = net.ResolveUDPAddr("udp4", "0.0.0.0"+":"+listener.MyPort) // We have to bind to 0.0.0.0 to be able to pickup broadcasts
	if err != nil {                                                         // ADD ERROR HANDLING
		log.Fatalf("Failed to bind UDP socket %v", err)
	}

	listener.MyConn, err = net.ListenUDP("udp4", myAddr)
	// ADD ERROR HANDLING
}

func (listener *NetworkListener) readFromNetwork() (recvAddr *net.UDPAddr, recvMsg Message) { // MAYBE RETURN ID OF ELEVATOR PEER ALSO
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

	recvMsg = reconstructMessageFromSlice(msgBuffer, msgSize)

	return
}

func reconstructMessageFromSlice(msgBuffer []byte, msgSize int) (recvMsg Message) {
	err := json.Unmarshal(msgBuffer[:msgSize], &recvMsg)
	if err != nil {
		fmt.Println("unmarshal error:", err)
	}

	return
}

func Network_ListenerFSM(newPeerCh chan<- string) {
	var listener NetworkListener
	listener.networkListenerInit()
	defer listener.MyConn.Close()

	for {
		recvAddr, recvMsg := listener.readFromNetwork()
		if !(recvAddr.IP.String() == listener.MyIP) {
			newPeerCh <- recvAddr.IP.String()
			//testPrintRecvMsg(&recvMsg) // For testing
		}
	}
}

func testPrintRecvMsg(recvMsg *Message) {
	fmt.Println("------")
	fmt.Println(recvMsg.Peer.Floor)
	fmt.Println(Elevator_BehaviourToString(recvMsg.Peer.Behaviour))
	fmt.Println(Elevator_MotorDirectionToString(recvMsg.Peer.Direction))
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
