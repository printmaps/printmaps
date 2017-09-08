/*
Purpose:
- Printmaps purger

Description:
- Remove outdated map directories.

Releases:
- 0.1.0 - 2017/05/27 : beta 1

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

Links:
- http://www.printmaps-osm.de
*/

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// general program info
var (
	progName    = os.Args[0]
	progVersion = "0.1.0"
	progDate    = "2017/05/27"
	progPurpose = "Printmaps Purger"
	progInfo    = "Remove outdated map directories."
)

// Config defines all program settings
type Config struct {
	Logfile  string
	Mapdir   string
	Maturity int
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

	var err error

	configfile := "printmaps_purger.yaml"
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
	log.Printf("config mapdir = %s", config.Mapdir)
	log.Printf("config maturity = %d", config.Maturity)

	pid := os.Getpid()
	log.Printf("own process identifier (pid) = %d", pid)
	log.Printf("shutdown with SIGINT or SIGTERM")

	// start timer trigger
	timerTrigger := time.Tick(time.Second * 60 * 60)

	// start shutdown trigger (subscribe to signals)
	shutdownTrigger := make(chan os.Signal)
	signal.Notify(shutdownTrigger, syscall.SIGINT)  // kill -SIGINT pid -> interrupt
	signal.Notify(shutdownTrigger, syscall.SIGTERM) // kill -SIGTERM pid -> terminated

ForeverLoop:
	for {
		removeOutdatedDirectories(config.Mapdir, config.Maturity)

		// wait for timer or shutdown trigger
		select {
		case <-timerTrigger:
			// unblock select
		case <-shutdownTrigger:
			// initiate shutdown
			break ForeverLoop
		}
	}

	log.Printf("%s (%s) shut down", progName, progPurpose)
}

/*
removeOutdatedDirectories removes outdated map directories
*/
func removeOutdatedDirectories(baseDirectory string, maturity int) {

	files, err := ioutil.ReadDir(baseDirectory)
	if err != nil {
		log.Printf("error <%v> at ioutil.ReadDir(), dir = <%s>", err, baseDirectory)
		return
	}

	now := time.Now()
	for _, file := range files {
		if file.IsDir() {
			diff := now.Sub(file.ModTime())
			if int(diff/time.Hour) >= maturity {
				// safety check to avoid unwanted deletes due to missconfiguration
				// subdirectory name must be conform with the UUID4 naming schema
				if verifyDirectoryName(file.Name()) == false {
					log.Printf("WARNING: map directory <%v> NOT REMOVED (has not the expected UUID4 naming schema)", file.Name())
					continue
				}
				removeDirectory := filepath.Join(baseDirectory, file.Name())
				err := os.RemoveAll(removeDirectory)
				if err != nil {
					log.Printf("error <%v> at os.RemoveAll(), dir = <%s>", err, removeDirectory)
					continue
				}
				log.Printf("outdated (%s) map directory removed: %s", file.ModTime().Format(time.Stamp), file.Name())
			}
		}
	}
}

/*
verifyDirectoryName verifies that the directory has a proper UUID4 naming schema
*/
func verifyDirectoryName(path string) bool {

	var regexUUID4 = "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
	regex := regexp.MustCompile(regexUUID4)

	return regex.MatchString(path)
}
