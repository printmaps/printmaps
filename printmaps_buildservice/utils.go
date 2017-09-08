// utils and helper

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

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
		if exitError, ok := err.(*exec.ExitError); ok {
			// command fails because of an unsuccessful exit code
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			log.Printf("command exit code = <%d>", waitStatus.ExitStatus())
		}
		log.Printf("error <%v> at cmd.CombinedOutput()", err)
		log.Printf("command (not successful) = <%s>", strings.Join(cmd.Args, " "))
		if len(commandOutput) > 0 {
			log.Printf("command output (stdout, stderr) =\n%s", string(commandOutput))
		}
	} else {
		// command was successful
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
		if config.Testmode {
			log.Printf("command (successful) = <%s>", strings.Join(cmd.Args, " "))
			log.Printf("command exit code = <%d>", waitStatus.ExitStatus())
			if len(commandOutput) > 0 {
				log.Printf("command output (stdout, stderr) =\n%s", string(commandOutput))
			}
		}
	}

	commandExitStatus = waitStatus.ExitStatus()
	return
}

/*
nextBuildOrder returns the name of the oldest build order file
*/
func nextBuildOrder() string {

	path := filepath.Join(PathWorkdir, PathOrders)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatalf("fatal error <%v> at ioutil.ReadDir(), path = <%v>", err, path)
	}

	if len(files) == 0 {
		return ""
	}

	sort.Slice(files, func(i, j int) bool { return files[i].ModTime().Unix() < files[j].ModTime().Unix() })

	for _, fileInfo := range files {
		if fileInfo.IsDir() == false {
			return fileInfo.Name()
		}
	}

	return ""
}

/*
modifyPDFMetadata modifies the metadata of the pdf file
*/
func modifyPDFMetadata(pmData PrintmapsData) error {

	sourcePDF := mapBasename + ".pdf"
	destinationPDF := pdfMetaname
	metadataConfig := "metadata.cfg"

	// create control file with pdf meta data (for usage with pdftk)
	file, err := os.OpenFile(metadataConfig, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("error <%v> at os.OpenFile(), file = <%v>", err, metadataConfig)
		return err
	}

	writer := bufio.NewWriter(file)

	fmt.Fprintf(writer, "InfoBegin\n")
	fmt.Fprintf(writer, "InfoKey:Title\n")
	fmt.Fprintf(writer, "InfoValue:Map image based on OpenStreetMap data\n")
	fmt.Fprintf(writer, "InfoBegin\n")
	fmt.Fprintf(writer, "InfoKey:Author\n")
	fmt.Fprintf(writer, "InfoValue:http://www.printmaps-osm.de\n")
	fmt.Fprintf(writer, "InfoBegin\n")
	fmt.Fprintf(writer, "InfoKey:Subject\n")
	fmt.Fprintf(writer, "InfoValue:(c) OpenStreetMap contributors\n")
	fmt.Fprintf(writer, "InfoBegin\n")
	fmt.Fprintf(writer, "InfoKey:Keywords\n")

	fmt.Fprintf(writer, "InfoValue:style='%s', lat='%f', lon='%f', scale='1:%d'\n",
		pmData.Data.Attributes.Style, pmData.Data.Attributes.Latitude, pmData.Data.Attributes.Longitude, pmData.Data.Attributes.Scale)

	fmt.Fprintf(writer, "InfoBegin\n")
	fmt.Fprintf(writer, "InfoKey:Creator\n")
	fmt.Fprintf(writer, "InfoValue:Mapnik Toolkit 3.x (http://mapnik.org)\n")
	fmt.Fprintf(writer, "InfoBegin\n")
	fmt.Fprintf(writer, "InfoKey:CreationDate\n")

	// Formatbeispiel: InfoValue:D:20150504160000+01'00'
	fmt.Fprintf(writer, "InfoValue:D:%s'00'\n", time.Now().Format("20060102150405-07"))

	if err = writer.Flush(); err != nil {
		log.Printf("error <%v> at writer.Flush(), file = <%v>", err, metadataConfig)
		return err
	}

	if err = file.Close(); err != nil {
		log.Printf("error <%v> at file.Close(), file = <%v>", err, metadataConfig)
		return err
	}

	// modify (write) pdf meta data

	command := fmt.Sprintf("pdftk %s update_info %s output %s", sourcePDF, metadataConfig, destinationPDF)
	_, _, err = runCommand(command)
	if err != nil {
		log.Printf("error <%v> at runCommand()", err)
		return err
	}

	return nil
}
