package supervisor

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func processPairs() {

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
	buffer := make([]byte, 1024)

	for { // Først lytte
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
	}
	connListen.Close()

	// Drepe den første prosessen
	var primaryPID int
	if len(os.Args) > 1 {
		pidStr := os.Args[1]
		pid, err := strconv.Atoi(pidStr)
		if err == nil {
			primaryPID = pid
		}
	}
	proc, err := os.FindProcess(primaryPID)
	if err == nil {
		proc.Kill()
	}

	// starte en ny terminal med en ny prosess
	pid := os.Getpid()
	cmd := exec.Command(
		"gnome-terminal",
		"--",
		"bash",
		"-c", //LEGG TIL RIKTIG PATH HER!!!
		fmt.Sprintf("cd /path/to/project && go run main.go %d; exec bash", pid),
	)
	err = cmd.Start()
	if err != nil {
		fmt.Println("Feil ved oppstart av ny terminal:", err)
	}
}
