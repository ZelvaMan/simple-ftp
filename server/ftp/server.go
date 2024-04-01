package ftp

import (
	"fmt"
	"log"
	"net"
)

type FtpServer struct {
	controlConnectionListener net.Listener
	listenAddr                string
	nextConnectionId          int
	sessionClosedChannel      chan int
}

func StartFTPServer(listenAddress string) (*FtpServer, error) {
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %s", listenAddress, err)
	}

	server := &FtpServer{
		controlConnectionListener: listener,
		listenAddr:                listenAddress,
		nextConnectionId:          0,
	}

	go server.handleConnections()

	return server, nil
}

func (server *FtpServer) Stop() error {

	err := server.controlConnectionListener.Close()

	if err != nil {
		return err
	}

	return nil
}

func (server *FtpServer) handleConnections() {
	for {
		newConnection, err := server.controlConnectionListener.Accept()
		if err != nil {
			log.Printf("error accepting control connection: %s", err)
		}

		log.Printf("new connection accepted from %s", newConnection.RemoteAddr().String())

		session, err := createSession(&newConnection)

		// this is a main thread for tcp sessions
		go session.Start()

		if err != nil {
			log.Printf("error starting session: %s", err)
		}

	}
}

func (server *FtpServer) getSessionId() int {
	id := server.nextConnectionId
	server.nextConnectionId = id + 1
	return id
}
