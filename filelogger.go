package main

import (
	"bufio"
	"log"
	"os"
	"time"
)

const (
	NOTSET   = 0
	DEBUG    = 10
	INFO     = 20
	WARNING  = 30
	ERROR    = 40
	CRITICAL = 50
)

const (
	FLUSH_LEVEL = ERROR
	CHAN_SIZE   = 100
)

type Record struct {
	Level int
	Msg   []byte
}

type FileLogger struct {
	path        string
	lastRotate  time.Time
	file        *bufio.Writer
	realfile    *os.File
	record      chan *Record
	flush       chan bool
	flushLevel  int
	buffer_size int
}

// Create file logger
// Param buffer - max records in byte stored in memory.
// Param flush_time - interval in second to flush records to file.
func NewFileLogger(path string, buffer int, flush_time float64) (*FileLogger, error) {
	file, realfile, err := openFile(path, buffer)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	filelogger := FileLogger{
		path,
		now,
		file,
		realfile,
		make(chan *Record, CHAN_SIZE),
		make(chan bool, 1),
		FLUSH_LEVEL,
		buffer,
	}

	go func() {
		defer func() {
			filelogger.file.Flush()
			filelogger.realfile.Close()
		}()

		// time out flush and rotate
		go func() {
			ticker := time.NewTicker(time.Duration(flush_time * float64(time.Second)))
			for {
				<-ticker.C
				filelogger.Flush()
			}
		}()

		// loop for writing log.
		for {
			select {
			// flush manually
			case <-filelogger.flush:
				filelogger.file.Flush()

			// write log.
			case newrecord, ok := <-filelogger.record:
				if !ok { // chan closed, end function.
					return
				}
				filelogger.file.Write(newrecord.Msg)
				newline := []byte{0x0a}
				filelogger.file.Write(newline) // write a new line.
				filelogger.hourlyRotate()
				if newrecord.Level > filelogger.flushLevel {
					filelogger.Flush()
				}
			}
		}
	}()

	return &filelogger, nil
}

func openFile(path string, buffer int) (*bufio.Writer, *os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
	if err != nil {
		return nil, nil, err
	}
	wr := bufio.NewWriterSize(file, buffer)
	return wr, file, nil
}

// if one hour elapsed, rotate the file.
func (f *FileLogger) hourlyRotate() {

	now := time.Now()
	if f.lastRotate.Hour() == now.Hour() {
		return
	}

	srcname := f.path
	dstname := f.path + f.lastRotate.Format(".2006-01-02_15")
	err := os.Rename(srcname, dstname)
	if err != nil {
		log.Printf("Error on rename log file: %s.\n", err.Error())
		return
	}

	file, realfile, err := openFile(f.path, f.buffer_size)
	if err != nil {
		log.Printf("Error on opening new log file: %s.\n", err.Error())
		return
	}

	// flush and close old fd.
	f.file.Flush()
	f.file = nil
	f.realfile.Close()
	f.realfile = nil

	// assigned to new file.
	f.file = file
	f.realfile = realfile
	f.lastRotate = now

	log.Printf("INFO: Rotate file to %s.\n", dstname)
}

func (f *FileLogger) Flush() {
	f.flush <- true
}

// a buffered write function
func (f *FileLogger) Write(record *Record) {
	f.record <- record
}

func (f *FileLogger) Close() {
	close(f.record)
}
