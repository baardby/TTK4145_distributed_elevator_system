package main //Må endres hvis det puttes inn i en mappe

//type enum progress					//Enum for å fikse Union-problemet
//		none, unconfirmed, confirmed
//		Barriers?

//type request							//Hver og en bestilling er av denne typen
// 		floor
// 		button_type
// 		elevator
// 		progress

//type requests_queue					//Hele køen, SKAL DELES (Stor bokstav)
//		attr: list of requests

//func init_queue						//Nødvendig?
//		input: none
//		Gjør noe
//		return: none

//func restoreQueueAfterDisconnect 		//For å fikse når man kommer tilbake på nettet
//		input: None
//		Listen for other queue
//		Figure out if two different queues are recieved.
//		if riecieved:
//			Adopt
//		else:
//			init_queue()
//		return: None

//func broadcastQueue					//Fortelle andre om køen min
//		input: none
//		call broadcastInfo with queue as message
//		return: none

//func QueueUnion						//Motatt kø fra andre, må fikse unionen
//		input:QueueFromElevator#
//		figure out logic with Union
//		Barriers?
//		Update requests_queue
//		return: none

//func newRequest
//		input: request
//		legg til å køen
//		return: none
