package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

var logFile *os.File

func InitLogger() {
	os.RemoveAll("/tmp/gfm.log")
	file, err := os.OpenFile("/tmp/gfm.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open log file")
		os.Exit(1)
	}

	logFile = file
	log.SetOutput(io.Writer(file))
}

func CloseLogFile() {
	logFile.Close()
}

func stopLogger() {
	CloseLogFile()
	log.SetOutput(os.Stdout)
}
