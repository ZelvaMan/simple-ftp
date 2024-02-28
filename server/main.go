package main

import (
	"log"
	"os"
	"os/signal"
	"server/ftp"
	"syscall"
)

const address = ":21"

func main() {
	cancelChan := make(chan os.Signal, 1)
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	log.Print("Simple ftp server")
	log.Print("Starting...")
	_, err := ftp.StartFTPServer(address)
	log.Printf("Server is ready to accept connection on %s",
		address)
	if err != nil {
		log.Printf("Error starting ftp server: %s", err)
	}

	sig := <-cancelChan
	log.Printf("Caught signal %v", sig)
}
