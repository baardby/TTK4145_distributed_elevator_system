package supervisor

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
)

func BackupPhase(myId int, primaryPID int) {

	//Set up UDP connection
	localHostID := "127.0.0.1"
	localHostPort := 30000
	localHostAddr := &net.UDPAddr{
		IP:   net.ParseIP(localHostID),
		Port: localHostPort,
	}
	connListen, err := net.ListenUDP("udp", localHostAddr)
	if err != nil {
		panic(err)
	}
	defer connListen.Close()

	//Set up some kind of way to store checkpoint
	buffer := make([]byte, 1)

	lastSeenMyID := time.Now()

	for { // Først lytte
		if time.Since(lastSeenMyID) > 2*time.Second {
			fmt.Println("Timeout - ingen melding fra riktig ID på 2 sekunder")
			break
		}
		connListen.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, _, err := connListen.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("Timeout - ingen melding på 2 sekunder")
				break
			} else {
				fmt.Println("Annen feil:", err)
			}
		}
		msgID := int(buffer[0])

		if msgID == myId {
			lastSeenMyID = time.Now()
		}

	}
	connListen.Close()

	// Drepe den første prosessen
	if primaryPID > 0 && primaryPID != os.Getpid() {
		proc, _ := os.FindProcess(primaryPID)
		_ = proc.Kill()
	}

	// starte en ny terminal med en ny prosess
	pid := os.Getpid()
	path, err := os.Getwd()
	if err != nil {
		fmt.Println("Feil ved henting av eksekverbar fil:", err)
		return
	}
	cmd := exec.Command(
		"gnome-terminal",
		"--",
		"bash",
		"-c",
		fmt.Sprintf("cd %s && go run main.go --id %d --pid %d; exec bash", path, myId, pid),
	)
	err = cmd.Start()
	if err != nil {
		fmt.Println("Feil ved oppstart av ny terminal:", err)
	}
}
