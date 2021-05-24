package main

// Copyright 2021 by tobi@backfrak.de. All
// rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the
// LICENSE file.

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
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

// The logger for this programm
var logger commonbl.Logger

// Path to the smbstatus executable
var smbstatusPath string

func main() {
	handleComandlineOptions()
	requestHandler := *commonbl.NewPipeHandler(params.Test, commonbl.RequestPipe)
	responseHandler := *commonbl.NewPipeHandler(params.Test, commonbl.ResposePipe)
	logger = *commonbl.NewLogger(params.Verbose)
	if params.Verbose {
		args := ""
		for _, arg := range os.Args {
			args = fmt.Sprintf("%s %s", args, arg)
		}
		fmt.Fprintln(os.Stdout, fmt.Sprintf("Call: %s", args))
		if !params.PrintVersion {
			printVersion()
		}
	}
	logger.WriteVerbose(fmt.Sprintf("Named pipe for requests: %s", requestHandler.GetPipeFilePath()))
	logger.WriteVerbose(fmt.Sprintf("Named pipe for response: %s", responseHandler.GetPipeFilePath()))

	if params.PrintVersion {
		printVersion()
		os.Exit(0)
	}

	if params.Help {
		flag.Usage()
		os.Exit(0)
	}

	if !params.Test {
		var errLookPath error
		smbstatusPath, errLookPath = exec.LookPath("smbstatus")
		if errLookPath != nil {
			logger.WriteErrorMessage("Can not find \"smbstatus\" executable. Please install the needed package.")
			os.Exit(-3)
		} else {
			logger.WriteVerbose(fmt.Sprintf("Use %s to get samba status.", smbstatusPath))
		}
	}

	// Ensure we exit clean on term and kill signals
	go waitforKillSignalAndExit()
	go waitforTermSignalAndExit()

	// Wait for pipe input and process it in an infinite loop
	logger.WriteInformation(fmt.Sprintf("Started %s, waiting for requests in pipe", os.Args[0]))
	for {
		logger.WriteVerbose(fmt.Sprintf("Wait for requests in: %s", requestHandler.GetPipeFilePath()))
		received, errRecv := requestHandler.WaitForPipeInputString()
		if errRecv != nil {
			logger.WriteErrorMessage(fmt.Sprintf("Receive data from the pipe: %s", errRecv))
			os.Exit(-1)
		}

		var err error = nil
		if strings.HasPrefix(received, string(commonbl.PROCESS_REQUEST)) {
			err = handleRequest(responseHandler, received, commonbl.PROCESS_REQUEST, processResponse, testProcessResponse)
		} else if strings.HasPrefix(received, string(commonbl.SHARE_REQUEST)) {
			err = handleRequest(responseHandler, received, commonbl.SHARE_REQUEST, shareResponse, testShareResponse)
		} else if strings.HasPrefix(received, string(commonbl.LOCK_REQUEST)) {
			err = handleRequest(responseHandler, received, commonbl.LOCK_REQUEST, lockResponse, testLockResponse)
		}

		if err != nil {
			logger.WriteErrorMessage(fmt.Sprintf("Handle request \"%s\"\n\n: %s", received, err))
			os.Exit(-2)
		}

		time.Sleep(time.Millisecond)
	}

}

func handleRequest(handler commonbl.PipeHandler, request string, requestType commonbl.RequestType, productiveFunc response, testFunc response) error {
	id, errConv := commonbl.GetIdFromRequest(request)
	if errConv != nil {
		return nil // In case we cant find an ID, we simply ingnor the request as any other invalid input
	}
	logger.WriteVerbose(fmt.Sprintf("Handle \"%s\" with id %d", requestType, id))

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
	header := commonbl.GetTestResponseHeader(commonbl.LOCK_REQUEST, id)
	data, err := exec.Command(smbstatusPath, "-L", "-n").Output()
	if err != nil {
		logger.WriteErrorMessage(fmt.Sprintf("\"%s -L -n\"  returned the following error: %s", smbstatusPath, err))
		os.Exit(-4)
	}
	response := commonbl.GetResponse(header, string(data))

	return handler.WritePipeString(response)
}

func shareResponse(handler commonbl.PipeHandler, id int) error {
	header := commonbl.GetTestResponseHeader(commonbl.SHARE_REQUEST, id)
	data, err := exec.Command(smbstatusPath, "-S", "-n").Output()
	if err != nil {
		logger.WriteErrorMessage(fmt.Sprintf("\"%s -S -n\"  returned the following error: %s", smbstatusPath, err))
		os.Exit(-4)
	}
	response := commonbl.GetResponse(header, string(data))

	return handler.WritePipeString(response)
}

func processResponse(handler commonbl.PipeHandler, id int) error {
	header := commonbl.GetTestResponseHeader(commonbl.PROCESS_REQUEST, id)
	data, err := exec.Command(smbstatusPath, "-p", "-n").Output()
	if err != nil {
		logger.WriteErrorMessage(fmt.Sprintf("\"%s -p -n\"  returned the following error: %s", smbstatusPath, err))
		os.Exit(-4)
	}
	response := commonbl.GetResponse(header, string(data))

	return handler.WritePipeString(response)
}

func testProcessResponse(handler commonbl.PipeHandler, id int) error {
	header := commonbl.GetTestResponseHeader(commonbl.PROCESS_REQUEST, id)
	response := commonbl.GetResponse(header, commonbl.TestProcessResponse)

	return handler.WritePipeString(response)
}

func testShareResponse(handler commonbl.PipeHandler, id int) error {
	header := commonbl.GetTestResponseHeader(commonbl.SHARE_REQUEST, id)
	response := commonbl.GetResponse(header, commonbl.TestShareResponse)

	return handler.WritePipeString(response)
}

func testLockResponse(handler commonbl.PipeHandler, id int) error {
	header := commonbl.GetTestResponseHeader(commonbl.LOCK_REQUEST, id)
	response := commonbl.GetResponse(header, commonbl.TestLockResponse)

	return handler.WritePipeString(response)
}

func waitforKillSignalAndExit() {
	killSignal := make(chan os.Signal, syscall.SIGKILL)
	signal.Notify(killSignal, os.Interrupt)
	<-killSignal

	logger.WriteInformation(fmt.Sprintf("End %s due to kill signal", os.Args[0]))

	os.Exit(0)
}

func waitforTermSignalAndExit() {
	termSignal := make(chan os.Signal, syscall.SIGTERM)
	signal.Notify(termSignal, os.Interrupt)
	<-termSignal

	logger.WriteInformation(fmt.Sprintf("End %s due to terminate signal", os.Args[0]))

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
