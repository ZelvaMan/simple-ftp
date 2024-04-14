package ftp

import (
	"errors"
	"fmt"
	"log"
	"server/respones"
	"server/sequences"
	"slices"
	"strings"
)

var publicCommands = []string{"USER", "PASS"}

func (session *SessionInfo) handleCommand(commandLine string) error {

	command, argument, hasArguments := strings.Cut(commandLine, " ")
	if !hasArguments {
		command = commandLine
	}

	// only allow some commands
	if !session.isLoggedIn && !slices.Contains(publicCommands, command) {
		err := session.Respond(respones.NotLoggedIn())
		if err != nil {
			return fmt.Errorf("sending not logged in response: %s", err)
		}
		return nil
	}

	var err error
	switch command {
	case "USER":
		err = session.handleUSER(argument)
	case "PASS":
		err = session.handlePASS(argument)
	case "LIST":
		err = session.handleLIST(argument)
	case "SYST":
		err = session.handleSYST()
	case "FEAT":
		err = session.handleFEAT()
	case "PWD":
		err = session.handlePWD()
	case "TYPE":
		err = session.handleTYPE(argument)
	case "EPSV":
		log.Printf("Extended passive mode requested")
		dataConn, err := openPassiveDataConnection()
		if err != nil {
			return fmt.Errorf("error opening data connection: %s", err)
		}
		// listener started
		session.dataConnection = dataConn

		log.Printf("Data conneciton listener started, gonna send response")
		// send port to listened on
		err = session.Respond(respones.EPSVEnabled(dataConn.address.Port))
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
	session.isLoggedIn = true
	session.commandSequence = sequences.NONE

	return nil
}

func (session *SessionInfo) handleLIST(argument string) error {
	testList := "-rw-------  1 peter         848 Dec 14 11:22 HELLO_WORLD.txt \r\n"

	log.Printf("starting the wait for data connection")

	dtc := session.dataConnection.getDataConnection()
	log.Printf("data connection is ready")

	// notify client that we will stand sending response
	err := session.Respond(respones.ListSendingResponse())
	if err != nil {
		return err
	}

	err = dtc.write(testList)
	if err != nil {
		return err
	}
	log.Printf("data written to data connection")

	// TODO this isnt always the case
	err = dtc.close()
	if err != nil {
		return err
	}
	// acnowledge that all data was send
	err = session.Respond(respones.ListOk())
	return err
}

func (session *SessionInfo) handleSYST() error {
	err := session.Respond(respones.ServerSystem())
	if err != nil {
		return fmt.Errorf("handling syst: %s", err)
	}

	return nil
}

func (session *SessionInfo) handleFEAT() error {
	features := []string{}

	err := session.Respond(respones.ListFeatures(features))

	if err != nil {
		return fmt.Errorf("handling feat: %s", err)
	}

	return nil
}

func (session *SessionInfo) handlePWD() error {
	err := session.Respond(respones.SendPWD(session.cwd))
	if err != nil {
		return fmt.Errorf("handling pwd: %s", err)
	}

	log.Printf("returned working directory")
	return nil
}

func (session *SessionInfo) handleTYPE(params string) error {
	dataType, dataFormat, hasFormatSet := strings.Cut(params, " ")

	switch dataType {
	case "A":
		session.dataType = TYPE_ASCII
	case "E":
		session.dataType = TYPE_EBCDIC
	case "I":
		session.dataType = TYPE_IMAGE
	case "L":
		session.dataType = TYPE_LOCAL
	default:
		return errors.New("data type not supported")
	}

	if hasFormatSet && (session.dataType == TYPE_ASCII || session.dataType == TYPE_EBCDIC) {
		switch dataFormat {
		case "N":
			session.dataFormat = FORMAT_NON_PRINT
		case "T":
			session.dataFormat = FORMAT_TELNET
		case "C":
			session.dataFormat = FORMAT_ASA
		default:
			return errors.New("data format not supported")
		}
	}

	err := session.Respond(respones.TypeChanged())

	if err != nil {
		return fmt.Errorf("error sending response: %s", err)
	}

	return nil
}
