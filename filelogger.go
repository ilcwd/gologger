package main

import (
	"bufio"
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
)

type Record struct {
	Level int
	Msg   []byte
}

type FileLogger struct {
	path       string
	lastRotate time.Time
	file       *bufio.Writer
	realfile   *os.File
	record     chan *Record
	flush      chan bool
	flushLevel int
}

// Create file logger
// Param buffer - max records in byte stored in memory.
// Param flush_time - interval in second to flush records to file.
func NewFileLogger(path string, buffer int, flush_time float64) (*FileLogger, error) {
	file, realfile, err := openfile(path)
	if err != nil {
		return nil, err
	}
	now := time.Now()

	filelogger := FileLogger{
		path,
		now,
		file,
		realfile,
		make(chan *Record, BUFFER_SIZE),
		make(chan bool),
		FLUSH_LEVEL,
	}

	go func() {
		defer func() {
			filelogger.file.Flush()
			filelogger.realfile.Close()
		}()
		ticker := time.NewTicker(time.Duration(flush_time * float64(time.Second)))
		for {
			select {
			case <-filelogger.flush:
				filelogger.file.Flush()
			// time out flush
			case <-ticker.C:
				filelogger.file.Flush()
			case newrecord, ok := <-filelogger.record:
				if !ok {
					return
				}
				filelogger.file.Write(newrecord.Msg)
				// newline := []byte{0x0a}
				// filelogger.file.Write(newline) // write a new line.
				if newrecord.Level > filelogger.flushLevel {
					filelogger.Flush()
				}
				// force to flush (by user)
			}
		}
	}()

	return &filelogger, nil
}

func openfile(path string) (*bufio.Writer, *os.File, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 666)
	if err != nil {
		return nil, nil, err
	}
	wr := bufio.NewWriterSize(file, BUFFER_SIZE)
	return wr, file, nil
}

// rotate hourly
func (f *FileLogger) shouldRotate(now time.Time) bool {
	return f.lastRotate.Hour() != now.Hour()
}

func (f *FileLogger) doRotate(now time.Time) {
	srcname := f.path
	dstname := f.path + now.Format(".2006-01-02_15")
	err := os.Rename(srcname, dstname)
	if err != nil {
		return
	}

	file, realfile, err := openfile(f.path)
	if err != nil {
		return
	}
	f.file = file
	f.realfile = realfile
	f.lastRotate = now
}

func (f *FileLogger) Flush() {
	f.flush <- true
}

func (f *FileLogger) Write(record *Record) {
	f.record <- record
}

func (f *FileLogger) Close() {
	close(f.record)
}
