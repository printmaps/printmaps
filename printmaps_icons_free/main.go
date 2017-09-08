/*
Purpose:
- modify svg map marker icons

Description:
- Printmaps utility program.

Releases:
- 1.0.0 - 2017/05/18 : initial release

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
- NN

Links:
- NN
*/

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// general program info
var (
	progName    = os.Args[0]
	progVersion = "0.1.0"
	progDate    = "2017/05/19"
	progPurpose = "modify svg map marker icons"
	progInfo    = "Printmaps utility program."
)

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

	printUsage()

	// über alle Dateien iterieren
	path := "./Centered"
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatalf("error <%v> at ioutil.ReadDir()", err)
	}

	for _, file := range files {
		inputFile := filepath.Join(path, file.Name())
		fmt.Printf("%s ... ", inputFile)

		// svg einlesen
		lines, err2 := slurpFile(inputFile)
		if err2 != nil {
			log.Fatalf("error <%v> at slurpFile()", err2)
		}

		// name modifizieren
		// '__' -> '_'
		// MapMarker -> 'Printmaps'
		outputFile := strings.Replace(file.Name(), "__", "_", -1)
		outputFile = strings.Replace(outputFile, "MapMarker", "Printmaps", -1)
		outputFile = filepath.Join("./markers/", outputFile)
		fmt.Printf("%s\n", outputFile)

		modifyCenteredMarkers(outputFile, lines)
	}

	// über alle Dateien iterieren
	path = "./NotCentered"
	files, err = ioutil.ReadDir(path)
	if err != nil {
		log.Fatalf("error <%v> at ioutil.ReadDir()", err)
	}

	for _, file := range files {
		inputFile := filepath.Join(path, file.Name())
		fmt.Printf("%s ... ", inputFile)

		// svg einlesen
		lines, err := slurpFile(inputFile)
		if err != nil {
			log.Fatalf("error <%v> at slurpFile()", err)
		}

		// name modifizieren
		// '__' -> '_'
		// MapMarker -> 'Printmaps'
		outputFile := strings.Replace(file.Name(), "__", "_", -1)
		outputFile = strings.Replace(outputFile, "MapMarker", "Printmaps", -1)
		outputFile = filepath.Join("./markers/", outputFile)
		fmt.Printf("%s\n", outputFile)

		modifyNotCenteredMarkers(outputFile, lines)
	}
}

/*
printUsage prints the usage of this program
*/
func printUsage() {

	fmt.Printf("\nProgram:\n")
	fmt.Printf("  Name    : %s\n", progName)
	fmt.Printf("  Release : %s - %s\n", progVersion, progDate)
	fmt.Printf("  Purpose : %s\n", progPurpose)
	fmt.Printf("  Info    : %s\n", progInfo)
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
modifyCenteredMarkers modifies and writes the svg file
*/
func modifyCenteredMarkers(filename string, lines []string) {

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error <%v> at os.OpenFile(), file = <%v>", err, filename)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalf("error <%v> at file.Close(), file = <%v>", err, filename)
		}
	}()

	writer := bufio.NewWriter(file)

	if strings.Index(filename, "_Point_") != -1 {
		// no modification
		for _, line := range lines {
			fmt.Fprintf(writer, "%s\n", line)
		}
	} else {
		for index, line := range lines {
			if index == 3 {
				fmt.Fprintf(writer, "\twidth=\"64px\" height=\"128px\" viewBox=\"0 0 64 128\" enable-background=\"new 0 0 64 128\" xml:space=\"preserve\">\n")
				fmt.Fprintf(writer, "<rect width=\"64px\" height=\"128px\" fill=\"green\" opacity=\"0.0\" />\n")
				continue
			}
			fmt.Fprintf(writer, "%s\n", line)
		}
	}

	if err := writer.Flush(); err != nil {
		log.Fatalf("error <%v> at writer.Flush(), file = <%v>", err, filename)
	}
}

/*
modifyNotCenteredMarkers modifies and writes the svg file
*/
func modifyNotCenteredMarkers(filename string, lines []string) {

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error <%v> at os.OpenFile(), file = <%v>", err, filename)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalf("error <%v> at file.Close(), file = <%v>", err, filename)
		}
	}()

	writer := bufio.NewWriter(file)

	if strings.Index(filename, "Flag1_Left_") >= 0 {
		for index, line := range lines {
			if index == 3 {
				fmt.Fprintf(writer, "\twidth=\"128px\" height=\"128px\" viewBox=\"0 0 128 128\" enable-background=\"new 0 0 128 128\" xml:space=\"preserve\">\n")
				fmt.Fprintf(writer, "<rect width=\"128px\" height=\"128px\" fill=\"green\" opacity=\"0.0\" />\n")
				fmt.Fprintf(writer, "<g transform=\"translate(16,0)\">\n")
				continue
			}
			if index == 4 {
				continue
			}
			fmt.Fprintf(writer, "%s\n", line)
		}
	} else if strings.Index(filename, "Flag1_Right_") >= 0 {
		for index, line := range lines {
			if index == 3 {
				fmt.Fprintf(writer, "\twidth=\"128px\" height=\"128px\" viewBox=\"0 0 128 128\" enable-background=\"new 0 0 128 128\" xml:space=\"preserve\">\n")
				fmt.Fprintf(writer, "<rect width=\"128px\" height=\"128px\" fill=\"green\" opacity=\"0.0\" />\n")
				fmt.Fprintf(writer, "<g transform=\"translate(48,0)\">\n")
				continue
			}
			if index == 4 {
				continue
			}
			fmt.Fprintf(writer, "%s\n", line)
		}
	} else if strings.Index(filename, "ChequeredFlag_Right_") >= 0 {
		for index, line := range lines {
			if index == 3 {
				fmt.Fprintf(writer, "\twidth=\"128px\" height=\"128px\" viewBox=\"0 0 128 128\" enable-background=\"new 0 0 128 128\" xml:space=\"preserve\">\n")
				fmt.Fprintf(writer, "<rect width=\"128px\" height=\"128px\" fill=\"green\" opacity=\"0.0\" />\n")
				fmt.Fprintf(writer, "<g transform=\"translate(64,0)\">\n")
				continue
			}
			if index == 4 {
				continue
			}
			fmt.Fprintf(writer, "%s\n", line)
		}
	} else if strings.Index(filename, "ChequeredFlag_Left_") >= 0 {
		for index, line := range lines {
			if index == 3 {
				fmt.Fprintf(writer, "\twidth=\"128px\" height=\"128px\" viewBox=\"0 0 128 128\" enable-background=\"new 0 0 128 128\" xml:space=\"preserve\">\n")
				fmt.Fprintf(writer, "<rect width=\"128px\" height=\"128px\" fill=\"green\" opacity=\"0.0\" />\n")
				continue
			}
			fmt.Fprintf(writer, "%s\n", line)
		}
	}

	if err := writer.Flush(); err != nil {
		log.Fatalf("error <%v> at writer.Flush(), file = <%v>", err, filename)
	}
}
