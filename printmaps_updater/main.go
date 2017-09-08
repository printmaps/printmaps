/*
Purpose:
- updates the printmaps database

Description:
- Wrapper for the osm database update processes.

Releases:
- 0.1.0 - 2017/05/23 : beta 1
- 0.2.0 - 2017/06/28 : beta 2
- 0.2.1 - 2017/08/16 : mail message improved

Author:
- Klaus Tockloth

Copyright and license:
- Copyright (c) 2017 Klaus Tockloth
- MIT license

Permission is hereby granted, free of charge, to any person obtaining a copy of this software
and associated documentation files (the Software), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

The software is provided 'as is', without warranty of any kind, express or implied, including
but not limited to the warranties of merchantability, fitness for a particular purpose and
noninfringement. In no event shall the authors or copyright holders be liable for any claim,
damages or other liability, whether in an action of contract, tort or otherwise, arising from,
out of or in connection with the software or the use or other dealings in the software.

Contact (eMail):
- printmaps.service@gmail.com

Remarks:
- Cross compilation for Linux: env GOOS=linux GOARCH=amd64 go build -v
- Usage: nohup ./printmaps_updater 1>printmaps_updater.out 2>&1 &

ToDo:
- eMail notification

Links:
- http://www.printmaps-osm.de
*/

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/scorredoira/email"

	yaml "gopkg.in/yaml.v2"
)

// general program info
var (
	progName    = os.Args[0]
	progVersion = "0.2.1"
	progDate    = "2017/08/16"
	progPurpose = "updates the printmaps database"
	progInfo    = "This program is a wrapper for the osm database update processes."
)

// Config defines all program settings
type Config struct {
	Logfile        string
	Workdir        string
	Preprocess     string
	Updateprocess  string
	Removeprocess  string
	Schedulehour   int
	Scheduleminute int
	Fromname       string
	Fromaddress    string
	Toaddress      string
	Authidentity   string
	Authusername   string
	Authpassword   string
	Authhost       string
}

var config Config

/*
init initializes this program
*/
func init() {

	// initialize logger
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}

/*
main starts this program
*/
func main() {

	configfile := "printmaps_updater.yaml"
	if len(os.Args) > 1 {
		configfile = os.Args[1]
	}
	source, err := ioutil.ReadFile(configfile)
	if err != nil {
		log.Fatalf("fatal error <%v> at ioutil.ReadFile(), file = <%s>", err, configfile)
	}

	if err = yaml.Unmarshal(source, &config); err != nil {
		log.Fatalf("fatal error <%v> at yaml.Unmarshal()", err)
	}

	logfile, err := os.OpenFile(config.Logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("fatal error <%v> at os.OpenFile(), file = <%v>", err, config.Logfile)
	}
	defer logfile.Close()
	log.SetOutput(logfile)

	// print start message to stdout
	fmt.Printf("%s (%s) startet. See '%s' for further details.\n", progName, progPurpose, config.Logfile)

	// log program start details
	log.Printf("name = %s", progName)
	log.Printf("release = %s - %s", progVersion, progDate)
	log.Printf("purpose = %s", progPurpose)
	log.Printf("info = %s", progInfo)

	log.Printf("config logfile = %s", config.Logfile)
	log.Printf("config workdir = %s", config.Workdir)
	log.Printf("config preprocess = %s", config.Preprocess)
	log.Printf("config updateprocess = %s", config.Updateprocess)
	log.Printf("config removeprocess = %s", config.Removeprocess)
	log.Printf("config schedulehour = %d", config.Schedulehour)
	log.Printf("config scheduleminute = %d", config.Scheduleminute)
	log.Printf("config fromname  = %s", config.Fromname)
	log.Printf("config fromaddress = %s", config.Fromaddress)
	log.Printf("config toaddress = %s", config.Toaddress)
	log.Printf("config authidentity = %s", config.Authidentity)
	log.Printf("config authusername = %s", config.Authusername)
	log.Printf("config authpassword = %s", config.Authpassword)
	log.Printf("config authhost = %s", config.Authhost)

	// change into working directory
	if err = os.Chdir(config.Workdir); err != nil {
		log.Fatalf("fatal error <%v> at os.Chdir(), dir = <%s>", err, config.Workdir)
	}

	// regular (single) update : true,nil ... false,nil ... wait
	// queued (multiple) updates : true,nil ... true,nil ... true,nil ... false,nil ... wait
	for {
		update, err := updateData()
		log.Printf("========== updateData(), update = <%t>, err = <%v> ==========", update, err)
		if err != nil {
			log.Printf("error <%v> at updateData()", err)
		}

		if update {
			// try, if more work has to be done
			continue
		}

		// wait till next wake-up
		waitTillWakeup()
	}
}

/*
runCommand runs a command / program
*/
func runCommand(command string) (commandExitStatus int, commandOutput []byte, err error) {

	program := "/bin/bash"
	args := []string{"-c", command}
	cmd := exec.Command(program, args...)

	commandOutput, err = cmd.CombinedOutput()

	var waitStatus syscall.WaitStatus
	if err != nil {
		// command was not successful
		log.Printf("error <%v> at cmd.CombinedOutput()", err)
		log.Printf("command (not successful) = <%s>", strings.Join(cmd.Args, " "))
		if exitError, ok := err.(*exec.ExitError); ok {
			// command fails because of an unsuccessful exit code
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			log.Printf("command exit code = <%d>", waitStatus.ExitStatus())
		}
		if len(commandOutput) > 0 {
			log.Printf("command output (stdout, stderr) =\n%s", string(commandOutput))
		}

		// notify failure via email (and terminate the program)
		notification := fmt.Sprintf("command (not successful) = <%s>\n\ncommand output (stdout, stderr) =\n%s",
			strings.Join(cmd.Args, " "), string(commandOutput))
		notifyError(notification)
	} else {
		// command was successful
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
		log.Printf("command (successful) = <%s>", strings.Join(cmd.Args, " "))
		log.Printf("command exit code = <%d>", waitStatus.ExitStatus())
		if len(commandOutput) > 0 {
			log.Printf("command output (stdout, stderr) =\n%s", string(commandOutput))
		}
	}

	commandExitStatus = waitStatus.ExitStatus()
	return
}

/*
updateData updates the OSM database data
the content of the state file looks like this (example):
#Mon May 22 13:11:13 CEST 2017
sequenceNumber=1177
timestamp=2017-05-21T20\:43\:59Z
*/
func updateData() (update bool, err error) {

	// read old state file
	log.Printf("---------- state file (state.txt) ----------")
	filename := "state.txt"
	linesBefore, err := slurpFile(filename)
	if err != nil {
		message := fmt.Sprintf("error <%v> at slurpFile(); file = <%v>", err, filename)
		return false, errors.New(message)
	}
	log.Printf("Content of <%s> before running the preprocess:", filename)
	for index, line := range linesBefore {
		log.Printf("%d %s", index, line)
	}

	// download new state file and new change file
	log.Printf("---------- preprocess state and change file ----------")
	exitStatus, _, err := runCommand(config.Preprocess)
	if err != nil {
		message := fmt.Sprintf("error <%v> at runCommand(), exitStatus = <%v>, command = <%v>", err, exitStatus, config.Preprocess)
		return false, errors.New(message)
	}

	// read new state file
	log.Printf("---------- state file (state.txt) ----------")
	linesAfter, err := slurpFile(filename)
	if err != nil {
		message := fmt.Sprintf("error <%v> at slurpFile(); file = <%v>", err, filename)
		return false, errors.New(message)
	}
	log.Printf("Content of <%s> after running the preprocess:", filename)
	for index, line := range linesAfter {
		log.Printf("%d %s", index, line)
	}

	// contains the state file a new sequence number?
	newStateFile := true
	if linesBefore[1] == linesAfter[1] {
		log.Printf("no new changes available ... nothing to do")
		newStateFile = false
	}

	if newStateFile {
		log.Printf("---------- update database with change file content ----------")
		exitStatus, _, err = runCommand(config.Updateprocess)
		if err != nil {
			message := fmt.Sprintf("error <%v> at runCommand(), exitStatus = <%v>, command = <%v>", err, exitStatus, config.Updateprocess)
			return false, errors.New(message)
		}
	}

	log.Printf("---------- remove change file ----------")
	exitStatus, _, err = runCommand(config.Removeprocess)
	if err != nil {
		message := fmt.Sprintf("error <%v> at runCommand(), exitStatus = <%v>, command = <%v>", err, exitStatus, config.Removeprocess)
		return false, errors.New(message)
	}

	if newStateFile {
		return true, nil
	}
	return false, nil
}

/*
waitTillWakeup waits until the next scheduled start time
*/
func waitTillWakeup() {

	// calculate next wakeup/start time
	now := time.Now()
	wakeup := time.Date(now.Year(), now.Month(), now.Day(), config.Schedulehour, config.Scheduleminute, 0, 0, time.Local)

	if wakeup.Before(now) {
		// add one day
		wakeup = time.Date(now.Year(), now.Month(), now.Day()+1, config.Schedulehour, config.Scheduleminute, 0, 0, time.Local)
	}
	log.Printf("next start time = %s\n", wakeup)

	for {
		now = time.Now()
		log.Printf("waiting for next start time ...\n")
		if now.After(wakeup) {
			log.Printf("wakeup reached ...\n")
			break
		}
		time.Sleep(15 * time.Minute)
	}
}

/*
slurpFile slurps all lines of a text file into a slice of strings
*/
func slurpFile(filename string) ([]string, error) {

	var lines []string

	file, err := os.Open(filename)
	if err != nil {
		return lines, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	return lines, nil
}

/*
notifyError notifies (via email) about an error situation and terminates the program
*/
func notifyError(notification string) {

	subject := "Printmaps-Updater failed"
	text := `Next steps:
- 1. see log file for further details
- 2. correct the defect / misbehavior
- 3. decrement sequenceNumber in state.txt
- 4. decrement timestamp in state.txt
- 5. restart printmaps updater

Remarks:
The updater treats the sequenceNumber as a successfully applied update,
and increments the sequenceNumber on startup (if updates are available).

Example:
- sequenceNumber 1611 (with timestamp 2017-08-15T20\:44\:02Z) failed
- patch sequenceNumber to 1610
- patch timestamp to 2017-08-14T20\:44\:02Z
- restart printmaps updater
`

	body := text + "\n" + notification
	notificationMail := email.NewMessage(subject, body)
	notificationMail.From = mail.Address{Name: config.Fromname, Address: config.Fromaddress}
	notificationMail.To = []string{config.Toaddress}

	auth := smtp.PlainAuth(config.Authidentity, config.Authusername, config.Authpassword, config.Authhost)
	if err := email.Send(config.Authhost+":587", auth, notificationMail); err != nil {
		log.Printf("error <%v> at email.Send()", err)
	} else {
		log.Printf("error notification send to <%s>", config.Toaddress)
	}

	os.Exit(1)
}
