/*
Purpose:
- Printmaps build service

Description:
- Build service to build large printable maps.

Releases:
- 0.1.0 - 2017/05/23 : initial release (beta version)
- 0.1.1 - 2017/05/26 : security enhancements
- 0.1.2 - 2017/08/08 : more output for unsuccessful commands
- 0.1.3 - 2017/08/10 : bug concerning parallel Chdir() fixed
- 0.1.4 - 2017/08/17 : workaround for rare failure of os.Rename() (not working)
- 0.1.5 - 2017/08/28 : style without layers - misleading log message fixed
- 0.2.0 - 2018/12/01 : new projection option in mapnik driver
                       data structures modified (to allow more flexibility)
                       some changes are not compatible with initial release

Author:
- Klaus Tockloth

Copyright and license:
- Copyright (c) 2017,2018 Klaus Tockloth
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

Workflow (abstracted):
- waiting for build order
- build user mapnik xml
- build map (in temp dir)
- zip final map
- move final map to dest dir
- update map state
- delete temp dir

Contact (eMail):
- printmaps.service@gmail.com

Remarks:
- Cross compilation for Linux: env GOOS=linux GOARCH=amd64 go build -v

Logging:
- The log file is intended for reading by humans.
- It only contains service state and error informations.

ToDo:

Links:
- http://www.printmaps-osm.de
*/

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// general program info
var (
	progName    = os.Args[0]
	progVersion = "0.2.0"
	progDate    = "2018/12/01"
	progPurpose = "Printmaps Buildservice"
	progInfo    = "Build service to build large printable maps."
)

// Config defines all program settings
type Config struct {
	Logfile      string
	Workdir      string
	Maxprocs     int
	Graceperiod  int
	Metrics      bool
	Testmode     bool
	Mapnikdriver string
	Markersdir   string
	Styles       []struct {
		Name    string
		XMLPath string
		XMLFile string
	}
}

var config Config

const (
	mapBasename = "printmaps"
	pdfMetaname = "meta.pdf"
)

// BuildResult holds the result of the build process
type BuildResult struct {
	BuildSuccessful string
	BuildMessage    string
}

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

	configfile := "printmaps_buildservice.yaml"
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
	log.Printf("config maxprocs = %d", config.Maxprocs)
	log.Printf("config graceperiod = %d", config.Graceperiod)
	log.Printf("config metrics = %t", config.Metrics)
	log.Printf("config testmode = %t", config.Testmode)
	log.Printf("config mapnikdriver = %s", config.Mapnikdriver)
	log.Printf("config markersdir = %s", config.Markersdir)
	for _, style := range config.Styles {
		log.Printf("config map style: %s, %s, %s", style.Name, style.XMLPath, style.XMLFile)
	}

	// change into working directory
	if err = os.Chdir(config.Workdir); err != nil {
		log.Fatalf("fatal error <%v> at os.Chdir(), dir = <%s>", err, config.Workdir)
	}

	// save current working directory setting
	PathWorkdir, err = os.Getwd()
	if err != nil {
		log.Fatalf("fatal error <%v> at os.Getwd()", err)
	}
	log.Printf("config workdir (relative) = %s", config.Workdir)
	log.Printf("workdir (absolute) = %s", PathWorkdir)

	pid := os.Getpid()
	log.Printf("own process identifier (pid) = %d", pid)
	log.Printf("shutdown with SIGINT or SIGTERM")

	// create 'maps' and 'orders' directory (if necessary)
	createDirectories()

	// start timer trigger
	timerTrigger := time.Tick(time.Second * 5)

	// start shutdown trigger (subscribe to signals)
	shutdownTrigger := make(chan os.Signal)
	signal.Notify(shutdownTrigger, syscall.SIGINT)  // kill -SIGINT pid -> interrupt
	signal.Notify(shutdownTrigger, syscall.SIGTERM) // kill -SIGTERM pid -> terminated

	// start 'work done' trigger
	var workerCount = 0
	workDoneTrigger := make(chan struct{})

	// fetch work and start worker
ForeverLoop:
	for {
		if workerCount < config.Maxprocs {
			nextOrder := nextBuildOrder()
			if nextOrder != "" {
				workerCount++
				go buildMapMaster(nextOrder, workDoneTrigger)
			}
		}

		// wait for 'work done' event, timer or shutdown trigger
		select {
		case <-workDoneTrigger:
			workerCount--
		case <-timerTrigger:
			// unblock select
		case <-shutdownTrigger:
			// initiate shutdown
			break ForeverLoop
		}
	}

	gracePeriodTrigger := time.After(time.Second * time.Duration(config.Graceperiod))
	log.Printf("shutting down %s (%s), grace period = %d sec ...", progName, progPurpose, config.Graceperiod)
	for {
		if workerCount > 0 {
			log.Printf("shutdown in progress, waiting for %d worker(s) to finish ...", workerCount)
		} else {
			break
		}
		select {
		case <-workDoneTrigger:
			workerCount--
		case <-timerTrigger:
			// unblock select
		case <-gracePeriodTrigger:
			log.Printf("%s (%s) shutdown forced after end of grace period", progName, progPurpose)
			os.Exit(1)
		}
	}

	log.Printf("%s (%s) gracefully shut down", progName, progPurpose)
}

/*
buildMapMaster builds a map (master)
*/
func buildMapMaster(nextOrder string, chanOut chan<- struct{}) {

	// create temp directory
	tempdir, err := ioutil.TempDir(PathWorkdir, "printmaps_tempdir_")
	if err != nil {
		log.Fatalf("fatal error <%v> at ioutil.TempDir()", err)
	}

	// tempdir := filepath.Join(PathWorkdir, "printmaps_tempdir_"+nextOrder)
	// err := os.Mkdir(tempdir, 0700)
	// if err != nil {
	// 	log.Fatalf("fatal error <%v> at os.Mkdir(), dir = <%s>", err, tempdir)
	// }

	// move next order file into temp directory
	// the rename operation fails under some rare circumstances (heavy io load, Ubuntu 16.04 LTS)
	// in case of a failure (workaround, unfortunately not working):
	// - rename operation will be repeated after 10 seconds
	// - rename operation failure will only be logged
	source := filepath.Join(PathWorkdir, PathOrders, nextOrder)
	destination := filepath.Join(tempdir, nextOrder)
	if err = os.Rename(source, destination); err != nil {
		log.Printf("first attempt - critical error <%v> at os.Rename(), source = <%v>, destination = <%v>", err, source, destination)
		time.Sleep(9876 * time.Millisecond)
		if err2 := os.Rename(source, destination); err2 != nil {
			log.Printf("second attempt - critical error <%v> at os.Rename(), source = <%v>, destination = <%v>", err2, source, destination)
			chanOut <- struct{}{}
			return
		}
	}

	// build the printable map
	start := time.Now()
	buildMap(tempdir, nextOrder)
	elapsed := time.Since(start)

	// write metrics
	if config.Metrics {
		writeMetrics(tempdir, nextOrder, elapsed)
	}

	if config.Testmode == false {
		// remove temp directory
		if err = os.RemoveAll(tempdir); err != nil {
			log.Fatalf("fatal error <%v> at os.RemoveAll(); dir = <%s>", err, tempdir)
		}
	}

	// send 'work done' event
	chanOut <- struct{}{}
}

/*
buildMap builds a map
*/
func buildMap(tempdir string, order string) {

	var pmData PrintmapsData
	var pmState PrintmapsState
	var bResult BuildResult

	// read meta data of map order
	file := filepath.Join(tempdir, order)
	if err := readOrder(&pmData, file); err != nil {
		log.Printf("error <%v> at readOrder(), file = <%s>", err, file)
		return
	}

	// read state
	if err := readMapstate(&pmState, pmData.Data.ID); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("error <%v> at readMapstate()", err)
			log.Printf("pmData = %v", dumpPrintmapsData(pmData))
			return
		}
	}

	// write (update) state (map build started)
	pmState.Data.Attributes.MapBuildStarted = time.Now().Format(time.RFC3339)
	pmState.Data.Attributes.MapBuildCompleted = ""
	pmState.Data.Attributes.MapBuildSuccessful = ""
	pmState.Data.Attributes.MapBuildMessage = ""
	if err := writeMapstate(pmState); err != nil {
		log.Printf("error <%v> at writeMapstate()", err)
		// log.Printf("pmData = %v", dumpPrintmapsData(pmData))
		// log.Printf("pmState = %v", dumpPrintmapsState(pmState))
		return
	}

	// build mapnik map
	if err := buildMapnikMap(tempdir, pmData, &pmState); err != nil {
		bResult.BuildSuccessful = "no"
		bResult.BuildMessage = err.Error()
		setBuildResult(pmState, bResult)
		// log.Printf("error <%v> at buildMapnikMap()", err)
		// log.Printf("pmData = %v", dumpPrintmapsData(pmData))
		// log.Printf("pmState = %v", dumpPrintmapsState(pmState))
		return
	}

	if pmData.Data.Attributes.Fileformat == "pdf" {
		if err := modifyPDFMetadata(pmData); err != nil {
			bResult.BuildSuccessful = "no"
			bResult.BuildMessage = "error modifying pdf meta data"
			setBuildResult(pmState, bResult)
			log.Printf("error <%v> at modifyPDFMetadata()", err)
			// log.Printf("pmData = %v", dumpPrintmapsData(pmData))
			// log.Printf("pmState = %v", dumpPrintmapsState(pmState))
			return
		}
	}

	// zip map into standard download file (-j = junk directory names)
	zipfile := filepath.Join(tempdir, FileMapfile)
	mapfile := filepath.Join(tempdir, mapBasename+"."+pmData.Data.Attributes.Fileformat)
	command := fmt.Sprintf("zip -j %s %s", zipfile, mapfile)
	_, _, err := runCommand(command)
	if err != nil {
		bResult.BuildSuccessful = "no"
		bResult.BuildMessage = "error zipping map file"
		setBuildResult(pmState, bResult)
		log.Printf("error <%v> at runCommand()", err)
		// log.Printf("pmData = %v", dumpPrintmapsData(pmData))
		// log.Printf("pmState = %v", dumpPrintmapsState(pmState))
		return
	}

	// move map from temp directory to download location
	destination := filepath.Join(PathWorkdir, PathMaps, pmData.Data.ID, FileMapfile)
	if err := os.Rename(zipfile, destination); err != nil {
		bResult.BuildSuccessful = "no"
		bResult.BuildMessage = "error moving zipped map to download location"
		setBuildResult(pmState, bResult)
		log.Printf("error <%v> at os.Rename(), source = <%v>, destination = <%v>", err, zipfile, destination)
		// log.Printf("pmData = %v", dumpPrintmapsData(pmData))
		// log.Printf("pmState = %v", dumpPrintmapsState(pmState))
		return
	}

	// everything ok
	bResult.BuildSuccessful = "yes"
	bResult.BuildMessage = "map build successful"
	setBuildResult(pmState, bResult)
}

/*
setBuildResult sets the result state of the map build process
*/
func setBuildResult(pmState PrintmapsState, bResult BuildResult) error {

	// write (update) state (map build completed)
	pmState.Data.Attributes.MapBuildCompleted = time.Now().Format(time.RFC3339)
	pmState.Data.Attributes.MapBuildSuccessful = bResult.BuildSuccessful
	pmState.Data.Attributes.MapBuildMessage = bResult.BuildMessage
	if err := writeMapstate(pmState); err != nil {
		log.Printf("error <%v> at writeMapstate(), pmState = <%#v>", err, pmState)
		return err
	}

	return nil
}

/*
dumpPrintmapsData dumps (formats) a PrintmapsData object
*/
func dumpPrintmapsData(pmData PrintmapsData) string {

	dump, err := json.MarshalIndent(pmData, indentPrefix, indexString)
	if err != nil {
		message := fmt.Sprintf("error <%v> at json.MarshalIndent()", err)
		log.Printf("%s", message)
		return message
	}
	return string(dump)
}

/*
dumpPrintmapsState dumps (formats) a PrintmapsState object
*/
func dumpPrintmapsState(pmState PrintmapsState) string {

	dump, err := json.MarshalIndent(pmState, indentPrefix, indexString)
	if err != nil {
		message := fmt.Sprintf("error <%v> at json.MarshalIndent()", err)
		log.Printf("%s", message)
		return message
	}
	return string(dump)
}

/*
readOrder reads the map order (meta) data
*/
func readOrder(pmData *PrintmapsData, file string) error {

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("error <%v> at ioutil.ReadFile(), file = <%s>", err, file)
		return err
	}

	err = json.Unmarshal(data, pmData)
	if err != nil {
		log.Printf("error <%v> at json.Unmarshal()", err)
		return err
	}

	return nil
}

/*
writeMetrics writes a simple metrics string into the log
*/
func writeMetrics(tempdir string, order string, elapsed time.Duration) {

	var pmData PrintmapsData
	var pmState PrintmapsState

	// read meta data of map order
	file := filepath.Join(tempdir, order)
	if err := readOrder(&pmData, file); err != nil {
		log.Printf("error <%v> at readOrder(), file = <%s>", err, file)
		return
	}

	// read state
	if err := readMapstate(&pmState, pmData.Data.ID); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("error <%v> at readMapstate()", err)
			log.Printf("pmData = %v", dumpPrintmapsData(pmData))
			return
		}
	}

	// determine filesize (of final zip file)
	var filesize int64
	if pmState.Data.Attributes.MapBuildSuccessful == "yes" {
		file := filepath.Join(PathWorkdir, PathMaps, pmData.Data.ID, FileMapfile)
		fileinfo, err := os.Stat(file)
		if err == nil {
			filesize = fileinfo.Size()
		}
	}
	filesizeMB := float64(filesize) / (1024.0 * 1024.0)

	// format mapsize
	mapsize := fmt.Sprintf("%.1f x %.1f", pmData.Data.Attributes.PrintWidth, pmData.Data.Attributes.PrintHeight)

	// log metrics
	log.Printf("metrics: id=[%s], success=[%s], message=[%s], style=[%s], format=[%s], mapscale=[1:%d], mapsize=[%s mm], filesize=[%.2f MB], runtime=[%d sec]",
		pmData.Data.ID, pmState.Data.Attributes.MapBuildSuccessful, pmState.Data.Attributes.MapBuildMessage,
		pmData.Data.Attributes.Style, pmData.Data.Attributes.Fileformat, pmData.Data.Attributes.Scale, mapsize, filesizeMB, (elapsed / time.Second))
}
