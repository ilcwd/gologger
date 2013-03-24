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
	ROTATE_TYPE  = "D"
	LOGGING_NAME = "Default"
)

var (
	loggermapping map[string]*FileLogger

	protocol       string
	go_process_num int
	logging_path   string
	addr           string
	config_path    string
)

func initFlag() {
	flag.IntVar(&go_process_num, "c", GO_PROCESS, "max go processes number.")
	flag.StringVar(&addr, "h", ADDR, "bind address.")
	flag.StringVar(&protocol, "P", PROTOCOL, "protocol.")
	flag.StringVar(&config_path, "C", "", "JSON config file.")
	flag.Parse()
}

func initLogger() {
	// self logger
	log.SetFlags(log.Ldate | log.Ltime)

	// path -> filelogger
	path_to_logger := make(map[string]*FileLogger)
	// initial loggers mapping
	loggermapping = make(map[string]*FileLogger)

	config_items, err := LoadJSONConfig(config_path)
	if err != nil {
		log.Fatalf("[ERROR] Loading config = %s failed, err = %s", config_path, err.Error())
	}
	if len(config_items) == 0 {
		log.Fatalf("[ERROR] Empty config, path = %s.\n", config_path)
	}

	var logger *FileLogger
	var ok bool
	// initial logger
	for _, item := range config_items {
		// file logger
		log.Printf("Initialing logger = %s.\n", item.LoggingName)

		logger, ok = path_to_logger[item.LoggingPath]
		if ok {
			// if two loggers share the same file, they share the same logger
			log.Printf("Logger %s is exists.\n", item.LoggingName)
			goto FINAL
		}
		logger, err = NewFileLogger(
			item.LoggingPath,
			item.BufferSize,
			item.FlushTime,
			item.RotateType,
		)
		if err != nil {
			log.Fatalf("[ERROR] On creating logger = %v , err = %s", item, err.Error())
		}
		log.Printf("Logger %s buffer size %d bytes.\n",
			item.LoggingName, item.BufferSize)
		log.Printf("Logger %s flush time %f seconds.\n",
			item.LoggingName, item.FlushTime)

		path_to_logger[item.LoggingPath] = logger

	FINAL:
		loggermapping[item.LoggingName] = logger
	}

	// if the "default" logger is not presented, raise error.
	_, ok = loggermapping[RECORD_DEFAULT]
	if !ok {
		log.Fatalf("[ERROR] Logger default is not set.\n")
	}

}

func initSignal() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGUSR1)
	return c
}

func flushLogger() {
	for _, logger := range loggermapping {
		logger.Flush()
	}
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
		log.Fatalf("[ERROR] Unknown protocol %s.\n", prot)
	}

}
