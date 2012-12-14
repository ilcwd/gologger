package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

const (
	GO_PROCESS   = 1
	LOGGING_PATH = "/data/logs/gologger/logger.log"
	PROTOCOL     = "tcp"
	ADDR         = "0.0.0.0:30001"
	FLUSH_TIME   = 1.0
	BUFFER_SIZE  = 36 * 1024
)

var (
	logger *FileLogger

	go_process_num int
	logging_path   string
	addr           string
	buffer_size    int
	flush_time     float64
)

func initFlag() {
	flag.IntVar(&go_process_num, "c", GO_PROCESS, "max go processes number.")
	flag.Float64Var(&flush_time, "f", FLUSH_TIME, "flush time in second.")
	flag.IntVar(&buffer_size, "b", BUFFER_SIZE, "buffer size in byte.")
	flag.StringVar(&logging_path, "p", LOGGING_PATH, "log file to write.")
	flag.StringVar(&addr, "h", ADDR, "bind address.")
	flag.Parse()
}

func initLogger() {
	// self logger
	log.SetFlags(log.Ldate | log.Ltime)

	// file logger
	var err error
	logger, err = NewFileLogger(logging_path, buffer_size, flush_time)
	if err != nil {
		log.Fatalf("On creating logger: %s", err.Error())
	}
	log.Printf("Logger buffer size %d bytes.\n", buffer_size)
	log.Printf("Logger flush time %f seconds.\n", flush_time)
}

func initSignal() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGUSR1)
	return c
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

	initFlag()
	initLogger()

	runtime.GOMAXPROCS(go_process_num)
	log.Printf("Max go processes set to %d.\n", go_process_num)

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
			case syscall.SIGINT:
				log.Println("A signal INT catched, flush logger.")
				logger.Flush()
			case syscall.SIGTERM:
				log.Println("A signal TERM catched, terminate logger.")
				socket.Close()
				socket = nil
				// TODO: sub connections maybe not close.
				logger.Flush()
				os.Exit(0)
			}
		}

	}()

	log.Println("TCP Logger start.")

	// start loop.
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
				logger.Write(record)
			}
		}(conn)
	}
	// never here
	os.Exit(1)

}

func main() {
	tcpLogger()
}
