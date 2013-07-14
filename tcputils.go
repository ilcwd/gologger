package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strings"
)

const (
	SOCKET_READ_TIMEOUT = 1e9

	NAMED_MAGIC_CODE_SIZE = 2
	NAMED_SIZE            = 1
	MAGIC_CODE            = "\x00\x08"
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

func readUInt32(conn net.Conn) (uint32, error) {

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

func readNamedRecord(conn net.Conn) (*Record, error) {
	name_string := RECORD_DEFAULT
	is_named_record := false

	size, err := readUInt32(conn)
	if err != nil {
		return nil, err
	}

	// 2 - bytes for magic cocde
	magic_bytes, err := readAll(conn, NAMED_MAGIC_CODE_SIZE)
	if err != nil {
		return nil, err
	}

	// magic code for compatiblity
	if magic_bytes[0] == byte(0x00) && magic_bytes[1] == byte(0x08) {
		// a named record
		one_byte_array, err := readAll(conn, NAMED_SIZE)
		if err != nil {
			return nil, err
		}
		name_size := int(one_byte_array[0])
		bytes_name, err := readAll(conn, name_size)
		if err != nil {
			return nil, err
		}
		name_string = strings.ToLower(string(bytes_name))
		size -= uint32(NAMED_MAGIC_CODE_SIZE + NAMED_SIZE + name_size)
		is_named_record = true

	} else {
		size -= NAMED_MAGIC_CODE_SIZE
	}

	record_bytes, err := readAll(conn, int(size))
	if err != nil {
		return nil, err
	}
	if !is_named_record {
		buf := bytes.NewBuffer(magic_bytes)
		buf.Write(record_bytes)
		// record_bytes = bytes.Join([][]byte{magic_bytes, record_bytes}, []byte(""))
		return &Record{DEBUG, buf.Bytes(), name_string}, nil
	}
	return &Record{DEBUG, record_bytes, name_string}, nil
}
