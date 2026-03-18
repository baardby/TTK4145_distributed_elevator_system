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

type networkSender struct {
	destIP     string
	destPort   string
	destAddr   *net.UDPAddr
	myConn     *net.UDPConn
	myElevator ElevatorPeer
	hallOrders AllHallOrders
	cabOrders  AllCabOrders
}

type backupConn struct {
	addr *net.UDPAddr
	conn *net.UDPConn
}

func NetworkSender(myID int,
	updateElevatorStateEvent <-chan ElevatorPeer,
	updateOrderQueueEvent <-chan OrderQueue,
	heartBeatPing <-chan int) {

	var sender networkSender
	sender.networkSenderInit()
	defer sender.myConn.Close()

	backup := initializeBackupConn("30000")
	defer backup.conn.Close()

	var msgToSend Message
	msgToSend.ID = myID
	msgToSend.NetworkCode = NETWORK_CODE
	msgToSend.UpdateMessage(sender.myElevator, sender.hallOrders, sender.cabOrders)

	time.Sleep(200 * time.Millisecond) // Sleep to let other goroutines begin

	// Setting up periodic sending
	sendTicker := time.NewTicker(100 * time.Millisecond)
	defer sendTicker.Stop()

	for {
		select {
		case newElevator := <-updateElevatorStateEvent:
			sender.updateMyElevator(newElevator)

			msgToSend.UpdateMessage(sender.myElevator, sender.hallOrders, sender.cabOrders)

		case newOrderQueue := <-updateOrderQueueEvent:
			sender.updateHallOrderQueue(newOrderQueue.Hall[myID])
			sender.updateCabOrderQueue(newOrderQueue.Cab[myID])

			msgToSend.UpdateMessage(sender.myElevator, sender.hallOrders, sender.cabOrders)

		case heartBeat := <-heartBeatPing:
			sendHeartBeat(heartBeat, &backup)

		case <-sendTicker.C:
			sender.broadcastOnNetwork(msgToSend)

		}
	}
}

func (sender *networkSender) networkSenderInit() {
	var err error

	sender.destPort = "20003"
	sender.destIP = "255.255.255.255"

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < HallButtonsPerFloor; btn++ {
			sender.hallOrders[floor][btn] = HallOrder{
				State:      None,
				AssignedTo: NoElevatorAssigned,
			}
		}
		for elevatorID := 0; elevatorID < N_ELEVATORS; elevatorID++ {
			sender.cabOrders[floor][elevatorID] = None
		}
	}

	sender.destAddr, err = net.ResolveUDPAddr("udp4", sender.destIP+":"+sender.destPort)
	if err != nil {
		log.Fatalf("Could not resolve address: %v", err)
	}

	sender.myConn, err = net.ListenUDP("udp4", nil)
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}
}

func (sender *networkSender) broadcastOnNetwork(msg Message) error {
	_, err := sender.myConn.WriteToUDP(ConstructMessageToSlice(msg), sender.destAddr)
	if err != nil {
		fmt.Println("Sending message error:", err)
	}

	return err
}

func (sender *networkSender) updateMyElevator(newElevator ElevatorPeer) {
	sender.myElevator = newElevator
}

func (sender *networkSender) updateHallOrderQueue(newHallOrderQueue AllHallOrders) {
	sender.hallOrders = newHallOrderQueue
}

func (sender *networkSender) updateCabOrderQueue(newCabOrderQueue AllCabOrders) {
	sender.cabOrders = newCabOrderQueue
}

func initializeBackupConn(port string) backupConn {
	var backup backupConn
	var err error

	backup.addr, err = net.ResolveUDPAddr("udp4", "localhost:"+port)
	if err != nil {
		log.Fatalf("Could not resolve address: %v", err)
	}

	backup.conn, err = net.ListenUDP("udp4", nil)
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}

	return backup
}

func sendHeartBeat(myId int, backupConn *backupConn) error {
	heartBeatMsg := []byte{byte(myId)}

	_, err := backupConn.conn.WriteToUDP(heartBeatMsg, backupConn.addr)
	if err != nil {
		fmt.Println("Error sending heartbeat:", err)
	}

	return err
}
