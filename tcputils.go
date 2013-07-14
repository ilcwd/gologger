package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
)

const (
	SOCKET_READ_TIMEOUT = 1e9
)

var WaitForRecordError error = errors.New("Wait for record.")
var ConnCloseError error = errors.New("Connection closed.")

func readAll(conn net.Conn, n int) ([]byte, error) {
	buffer := make([]byte, n)
	read := 0

	var err error
	// Read always returns EOF when ends.
	for read < n && err == nil {
		var thisread int
		thisread, err = conn.Read(buffer[read:])
		read += thisread
	}

	if err == io.EOF {
		switch {
		case read >= n:
			err = nil
		case read == 0:
			err = ConnCloseError
		}
	}

	return buffer[:read], err
}

func readSize(conn net.Conn) (uint32, error) {

	buf, err := readAll(conn, 4)
	if err != nil {
		return 0, err
	}

	return byte2uint32(buf)
}

// Convert a 4-byte array into uint32, using BidEndian
// return 0 if failed.
func byte2uint32(b []byte) (uint32, error) {

	buf := bytes.NewBuffer(b)
	var res uint32
	// network bytes is BigEndian.
	err := binary.Read(buf, binary.BigEndian, &res)
	if err != nil {
		return 0, err
	}

	return res, nil
}

func readRecord(conn net.Conn) (*Record, error) {
	size, err := readSize(conn)
	if err != nil {
		return nil, err
	}
	record, err := readAll(conn, int(size))
	if err != nil {
		return nil, err
	}
	return &Record{DEBUG, record, RECORD_DEFAULT}, nil
}
