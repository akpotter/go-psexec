package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"

	"github.com/golang-devops/go-psexec/client"
	"github.com/golang-devops/go-psexec/shared"
)

var (
	serverFlag    = flag.String("server", "http://localhost:62677", "The endpoint server address")
	executorFlag  = flag.String("executor", "winshell", "The executor to use")
	clientPemFlag = flag.String("client_pem", "client.pem", "The file path for the client pem (private+public) key file")
)

func handleRecovery() {
	if r := recover(); r != nil {
		log.Printf("ERROR: %s\n", getErrorStringFromRecovery(r))
	}
}

func main() {
	defer handleRecovery()

	flag.Parse()

	exeAndArgs := flag.Args()
	if len(exeAndArgs) == 0 {
		log.Fatal("Need at least one additional argument")
	}

	var exe string
	var args []string = []string{}

	exe = exeAndArgs[0]
	if len(exeAndArgs) > 1 {
		args = exeAndArgs[1:]
	}

	pvtKey, err := shared.ReadPemKey(*clientPemFlag)
	if err != nil {
		log.Fatalf("Cannot read client pem file, error: %s", err.Error())
	}

	c := client.New(pvtKey)

	session, err := c.RequestNewSession(*serverFlag)
	if err != nil {
		log.Fatalf("Unable to create new session, error: %s", err.Error())
	}

	fmt.Printf("Using session id: %d\n", session.SessionId())

	resp, err := session.StartExecRequest(&shared.ExecDto{*executorFlag, exe, args})
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fmt.Println(scanner.Text())

		/*cipher := scanner.Bytes()
		  plaintextBytes, err := shared.DecryptSymmetric(session.SessionToken(), cipher)
		  if err != nil {
		      log.Fatalf("Unable read encrypted server response, error: %s", err.Error())
		  }
		  fmt.Println(string(plaintextBytes))
		*/
	}
}