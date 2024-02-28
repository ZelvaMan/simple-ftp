package ftp

import (
	"fmt"
	"log"
	"net"
)

type SessionInfo struct {
	controlConnection *connection
	dataConnection    *connection
	cwd               string
	isLogged          bool
	username          string
}

func createSession(controlConnection *net.Conn) (*SessionInfo, error) {

	session := &SessionInfo{
		controlConnection: newConnection(controlConnection),
		dataConnection:    nil,
		cwd:               "",
		isLogged:          false,
		username:          "",
	}

	return session, nil

}

func (session *SessionInfo) Start() {
	for {

		line, err := session.controlConnection.readLine()

		if err != nil {
			log.Printf("control connection read line: %s", err)

			// TODO let server know that session was closed
			break

		}

		log.Printf("line received from control '%s'", line)

		err = session.handleCommand(string(line))
		if err != nil {
			log.Printf("handling command")
			break
		}
	}

	// close the connection
}

func (session *SessionInfo) handleCommand(command string) error {
	err := session.controlConnection.writeLine("Hello from server")
	if err != nil {
		return fmt.Errorf("handling command: %s", err)
	}
	return nil
}

// about session is case of server shutdown
func (session *SessionInfo) Abort() {

	// TODO send abort message
}
