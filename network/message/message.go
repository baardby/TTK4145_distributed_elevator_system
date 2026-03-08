package message

import (
	. "distributed_elevator/global_state_manager/elevator_states"
	. "distributed_elevator/global_state_manager/order_queue"
	"encoding/json"
	"fmt"
)

type Message struct {
	ID          int
	Peer        ElevatorPeer
	GlobalQueue OrderQueue
}

func ReconstructMessageFromSlice(msgBuffer []byte, msgSize int) (recvMsg Message) {
	err := json.Unmarshal(msgBuffer[:msgSize], &recvMsg)
	if err != nil {
		fmt.Println("unmarshal error:", err)
	}

	return
}

func ConstructMessageToSlice(myself ElevatorPeer, msg Message) []byte {
	msg.Peer = myself

	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("marshal error:", err)
	}

	return data
}
