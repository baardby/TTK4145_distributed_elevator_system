package message

import (
	. "distributed_elevator/global_state_manager/elevator_states"
	. "distributed_elevator/global_state_manager/order_queue"
	"encoding/json"
	"fmt"
)

type Message struct {
	NetworkCode string
	ID          int
	Peer        ElevatorPeer
	HallOrders  AllHallOrders
	CabOrders   AllCabOrders
}

func ReconstructMessageFromSlice(msgBuffer []byte, msgSize int) (recvMsg Message, deconstructErr error) {
	deconstructErr = json.Unmarshal(msgBuffer[:msgSize], &recvMsg)
	if deconstructErr != nil {
		fmt.Println("unmarshal error:", deconstructErr)
	}

	return
}

func ConstructMessageToSlice(msg Message) []byte {
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("marshal error:", err)
	}

	return data
}

func (msg *Message) UpdateMessage(myElevator ElevatorPeer, hallOrders AllHallOrders, cabOrders AllCabOrders) {
	msg.Peer = myElevator
	msg.HallOrders = hallOrders
	msg.CabOrders = cabOrders
}
