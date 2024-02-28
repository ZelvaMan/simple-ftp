package ftp

import (
	"fmt"
	"log"
	"net"
)

type FtpServer struct {
	controlConnectionListener net.Listener
	sessions                  map[int]*SessionInfo //id is used for closing of channels
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
		sessions:                  map[int]*SessionInfo{},
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
		go session.Start()
		if err != nil {
			log.Printf("error starting session: %s", err)
		}

		sessionId := server.getSessionId()
		server.sessions[sessionId] = session
	}
}

func (server *FtpServer) getSessionId() int {
	id := server.nextConnectionId
	server.nextConnectionId = id + 1
	return id
}
