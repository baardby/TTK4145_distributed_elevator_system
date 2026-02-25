package supervisor //Må endres hvis det puttes inn i en mappe

//func prossessPairs
//		input: none
//		lytt først
//		Oppdater state
//		Når ingen beskjed motatt
//		drep primary
//		Spawn ny terminal med samme kode
//		send meldinger kontinuerlig
//		Lytt etter andre sin requestqueue
//			adopt hvis motatt
//		Kjør all annen funksjonalitet
//		return none

//func amIWorking
//		input: none
//		forskjellgie tester på meg selv

//func listenForErrors
//		input: error
//		kjør restart, eller løs problemet.
//		return: none

//func passOnInfo
//		input: message from network_listener
//		Gir meldingen mening? 					//Kan kanskje klare oss uten
//		Del opp info
//		kaller elevatorInfo og request_queueInfo
//		Resetter timer på elevator #
//		return: none

//func request_queueInfo
//		input: køen til en annen heis
//		kaller queueUnion
//		return: none

//func elevatorInfo
//		input: info om elevator #
//		sjekker om infoen er ulik den gamle
//		hvis ulik:								//sjekker ikke om stemmer, siden ny info mottas hele tiden
//			kaller updateElevatorStates med input elevator #
//		else: none
//		return: none

//func watchDogTimer							//Hva er denne egt til?
