package main

import (
	"io"
	"log"
	"net"
	"os"
	"syscall"
	"time"
)

func tcpLogger() {

	socket, err := net.Listen(PROTOCOL, addr)
	if err != nil {
		log.Fatalf("Cannot bind to %s, error is %s .\n", PROTOCOL, addr, err.Error())
	}
	defer func() {
		if socket != nil {
			socket.Close()
		}
	}()
	log.Printf("Bind to %s.\n", addr)

	// handle signals
	sigc := initSignal()
	go func() {
		for {
			switch <-sigc {
			case syscall.SIGHUP:
				log.Println("A signal HUP catched, flush logger.")
				logger.Flush()

			case syscall.SIGINT:
				fallthrough
			case syscall.SIGTERM:
				log.Println("A signal TERM or INT catched, terminate logger.")
				socket.Close()
				socket = nil
				// TODO: sub connections maybe not close.
				logger.Flush()
				os.Exit(0)
			}
		}

	}()

	// start loop.
	log.Printf("%s Logger starts.\n", PROTOCOL)
	for {
		conn, err := socket.Accept()
		if err != nil {
			log.Printf("Error while accepting: %s.", err)
		}

		go func(conn net.Conn) {
			defer conn.Close()

			for {
				// set max keep alive timeout.
				conn.SetReadDeadline(time.Now().Add(time.Duration(KEEP_ALIVE)))
				record, err := readRecord(conn)
				if err != nil {
					switch {
					case err == ConnCloseError:
						log.Printf("Connection from %s closed.", conn.RemoteAddr())
					case err == io.EOF:
						log.Printf("Connection from %s unexpected closed.", conn.RemoteAddr())
					default:
						log.Printf("Error reading: %s.", err.Error())
					}
					return
				}
				logger.Write(record)
			}
		}(conn)
	}
	// never here
	os.Exit(1)

}
