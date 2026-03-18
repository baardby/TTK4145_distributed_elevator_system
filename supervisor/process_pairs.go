package supervisor

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
)

func BackupPhase(myID int, primaryPID int) {

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

	buffer := make([]byte, 1)

	lastSeenMyID := time.Now()

	for { // First listen
		if time.Since(lastSeenMyID) > 2*time.Second {
			fmt.Println("Timeout - no heartbeat from myID within 2 seconds")
			break
		}
		connListen.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, _, err := connListen.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("Timeout - no heartbeat within 2 seconds")
				break
			} else {
				fmt.Println("Other error:", err)
			}
		}
		msgID := int(buffer[0])

		if msgID == myID {
			lastSeenMyID = time.Now()
		}

	}
	connListen.Close()

	// Kill primary process if it is still running
	if primaryPID > 0 && primaryPID != os.Getpid() {
		proc, _ := os.FindProcess(primaryPID)
		_ = proc.Kill()
	}

	// Start a new terminal with a new instance of the elevator process
	pid := os.Getpid()
	path, err := os.Getwd()
	if err != nil {
		fmt.Println("Could not fetch path:", err)
		return
	}
	cmd := exec.Command(
		"gnome-terminal",
		"--",
		"bash",
		"-c",
		fmt.Sprintf("cd %s && go run main.go --id %d --pid %d; exec bash", path, myID, pid),
	)
	err = cmd.Start()
	if err != nil {
		fmt.Println("Could not start antother terminal:", err)
	}
}
