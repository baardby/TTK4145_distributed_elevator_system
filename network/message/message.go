package message

import (
	. "distributed_elevator/elevalgo"
	. "distributed_elevator/elevio"
	. "distributed_elevator/global_state_manager/elevator_states"
	"encoding/json"
	"fmt"
)

type Message struct {
	ID                int
	Peer              ElevatorPeer
	RequestStates     [N_FLOORS][N_BUTTONS]byte
	RequestAssignedTo [N_FLOORS][N_BUTTONS]byte
}

func ReconstructMessageFromSlice(msgBuffer []byte, msgSize int) (recvMsg Message) {
	err := json.Unmarshal(msgBuffer[:msgSize], &recvMsg)
	if err != nil {
		fmt.Println("unmarshal error:", err)
	}

	return
}

func ConstructMessageToSlice(myself Elevator, msg Message) []byte {
	msg.Peer.Floor = myself.Floor
	msg.Peer.Behaviour = myself.Behaviour
	msg.Peer.Direction = myself.Direction
	msg.Peer.Alive = true

	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("marshal error:", err)
	}

	return data
}
