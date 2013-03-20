package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

const (
	GO_PROCESS   = 1
	LOGGING_PATH = "/data/logs/gologger/logger.log"
	PROTOCOL     = "tcp"
	ADDR         = "0.0.0.0:30001"
	FLUSH_TIME   = 2.0
	BUFFER_SIZE  = 36 * 1024
)

var (
	logger *FileLogger

	protocol       string
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
	flag.StringVar(&protocol, "P", PROTOCOL, "protocol.")
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

func main() {

	initFlag()
	initLogger()

	runtime.GOMAXPROCS(go_process_num)
	log.Printf("Max go processes set to %d.\n", go_process_num)

	prot := protocol
	switch prot {
	case "tcp":
		tcpLogger()
	case "http":
		httpLogger()
	default:
		log.Fatalf("Unknown protocol %s.\n", prot)
	}

}
