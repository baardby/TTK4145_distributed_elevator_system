package supervisor //Må endres hvis det puttes inn i en mappe

import (
	"time"
)

type elevatorTimer struct {
	startTime time.Time
	duration  time.Duration
}

type supervisor struct {
	elevatorTimerList []elevatorTimer
}

func (s *supervisor) resetElevatorTimer(elevatorID int) {
	s.elevatorTimerList[elevatorID].startTime = time.Now()
}

func (s *supervisor) checkElevatorTimer(elevatorID int) bool {
	currentTime := time.Now()
	elapsedTime := currentTime.Sub(s.elevatorTimerList[elevatorID].startTime)
	return elapsedTime > s.elevatorTimerList[elevatorID].duration
}

func Supervisor() {
	processPairs()
}

//func amIWorking
//		input: none
//		forskjellgie tester på meg selv

//func listenForErrors
//		input: error
//		kjør restart, eller løs problemet.
//		return: none

//func watchDogTimer							//Hva er denne egt til?
