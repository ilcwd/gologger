package main

import (
	"io"
	"log"
	"net/http"
)

func handleLog(resp http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	body := make([]byte, req.ContentLength)
	io.ReadFull(req.Body, body)
	record := &Record{INFO, body, RECORD_DEFAULT}
	loggermapping[RECORD_DEFAULT].Write(record)
	resp.Write([]byte("ok,\n"))
}

func handleError(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte("Not Found.\n"))
}

func httpLogger() {
	log.Printf("Serves at %s .", addr)
	http.HandleFunc("/log", handleLog)
	http.HandleFunc("/", handleError)
	http.ListenAndServe(addr, nil)
}
