package network

import (
	. "distributed_elevator/elevio"
	. "distributed_elevator/global_state_manager/elevator_states"
	. "distributed_elevator/global_state_manager/order_queue"
	. "distributed_elevator/network/message"
	"fmt"
	"log"
	"net"
	"time"
)

type NetworkSender struct {
	DestIP     string
	DestPort   string
	DestAddr   *net.UDPAddr
	MyConn     *net.UDPConn // Remember to add defer myConn.Close() in the loop the sender is run
	MyElevator ElevatorPeer
	HallOrders AllHallOrders
	CabOrders  AllCabOrders
}

func (sender *NetworkSender) networkSenderInit() {
	var err error

	sender.DestPort = "20003"
	sender.DestIP = "255.255.255.255"

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < HallButtonsPerFloor; btn++ {
			sender.HallOrders[floor][btn] = HallOrder{
				State:      None,
				AssignedTo: NoElevatorAssigned,
			}
		}
		for elevatorID := 0; elevatorID < N_ELEVATORS; elevatorID++ {
			sender.CabOrders[floor][elevatorID] = None
		}
	}

	sender.DestAddr, err = net.ResolveUDPAddr("udp4", sender.DestIP+":"+sender.DestPort)
	if err != nil { // ADD ERROR HANDLING
		log.Fatalf("Could not resolve address: %v", err)
	}

	sender.MyConn, err = net.ListenUDP("udp4", nil)
	if err != nil { // ADD ERROR HANDLING
		log.Fatalf("Error dialing: %v", err)
	}
}

func (sender *NetworkSender) broadcastOnNetwork(msg Message) error {
	_, err := sender.MyConn.WriteToUDP(ConstructMessageToSlice(msg), sender.DestAddr)
	if err != nil { // ADD ERROR HANDLING
		fmt.Println("Sending message error:", err)
	}

	return err
}

func (sender *NetworkSender) updateMyElevator(newElevator ElevatorPeer) {
	sender.MyElevator = newElevator
}

func (sender *NetworkSender) updateHallOrderQueue(newHallOrderQueue AllHallOrders) {
	sender.HallOrders = newHallOrderQueue
}

func (sender *NetworkSender) updateCabOrderQueue(newCabOrderQueue AllCabOrders) {
	sender.CabOrders = newCabOrderQueue
}

func Network_SenderLoop(myID int,
	updateElevatorStateEvent <-chan ElevatorPeer,
	updateOrderQueueEvent <-chan OrderQueue) {

	var sender NetworkSender
	sender.networkSenderInit()
	defer sender.MyConn.Close()

	var msgToSend Message
	msgToSend.ID = myID
	msgToSend.NetworkCode = NETWORK_CODE

	time.Sleep(200 * time.Millisecond) // Sleep to let other goroutines begin

	// Setting up periodic sending
	sendTicker := time.NewTicker(100 * time.Millisecond) // TODO: Change to 10Hz
	defer sendTicker.Stop()

	for {
		select {
		case newElevator := <-updateElevatorStateEvent:
			sender.updateMyElevator(newElevator)

			msgToSend.UpdateMessage(sender.MyElevator, sender.HallOrders, sender.CabOrders)

		case newOrderQueue := <-updateOrderQueueEvent:
			sender.updateHallOrderQueue(newOrderQueue.Hall[myID])
			sender.updateCabOrderQueue(newOrderQueue.Cab[myID])

			msgToSend.UpdateMessage(sender.MyElevator, sender.HallOrders, sender.CabOrders)

		case <-sendTicker.C:
			sender.broadcastOnNetwork(msgToSend)
		}
	}
}
