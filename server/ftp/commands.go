package ftp

import (
	"fmt"
	"log"
	"server/respones"
	"server/sequences"
	"strings"
)

func (session *SessionInfo) handleCommand(commandLine string) error {

	command, argument, hasArguments := strings.Cut(commandLine, " ")
	if !hasArguments {
		command = commandLine
	}

	var err error
	switch command {
	case "USER":
		err = session.handleUSER(argument)
	case "PASS":
		err = session.handlePASS(argument)
	case "PASV":
		log.Printf("passive connection requested")
		dataConn, err := openPassiveDataConnection()
		if err != nil {
			return fmt.Errorf("error opening data connection: %s", err)
		}
		// listener started
		session.dataConnection = dataConn
	default:
		log.Printf("Command %s is not implemented", command)

		err = session.Respond(respones.NotImplemented())
	}

	if err != nil {
		return fmt.Errorf("handling command: %s", err)
	}

	return nil
}

func (session *SessionInfo) handleUSER(username string) error {
	session.username = username

	err := session.Respond(respones.PasswordNeeded())
	if err != nil {
		return fmt.Errorf("handling USER command: %s", err)
	}

	session.commandSequence = sequences.PASSWORD
	return nil
}

func (session *SessionInfo) handlePASS(password string) error {
	// check sequence
	if session.commandSequence != sequences.PASSWORD {
		log.Printf("wrong sequence")
		err := session.Respond(respones.BadSequence())

		if err != nil {
			return fmt.Errorf("error sending bad sequence: %s", err)
		}
	}

	log.Printf("trying to authenticate user %s", session.username)

	// wrong password/username
	if !authenticateUser(session.username, password) {
		log.Printf("Wrong user name or pasword")

		err := session.Respond(respones.NotLoggedIn())
		if err != nil {
			return fmt.Errorf("error sending not logged in: %s", err)
		}

		return nil
	}

	log.Printf("user authenticated")

	err := session.Respond(respones.UserLoggedIn())
	if err != nil {
		return fmt.Errorf("error sending logged in: %s", err)
	}

	// login ok
	session.isLogged = true
	session.commandSequence = sequences.NONE

	return nil
}
