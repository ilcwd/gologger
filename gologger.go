package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"runtime"
	"time"
)

var logger *FileLogger

func initLogger() {
	runtime.GOMAXPROCS(1)
	var err error
	logger, err = NewFileLogger("./a.out")
	log.SetFlags(log.Llongfile)
	if err != nil {
		log.Fatalf("On creating logger: %s", err.Error())
	}
}

func handleLog(resp http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	body := make([]byte, req.ContentLength)
	io.ReadFull(req.Body, body)
	record := &Record{INFO, body}
	logger.Write(record)
	resp.Write([]byte("ok,\n"))
}

func handleError(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte("Not Found.\n"))
}

func httpLogger() {
	initLogger()
	log.Println("HTTP Logger start.")
	http.HandleFunc("/log", handleLog)
	http.HandleFunc("/", handleError)
	http.ListenAndServe(":10001", nil)
}

func tcpLogger() {
	initLogger()

	protocol := "tcp"
	addr := "0.0.0.0:10001"

	socket, err := net.Listen(protocol, addr)
	if err != nil {
		log.Fatalf("Cannot listen to %s:%s, error is %f", protocol, addr, err.Error())
	}

	defer socket.Close()
	log.Println("TCP Logger start.")

	for {
		conn, err := socket.Accept()
		if err != nil {
			log.Printf("Error while accepting: %s.", err)
		}

		go func(conn net.Conn) {
			defer conn.Close()

			for {
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
				// log.Printf("Receive: %s", record.Msg)
				logger.Write(record)
			}
		}(conn)
	}

}
func main() {
	tcpLogger()
}
