package elevio

const N_BUTTON_EVENTS = N_BUTTONS*N_FLOORS - 2

type BtnEventQueue struct {
	queue         [N_BUTTON_EVENTS]ButtonEvent
	queueHead     int
	queueTail     int
	numOfElements int
}

func MakeButtonEventQueue() BtnEventQueue {
	return BtnEventQueue{
		queueHead:     0,
		queueTail:     0,
		numOfElements: 0,
	}
}

func (btnEventQueue *BtnEventQueue) EnqueueButtonEvent(btnEvent ButtonEvent) {
	if !(btnEvent.Button == BT_HallDown && btnEvent.Floor == 0) {
		if !(btnEvent.Button == BT_HallUp && btnEvent.Floor == 3) {
			if !btnEventQueue.ButtonEventIsInQueue(btnEvent) {
				btnEventQueue.queue[btnEventQueue.queueTail] = btnEvent
				btnEventQueue.queueTail = (btnEventQueue.queueTail + 1) % (N_BUTTON_EVENTS)
				btnEventQueue.numOfElements++
			}
		}
	}
}

func (btnEventQueue *BtnEventQueue) DequeueButtonEvent() (btnEvent ButtonEvent) {
	if btnEventQueue.numOfElements != 0 {
		btnEvent = btnEventQueue.queue[btnEventQueue.queueHead]
		btnEventQueue.queueHead = (btnEventQueue.queueHead + 1) % (N_BUTTON_EVENTS)
		btnEventQueue.numOfElements--
		return
	}
	return
}

func (btnEventQueue *BtnEventQueue) ButtonEventIsInQueue(btnEvent ButtonEvent) bool {
	for btn := btnEventQueue.queueHead; btn < btnEventQueue.queueTail; btn++ {
		if btnEvent.Floor == btnEventQueue.queue[btn].Floor {
			if int(btnEvent.Button) == int(btnEventQueue.queue[btn].Button) {
				return true
			}
		}
	}
	return false
}

func (btnEventQueue *BtnEventQueue) IsEmpty() bool {
	return btnEventQueue.numOfElements == 0
}
