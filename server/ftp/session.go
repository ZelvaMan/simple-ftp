package ftp

import (
	"fmt"
	"log"
	"net"
)

type SessionInfo struct {
	controlConnection *connection
	dataConnection    *dataConnection
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
	log.Printf("session is starting...")

	err := session.controlConnection.writeLine(formatResponse(220, "zmftp ready for new user"))

	if err != nil {
		log.Printf("Error sending hello msg: %s", err)
	}

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

		log.Printf("command handeled")

	}

	// close the connection
	session.controlConnection.close()
	session.dataConnection.close()
}

func (session *SessionInfo) handleCommand(commandLine string) error {

	//splittedLine := strings.Split(commandLine, " ")
	//command := splittedLine[0]
	//
	//switch command {
	//case "PASV":
	//	dataConn, err := openPassiveDataConnection()
	//	if err != nil {
	//		return fmt.Errorf("error opening data connection: %s", err)
	//	}
	//	// listener started
	//	session.dataConnection = dataConn
	//
	//}

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
