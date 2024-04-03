package ftp

import (
	"fmt"
	"log"
	"strings"
)

func (session *SessionInfo) handleCommand(commandLine string) error {

	splittedLine := strings.Split(commandLine, " ")
	command := splittedLine[0]

	switch command {
	case "USER":
		if len(splittedLine) != 2 {
			return fmt.Errorf("incorrect format for USER command: %s", splittedLine)
		}

		session.username = splittedLine[1]
		err := session.controlConnection.writeLine(userLoggedInResponse())
		if err != nil {
			return fmt.Errorf("error sending response: %s", err)
		}
	case "PASV":
		log.Printf("passive connection requested")
		dataConn, err := openPassiveDataConnection()
		if err != nil {
			return fmt.Errorf("error opening data connection: %s", err)
		}
		// listener started
		session.dataConnection = dataConn

	}

	err := session.controlConnection.writeLine("Hello from server")

	if err != nil {
		return fmt.Errorf("handling command: %s", err)
	}

	return nil
}
