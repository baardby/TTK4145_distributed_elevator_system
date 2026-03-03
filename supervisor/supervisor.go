package supervisor //Må endres hvis det puttes inn i en mappe

import (
	"distributed_elevator/HealthTimers"
)

type Ports struct {
	TimerCommand chan<- HealthTimers.TimerCommand
	TimerEvent   <-chan HealthTimers.TimerEvent

	//Legg til flere kanaler for de forskjellige modulene
}

func handleTimeEvent(timeEvent int) {
	//handle time event
}

func handleHardwareEvent(hardwareEvent int) {
	//handle hardware event
}

func handleNetwork_ListenerEvent(network_listenerEvent int) {
	//handle network listener event
}

func Supervisor() {
	backupPhase()
	//listen for others queue
	//if no response, init queue
	//make all the channels and init funktion for the different modules
	//set own state
	timeChan := make(chan int)             //channel for timers
	hardwareChan := make(chan int)         //channel for hardware events
	network_listenerChan := make(chan int) //channel for network listener events

	for {
		//send checkpoint to backup
		select {
		case timeEvent := <-timeChan:
			handleTimeEvent(timeEvent)
		case hardwareEvent := <-hardwareChan:
			handleHardwareEvent(hardwareEvent)
		case network_listenerEvent := <-network_listenerChan:
			handleNetwork_ListenerEvent(network_listenerEvent)
			//default: //kjør acceptance test, og send melding til backup
		}
	}
}

//func amIWorking
//		input: none
//		forskjellgie tester på meg selv
//			Kjører jeg når jeg skal kjøre?
//		return: bool

//func listenForErrors
//		input: error
//		kjør restart, eller løs problemet.
//		return: none
