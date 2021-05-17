package main

// Copyright 2021 by tobi@backfrak.de. All
// rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the
// LICENSE file.

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"tobi.backfrak.de/internal/commonbl"
)

// Authors - Information about the authors of the program. You might want to add your name here when contributing to this software
const Authors = "tobi@backfrak.de"

// The version of this program, will be set at compile time by the gradle build script
var version = "undefined"

// Type for functions that can create a response string
type response func(commonbl.PipeHandler, int) error

func main() {
	handleComandlineOptions()
	pipeHandler := *commonbl.NewPipeHandler(params.Test)
	if params.Verbose {
		args := ""
		for _, arg := range os.Args {
			args = fmt.Sprintf("%s %s", args, arg)
		}
		fmt.Fprintln(os.Stdout, fmt.Sprintf("Call: %s", args))
		if !params.PrintVersion {
			printVersion()
		}
		fmt.Fprintln(os.Stdout, fmt.Sprintf("Use named pipe: %s", pipeHandler.GetPipeFilePath()))
	}

	if params.PrintVersion {
		printVersion()
		os.Exit(0)
	}

	if params.Help {
		flag.Usage()
		os.Exit(0)
	}

	// Ensure we exit clean on term and kill signals
	go waitforKillSignalAndExit()
	go waitforTermSignalAndExit()

	// Wait for pipe input and process it in an infinite loop
	for {
		received, errRecv := pipeHandler.WaitForPipeInputString()
		if errRecv != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("Error while receive data from the pipe: %s", errRecv))
			os.Exit(-1)
		}

		var err error = nil
		if strings.HasPrefix(received, commonbl.PROCESS_REQUEST) {
			err = handleRequest(pipeHandler, received, commonbl.PROCESS_REQUEST, processResponse, testProcessResponse)
		} else if strings.HasPrefix(received, commonbl.SERVICE_REQUEST) {
			err = handleRequest(pipeHandler, received, commonbl.SERVICE_REQUEST, serviceResponse, testServiceResponse)
		} else if strings.HasPrefix(received, commonbl.LOCK_REQUEST) {
			err = handleRequest(pipeHandler, received, commonbl.LOCK_REQUEST, lockResponse, testLockResponse)
		}

		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("Error while handle request \"%s\": %s", received, err))
			os.Exit(-2)
		}

		time.Sleep(time.Millisecond)
	}

}

func handleRequest(handler commonbl.PipeHandler, request string, requestType string, productiveFunc response, testFunc response) error {
	id, errConv := getIdFromRequest(request)
	if errConv != nil {
		return nil // In case we cant find an ID, we simply ingnor the request as any other invalid input
	}
	if params.Verbose {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("Handle \"%s\" with id %d", requestType, id))
	}

	var writeErr error
	if !params.Test {
		writeErr = productiveFunc(handler, id)
	} else {
		writeErr = testFunc(handler, id)
	}
	if writeErr != nil {
		return writeErr
	}

	return nil
}

func lockResponse(handler commonbl.PipeHandler, id int) error {
	fmt.Fprintln(os.Stderr, fmt.Sprintf("Error: Productive code not implemented yet"))

	return &strconv.NumError{}
}

func serviceResponse(handler commonbl.PipeHandler, id int) error {
	fmt.Fprintln(os.Stderr, fmt.Sprintf("Error: Productive code not implemented yet"))

	return &strconv.NumError{}
}

func processResponse(handler commonbl.PipeHandler, id int) error {
	fmt.Fprintln(os.Stderr, fmt.Sprintf("Error: Productive code not implemented yet"))

	return &strconv.NumError{}
}

func testProcessResponse(handler commonbl.PipeHandler, id int) error {
	return handler.WritePipeString(fmt.Sprintf("%s Test Response for request %d", commonbl.PROCESS_REQUEST, id))
}

func testServiceResponse(handler commonbl.PipeHandler, id int) error {
	return handler.WritePipeString(fmt.Sprintf("%s Test Response for request %d", commonbl.SERVICE_REQUEST, id))
}

func testLockResponse(handler commonbl.PipeHandler, id int) error {
	return handler.WritePipeString(fmt.Sprintf("%s Test Response for request %d", commonbl.LOCK_REQUEST, id))
}

func getIdFromRequest(request string) (int, error) {
	splitted := strings.Split(request, ":")

	if len(splitted) != 2 {
		return 0, &strconv.NumError{}
	}

	idStr := strings.TrimSpace(splitted[1])
	id, errConv := strconv.Atoi(idStr)
	if errConv != nil {
		return 0, &strconv.NumError{}
	}

	return id, nil
}

func waitforKillSignalAndExit() {
	killSignal := make(chan os.Signal, syscall.SIGKILL)
	signal.Notify(killSignal, os.Interrupt)
	<-killSignal

	if params.Verbose {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("End: %s due to kill signal", os.Args[0]))
	}
	os.Exit(0)
}

func waitforTermSignalAndExit() {
	termSignal := make(chan os.Signal, syscall.SIGTERM)
	signal.Notify(termSignal, os.Interrupt)
	<-termSignal
	if params.Verbose {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("End: %s due to terminate signal", os.Args[0]))
	}
	os.Exit(0)
}

// Prints the version string
func printVersion() {
	fmt.Fprintln(os.Stdout, getVersion())
}

// Get the version string
func getVersion() string {
	return fmt.Sprintf("Version: %s", version)
}
