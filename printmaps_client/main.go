/*
Purpose:
- Printmaps Command Line Client

Description:
- Creates large-sized maps in print quality.

Releases:
- v0.1.0 - 2017/05/23 : beta 1
- v0.1.1 - 2017/05/26 : general improvements
- v0.1.2 - 2017/05/28 : template improved
- v0.1.3 - 2017/06/01 : textual corrections
- v0.1.4 - 2017/06/27 : template modified
- v0.1.5 - 2017/07/04 : problem with upload filepath fixed
- v0.2.0 - 2018/10/24 : new helper 'coordgrid'
- v0.3.0 - 2018/12/04 : helper 'coordgrid' renamed to 'latlongrid'
					    helper 'rectangle' simplified
					    new helper 'utmgrid', utm2latlon', latlon2utm'
					    new helper 'bearingline', 'latlonline', 'utmline'
					    new helper 'passepartout', 'cropmarks'
					    map projection setting added
					    new helper 'runlua'
- v0.3.1 - 2018/12/10 : refactoring (data.go as package)
- v0.3.2 - 2019/01/21 : template modified
- v0.3.3 - 2019/01/22 : logic error fixed
- v0.3.4 - 2019/02/07 : client timeout setting removed
- v0.3.5 - 2019/02/14 : map definition template enhanced
- v0.3.6 - 2019/04/20 : hint at 'create()' added, path from 'progName' removed
- v0.3.7 - 2019/04/21 : help text improved
- v0.4.0 - 2019/05/18 : template modified
- v0.4.1 - 2019/05/19 : typo in template fixed
- v0.5.0 - 2019/05/21 : unzip action implemented
- v0.5.1 - 2019/06/27 : template modified
- v0.5.2 - 2020/05/22 : template modified
- v0.5.3 - 2020/07/04 : typo in help text corrected
- v0.5.4 - 2020/07/08 : minor correction
- v0.6.0 - 2020/08/03 : template removed
- v0.7.0 - 2021/06/12 : switch to modules, third-party libs updated, go 1.16.5
- v0.8.0 - 2025/01/04 : libs updated, go 1.23.4

Author:
- Klaus Tockloth

Copyright and license:
- Copyright (c) 2017-2025 Klaus Tockloth
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
- Lint: golangci-lint run --no-config --enable gocritic
- Vulnerability detection: govulncheck ./...

Links:
- http://www.printmaps-osm.de
*/

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/StefanSchroeder/Golang-Ellipsoid/ellipsoid"
	"github.com/davecgh/go-spew/spew"
	"github.com/im7mortal/UTM"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/printmaps/printmaps/pd"
	lua "github.com/yuin/gopher-lua"
	yaml "gopkg.in/yaml.v2"
)

// general program info
var (
	progName    = os.Args[0]
	progVersion = "v0.8.0"
	progDate    = "2025/01/04"
	progPurpose = "Printmaps Command Line Interface Client"
	progInfo    = "Creates large-sized maps in print quality."
)

// MapConfig represents the map configuration
type MapConfig struct {
	ServiceURL  string      `yaml:"ServiceURL"`
	Metadata    pd.Metadata `yaml:"Metadata,inline"`
	UploadFiles []string    `yaml:"UserFiles"`
}

var mapConfig MapConfig

// map identifier
var mapID string

// control files
var (
	mapDefinitionFile = "map.yaml"
	mapIDFile         = "map.id"
)

// http client
var netClient = &http.Client{}

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
	_, progName = filepath.Split(progName)

	if len(os.Args) == 1 {
		printUsage()
	}

	// read map definition
	if _, err := os.Stat(mapDefinitionFile); err == nil {
		source, err := os.ReadFile(mapDefinitionFile)
		if err != nil {
			log.Fatalf("error <%v> at os.ReadFile(), file = <%s>", err, mapDefinitionFile)
		}
		err = yaml.Unmarshal(source, &mapConfig)
		if err != nil {
			log.Fatalf("error <%v> at yaml.Unmarshal()", err)
		}
	}
	// fmt.Printf("\nmapConfig = %#v\n", mapConfig)
	// dumpData(os.Stdout, "mapConfig", mapConfig)

	fmt.Printf("\nurl = %v\n", mapConfig.ServiceURL)

	// read map id
	if _, err := os.Stat(mapIDFile); err == nil {
		filedata, err := os.ReadFile(mapIDFile)
		if err != nil {
			log.Fatalf("error <%v> at os.ReadFile(), file = <%s>", err, mapIDFile)
		}
		mapID = string(filedata)
	}
	fmt.Printf("map ID = %v\n", mapID)

	action := strings.ToLower(strings.Trim(os.Args[1], " ,"))
	fmt.Printf("action = %v\n", action)

	if action == "create" {
		checkMapDefinitionFile()
		create()
	} else if action == "update" {
		checkMapDefinitionFile()
		checkMapIDFile()
		update()
	} else if action == "upload" {
		checkMapDefinitionFile()
		checkMapIDFile()
		for _, file := range mapConfig.UploadFiles {
			upload(file)
		}
	} else if action == "state" {
		checkMapDefinitionFile()
		checkMapIDFile()
		fetch(action)
	} else if action == "order" {
		checkMapDefinitionFile()
		checkMapIDFile()
		order()
	} else if action == "download" {
		checkMapDefinitionFile()
		checkMapIDFile()
		download()
	} else if action == "data" {
		checkMapDefinitionFile()
		checkMapIDFile()
		fetch(action)
	} else if action == "delete" {
		checkMapDefinitionFile()
		checkMapIDFile()
		delete()
	} else if action == "capabilities" {
		fetch(action)
	} else if action == "unzip" {
		unzip()
	} else if action == "passepartout" {
		passepartout()
	} else if action == "rectangle" {
		rectangle()
	} else if action == "cropmarks" {
		cropmarks()
	} else if action == "latlongrid" {
		latlongrid()
	} else if action == "utmgrid" {
		utmgrid()
	} else if action == "latlon2utm" {
		latlon2utm()
	} else if action == "utm2latlon" {
		utm2latlon()
	} else if action == "bearingline" {
		bearingline()
	} else if action == "latlonline" {
		latlonline()
	} else if action == "utmline" {
		utmline()
	} else if action == "runlua" {
		runlua()
	} else {
		fmt.Printf("action <%v> not supported\n", action)
	}

	fmt.Printf("\n")
	os.Exit(0)
}

/*
checkMapDefinitionFile checks if the map definition file exists
*/
func checkMapDefinitionFile() {
	if _, err := os.Stat(mapDefinitionFile); os.IsNotExist(err) {
		fmt.Printf("\nERROR - PRECONDITION FAILED:\n")
		fmt.Printf("- the map definition file <%s> doesn't exists\n", mapDefinitionFile)
		os.Exit(1)
	}
}

/*
checkMapIDFile checks if the map id file exists
*/
func checkMapIDFile() {
	if _, err := os.Stat(mapIDFile); os.IsNotExist(err) {
		fmt.Printf("\nERROR - PRECONDITION FAILED:\n")
		fmt.Printf("- the map ID file <%s> doesn't exists\n", mapIDFile)
		fmt.Printf("- apply the 'create' action to create the file\n\n")
		os.Exit(1)
	}
}

/*
printUsage prints the usage of this program
*/
func printUsage() {
	fmt.Printf("\nProgram:\n")
	fmt.Printf("  Name         : %s\n", progName)
	fmt.Printf("  Release      : %s - %s\n", progVersion, progDate)
	fmt.Printf("  Purpose      : %s\n", progPurpose)
	fmt.Printf("  Info         : %s\n", progInfo)

	fmt.Printf("\nUsage:\n")
	fmt.Printf("  %s <action>\n", progName)

	fmt.Printf("\nExample:\n")
	fmt.Printf("  %s create\n", progName)

	fmt.Printf("\nActions:\n")
	fmt.Printf("  Primary      : create, update, upload, order, state, download\n")
	fmt.Printf("  Secondary    : data, delete, capabilities\n")
	fmt.Printf("  Helper       : unzip\n")
	fmt.Printf("  Helper       : passepartout, rectangle, cropmarks\n")
	fmt.Printf("  Helper       : latlongrid, utmgrid\n")
	fmt.Printf("  Helper       : latlon2utm, utm2latlon\n")
	fmt.Printf("  Helper       : bearingline, latlonline, utmline\n")
	fmt.Printf("  Helper       : runlua\n")

	fmt.Printf("\nRemarks:\n")
	fmt.Printf("  create       : creates the meta data for a new map\n")
	fmt.Printf("  update       : updates the meta data of an existing map\n")
	fmt.Printf("  upload       : uploads a list of user supplied files\n")
	fmt.Printf("  order        : places a map build order\n")
	fmt.Printf("  state        : fetches the current state of the map\n")
	fmt.Printf("  download     : downloads a successful build map\n")
	fmt.Printf("  data         : fetches the current meta data of the map\n")
	fmt.Printf("  delete       : deletes all artifacts (files) of the map\n")
	fmt.Printf("  capabilities : fetches the capabilities of the map service\n")
	fmt.Printf("  unzip        : unzips the downloaded map file\n")
	fmt.Printf("  passepartout : calculates wkt passe-partout from base values\n")
	fmt.Printf("  rectangle    : calculates wkt rectangle from base values\n")
	fmt.Printf("  cropmarks    : calculates wkt crop marks from base values\n")
	fmt.Printf("  latlongrid   : creates lat/lon grid in geojson format\n")
	fmt.Printf("  utmgrid      : creates utm grid in geojson format\n")
	fmt.Printf("  latlon2utm   : converts coordinates from lat/lon to utm\n")
	fmt.Printf("  utm2latlon   : converts coordinates from utm to lat/lon\n")
	fmt.Printf("  bearingline  : creates geographic line in geojson format\n")
	fmt.Printf("  latlonline   : creates geographic line in geojson format\n")
	fmt.Printf("  utmline      : creates geographic line in geojson format\n")
	fmt.Printf("  runlua       : run lua 5.1 script for generating labels\n")

	fmt.Printf("\nFiles:\n")
	fmt.Printf("  %-13s: unique map identifier\n", mapIDFile)
	fmt.Printf("  %-13s: map definition parameters\n", mapDefinitionFile)

	fmt.Printf("\nHow to start:\n")
	fmt.Printf("  - Download and unzip the 'template' map from 'printmaps-osm.de'.\n")
	fmt.Printf("  - Build the 'template' map by running the actions:\n")
	fmt.Printf("    'create', 'upload', 'order', 'state', 'download', 'unzip'\n")
	fmt.Printf("  - View the map with an appropriate application.\n")
	fmt.Printf("  - Modify the map definition file '%s' to your needs.\n", mapDefinitionFile)
	fmt.Printf("\n")

	os.Exit(1)
}

/*
create creates a new map
*/
func create() {
	if mapID != "" {
		fmt.Printf("\nnothing to do ... '%s' already exists\n", mapIDFile)
		fmt.Printf("use '%s update' to update '%s'\n", progName, mapDefinitionFile)
		return
	}

	pmData := pd.PrintmapsData{}
	pmData.Data.Type = "maps"
	pmData.Data.ID = mapID
	pmData.Data.Attributes = mapConfig.Metadata

	requestURL := mapConfig.ServiceURL + "metadata"

	data, err := json.MarshalIndent(pmData, pd.IndentPrefix, pd.IndexString)
	if err != nil {
		log.Fatalf("error <%v> at json.MarshalIndent()", err)
	}

	req, err := http.NewRequest("POST", requestURL, bytes.NewReader(data))
	if err != nil {
		log.Fatalf("error <%v> at http.NewRequest()", err)
	}

	req.Header.Add("Content-Type", "application/vnd.api+json; charset=utf-8")
	req.Header.Add("Accept", "application/vnd.api+json; charset=utf-8")

	printRequest(req, true)

	resp, err := netClient.Do(req)
	if err != nil {
		log.Fatalf("error <%v> at http.Do()", err)
	}
	defer resp.Body.Close()

	printResponse(resp, true)
	printSuccess(resp, http.StatusCreated)

	// write map id to file
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error <%v> at io.ReadAll()", err) // nolint
	}

	pmDataResponse := pd.PrintmapsData{}
	err = json.Unmarshal(data, &pmDataResponse)
	if err != nil {
		log.Fatalf("error <%v> at json.Unmarshal()", err)
	}

	err = os.WriteFile(mapIDFile, []byte(pmDataResponse.Data.ID), 0666)
	if err != nil {
		log.Fatalf("error <%v> at os.WriteFile()", err)
	}
}

/*
update updates an existing map
*/
func update() {
	pmData := pd.PrintmapsData{}
	pmData.Data.Type = "maps"
	pmData.Data.ID = mapID
	pmData.Data.Attributes = mapConfig.Metadata

	requestURL := mapConfig.ServiceURL + "metadata"

	data, err := json.MarshalIndent(pmData, pd.IndentPrefix, pd.IndexString)
	if err != nil {
		log.Fatalf("error <%v> at json.MarshalIndent()", err)
	}

	req, err := http.NewRequest("PATCH", requestURL, bytes.NewReader(data))
	if err != nil {
		log.Fatalf("error <%v> at http.NewRequest()", err)
	}

	req.Header.Add("Content-Type", "application/vnd.api+json; charset=utf-8")
	req.Header.Add("Accept", "application/vnd.api+json; charset=utf-8")

	printRequest(req, true)

	resp, err := netClient.Do(req)
	if err != nil {
		log.Fatalf("error <%v> at http.Do()", err)
	}
	defer resp.Body.Close()

	printResponse(resp, true)
	printSuccess(resp, http.StatusOK)
}

/*
upload uploads an user supplied file
*/
func upload(filename string) {
	requestURL := mapConfig.ServiceURL + "upload/" + mapID

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	_, destinationFilename := filepath.Split(filename)
	fileWriter, err := bodyWriter.CreateFormFile("file", destinationFilename)
	if err != nil {
		log.Fatalf("error <%v> at CreateFormFile()", err)
	}

	fh, err := os.Open(filename)
	if err != nil {
		log.Fatalf("error <%v> at os.Open(), file = %s", err, filename)
	}

	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		log.Fatalf("error <%v> at io.Copy()", err)
	}

	bodyWriter.Close()

	req, err := http.NewRequest("POST", requestURL, bodyBuf)
	if err != nil {
		log.Fatalf("error <%v> at http.NewRequest()", err)
	}

	req.Header.Add("Accept", "application/vnd.api+json; charset=utf-8")
	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	printRequest(req, false)

	resp, err := netClient.Do(req)
	if err != nil {
		log.Fatalf("error <%v> at http.Do()", err)
	}
	defer resp.Body.Close()

	printResponse(resp, true)
	printSuccess(resp, http.StatusCreated)
}

/*
order places the map order
*/
func order() {
	requestURL := mapConfig.ServiceURL + "mapfile"
	requestString := fmt.Sprintf("{\n    \"Data\": {\n        \"Type\": \"maps\",\n        \"ID\": \"%s\"\n    }\n}", mapID)

	req, err := http.NewRequest("POST", requestURL, strings.NewReader(requestString))
	if err != nil {
		log.Fatalf("error <%v> at http.NewRequest()", err)
	}

	req.Header.Add("Content-Type", "application/vnd.api+json; charset=utf-8")
	req.Header.Add("Accept", "application/vnd.api+json; charset=utf-8")

	printRequest(req, true)

	resp, err := netClient.Do(req)
	if err != nil {
		log.Fatalf("error <%v> at http.Do()", err)
	}
	defer resp.Body.Close()

	printResponse(resp, true)
	printSuccess(resp, http.StatusAccepted)
}

/*
fetch fetches information concerning the map or the service
*/
func fetch(action string) {
	requestURL := ""
	if action == "state" {
		requestURL = mapConfig.ServiceURL + "mapstate/" + mapID
	} else if action == "data" {
		requestURL = mapConfig.ServiceURL + "metadata/" + mapID
	} else if action == "capabilities" {
		requestURL = mapConfig.ServiceURL + "capabilities/service"
	} else {
		log.Fatalf("unexpected action <%v> in function fetch()", action)
	}

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Fatalf("error <%v> at http.NewRequest()", err)
	}

	req.Header.Add("Accept", "application/vnd.api+json; charset=utf-8")

	printRequest(req, true)

	resp, err := netClient.Do(req)
	if err != nil {
		log.Fatalf("error <%v> at http.Do()", err)
	}
	defer resp.Body.Close()

	printResponse(resp, true)
	printSuccess(resp, http.StatusOK)

	if action == "state" {
		fmt.Printf("\nattend status of 'MapBuildSuccessful'\n")
	}
}

/*
download downloads the map
*/
func download() {
	filename := "printmaps.zip"
	requestURL := mapConfig.ServiceURL + "mapfile/" + mapID

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Fatalf("error <%v> at http.NewRequest()", err)
	}

	req.Header.Add("Accept", "application/vnd.api+json; charset=utf-8")

	printRequest(req, true)

	resp, err := netClient.Do(req)
	if err != nil {
		log.Fatalf("error <%v> at http.Do()", err)
	}
	defer resp.Body.Close()

	expectedString := strconv.Itoa(http.StatusOK) + " " + http.StatusText(http.StatusOK)
	if resp.Status == expectedString {
		printResponse(resp, false)
	} else {
		printResponse(resp, true)
		printSuccess(resp, http.StatusOK)
		return
	}

	filesize := float64(resp.ContentLength) / (1024.0 * 1024.0)
	fmt.Printf("downloading file '%s' (%.1f MB) ... ", filename, filesize)

	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("error <%v> at os.Create(), file = <%s>", err, filename) // nolint
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatalf("error <%v> at io.Copy()", err)
	}

	fmt.Printf("done\n")
}

/*
delete deletes a map (server data and local map ID file)
*/
func delete() {
	if mapID == "" {
		fmt.Printf("\nnothing to do, map ID empty\n")
		return
	}

	// delete server data
	fmt.Printf("\ndeleting server data ...\n")

	requestURL := mapConfig.ServiceURL + mapID

	req, err := http.NewRequest("DELETE", requestURL, nil)
	if err != nil {
		log.Fatalf("error <%v> at http.NewRequest()", err)
	}

	req.Header.Add("Accept", "application/vnd.api+json; charset=utf-8")

	printRequest(req, true)

	resp, err := netClient.Do(req)
	if err != nil {
		log.Fatalf("error <%v> at http.Do()", err)
	}
	defer resp.Body.Close()

	printResponse(resp, true)
	printSuccess(resp, http.StatusNoContent)

	// remove local map ID file
	fmt.Printf("\nremoving local map ID file '%s' ...\n", mapIDFile)
	err = os.Remove(mapIDFile)
	if err != nil {
		log.Fatalf("error <%v> at os.Remove()", mapIDFile) // nolint
	}
	fmt.Printf("done\n")
}

/*
printRequest prints the http response to stdout
*/
func printRequest(req *http.Request, body bool) {
	dump, err := httputil.DumpRequestOut(req, body)
	if err != nil {
		log.Fatalf("error <%v> at httputil.DumpRequestOut()", err)
	}

	fmt.Printf("\nhttp request\n")
	fmt.Printf("------------\n")
	fmt.Printf("\n%s", dump)

	if !body {
		fmt.Printf("<request body omitted>\n")
	}
	fmt.Printf("\n")
}

/*
printResponse prints the http response to stdout
*/
func printResponse(resp *http.Response, body bool) {
	dump, err := httputil.DumpResponse(resp, body)
	if err != nil {
		log.Fatalf("error <%v> at httputil.DumpResponse()", err)
	}

	fmt.Printf("\nhttp response\n")
	fmt.Printf("-------------\n")
	fmt.Printf("\n%s", dump)

	if !body {
		fmt.Printf("<response body omitted>\n")
	}
	fmt.Printf("\n")
}

/*
printSuccess prints the success / failsure of the request
*/
func printSuccess(resp *http.Response, expectedStatus int) {
	fmt.Printf("\naction result\n")
	fmt.Printf("-------------\n")

	expectedString := strconv.Itoa(expectedStatus) + " " + http.StatusText(expectedStatus)
	fmt.Printf("\nexpected status = '%s'\n", expectedString)
	fmt.Printf("received status = '%s'\n", resp.Status)
	if resp.StatusCode == expectedStatus {
		fmt.Printf("success\n")
	} else {
		fmt.Printf("FAILURE\n")
	}
}

/*
unzip unzips downloaded map file
*/
func unzip() {
	archive := "printmaps.zip"
	fmt.Printf("\nExtracting archive %s ...\n", archive)
	zipReader, err := zip.OpenReader(archive)
	if err != nil {
		log.Fatalf("error <%v> at zip.OpenReader(), archive = <%s>", err, archive)
	}

	for _, file := range zipReader.Reader.File {
		zippedFile, err := file.Open()
		if err != nil {
			log.Fatalf("error <%v> at file.Open(), file = <%s>", err, file.Name)
		}
		defer zippedFile.Close()

		extractedFilePath := filepath.Join("./", file.Name)
		if file.FileInfo().IsDir() {
			fmt.Println("  Directory created:", extractedFilePath)
			err = os.MkdirAll(extractedFilePath, file.Mode())
			if err != nil {
				log.Fatalf("error <%v> at os.MkdirAll(), path = <%s>", err, extractedFilePath) // nolint
			}
		} else {
			fmt.Println("  File extracted:", file.Name)
			outputFile, err := os.OpenFile(extractedFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				log.Fatalf("error <%v> at os.OpenFile(), file = <%s>", err, file.Name)
			}
			defer outputFile.Close()

			_, err = io.Copy(outputFile, zippedFile)
			if err != nil {
				log.Fatalf("error <%v> at io.Copy()", err)
			}
		}
	}
}

/*
passepartout calculates well-known-text passe-partout from base values
*/
func passepartout() {
	if len(os.Args) != 8 {
		fmt.Printf("\nUsage:\n")
		fmt.Printf("  %s passepartout  width  height  left  top  right  bottom\n", progName)
		fmt.Printf("\nExample:\n")
		fmt.Printf("  %s passepartout  420.0  594.0  20.0  20.0  20.0  20.0\n", progName)
		fmt.Printf("\nRemarks:\n")
		fmt.Printf("  width = map width in millimeters\n")
		fmt.Printf("  height = map height in millimeters\n")
		fmt.Printf("  left = size of left border in millimeters\n")
		fmt.Printf("  top = size of top border in millimeters\n")
		fmt.Printf("  right = size of right border in millimeters\n")
		fmt.Printf("  bottom = size of bottom border in millimeters\n")
		fmt.Printf("\nHints:\n")
		fmt.Printf("  useful to create the map frame\n")
		fmt.Printf("\n")
		os.Exit(1)
	}

	width, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		log.Fatalf("\nerror: width (%s) not a float value\n", os.Args[2])
	}
	height, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		log.Fatalf("\nerror: height (%s) not a float value\n", os.Args[3])
	}
	left, err := strconv.ParseFloat(os.Args[4], 64)
	if err != nil {
		log.Fatalf("\nerror: left border size (%s) not a float value\n", os.Args[4])
	}
	top, err := strconv.ParseFloat(os.Args[5], 64)
	if err != nil {
		log.Fatalf("\nerror: top border size (%s) not a float value\n", os.Args[5])
	}
	right, err := strconv.ParseFloat(os.Args[6], 64)
	if err != nil {
		log.Fatalf("\nerror: right border size (%s) not a float value\n", os.Args[6])
	}
	bottom, err := strconv.ParseFloat(os.Args[7], 64)
	if err != nil {
		log.Fatalf("\nerror: bottom border size (%s) not a float value\n", os.Args[7])
	}

	// as polygon with hole
	fmt.Printf("\nwell-known-text (wkt) as rectangle with hole (frame) (probably what you want) ...\n")
	fmt.Printf("POLYGON((%.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f), (%.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f))\n",
		// outer
		0.0, 0.0,
		0.0, height,
		width, height,
		width, 0.0,
		0.0, 0.0,
		// inner
		left, bottom,
		left, height-top,
		width-right, height-top,
		width-right, bottom,
		left, bottom)

	// as two rectangles
	fmt.Printf("\nwell-known-text (wkt) as rectangle outlines (outer and inner) ...\n")
	fmt.Printf("LINESTRING(%.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f)\n",
		0.0, 0.0,
		0.0, height,
		width, height,
		width, 0.0,
		0.0, 0.0)
	fmt.Printf("LINESTRING(%.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f)\n",
		left, bottom,
		left, height-top,
		width-right, height-top,
		width-right, bottom,
		left, bottom)
}

/*
rectangle calculates well-known-text rectangle from base values
*/
func rectangle() {
	if len(os.Args) != 6 {
		fmt.Printf("\nUsage:\n")
		fmt.Printf("  %s rectangle  x  y  width  height\n", progName)
		fmt.Printf("\nExample:\n")
		fmt.Printf("  %s rectangle  40.0  40.0  60.0  80.0\n", progName)
		fmt.Printf("\nRemarks:\n")
		fmt.Printf("  x y = start position in millimeters\n")
		fmt.Printf("  width = rectangle width in millimeters\n")
		fmt.Printf("  height = rectangle height in millimeters\n")
		fmt.Printf("\nHints:\n")
		fmt.Printf("  useful to create info boxes\n")
		fmt.Printf("\n")
		os.Exit(1)
	}

	x, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		log.Fatalf("\nerror: x (%s) not a float value\n", os.Args[2])
	}
	y, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		log.Fatalf("\nerror: y (%s) not a float value\n", os.Args[3])
	}
	width, err := strconv.ParseFloat(os.Args[4], 64)
	if err != nil {
		log.Fatalf("\nerror: width (%s) not a float value\n", os.Args[4])
	}
	height, err := strconv.ParseFloat(os.Args[5], 64)
	if err != nil {
		log.Fatalf("\nerror: height (%s) not a float value\n", os.Args[5])
	}

	fmt.Printf("\nwell-known-text (wkt) for rectangular box ...\n")
	fmt.Printf("LINESTRING(%.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f)\n",
		x, y,
		x, y+height,
		x+width, y+height,
		x+width, y,
		x, y)
}

/*
cropmarks calculates well-known-text crop marks from base values
*/
func cropmarks() {
	if len(os.Args) != 5 {
		fmt.Printf("\nUsage:\n")
		fmt.Printf("  %s cropmarks  width  height  size\n", progName)
		fmt.Printf("\nExample:\n")
		fmt.Printf("  %s cropmarks  420.0  594.0  5.0\n", progName)
		fmt.Printf("\nRemarks:\n")
		fmt.Printf("  width = map width in millimeters\n")
		fmt.Printf("  height = map width in millimeters\n")
		fmt.Printf("  size = crop mark size in millimeters\n")
		fmt.Printf("\nHints:\n")
		fmt.Printf("  useful to create technical crop marks\n")
		fmt.Printf("\n")
		os.Exit(1)
	}

	width, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		log.Fatalf("\nerror: width (%s) not a float value\n", os.Args[2])
	}
	height, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		log.Fatalf("\nerror: height (%s) not a float value\n", os.Args[3])
	}
	size, err := strconv.ParseFloat(os.Args[4], 64)
	if err != nil {
		log.Fatalf("\nerror: size (%s) not a float value\n", os.Args[4])
	}

	fmt.Printf("\nwell-known-text (wkt) for crop marks ...\n")
	fmt.Printf("MULTILINESTRING((%.1f %.1f, %.1f %.1f, %.1f %.1f), "+
		"(%.1f %.1f, %.1f %.1f, %.1f %.1f), "+
		"(%.1f %.1f, %.1f %.1f, %.1f %.1f), "+
		"(%.1f %.1f, %.1f %.1f, %.1f %.1f))\n",
		// lower left crop mark
		0.0+size, 0.0,
		0.0, 0.0,
		0.0, 0.0+size,
		// upper left crop mark
		0.0+size, height,
		0.0, height,
		0.0, height-size,
		// upper right crop mark
		width-size, height,
		width, height,
		width, height-size,
		// lower right crop mark
		width-size, 0.0,
		width, 0.0,
		width, 0.0+size)
}

/*
latlongrid calculates/creates a lat/lon grid and saves it as GeoJSON files
*/
func latlongrid() {
	if len(os.Args) != 7 {
		fmt.Printf("\nUsage:\n")
		fmt.Printf("  %s latlongrid  latmin  lonmin  latmax  lonmax  distance\n", progName)
		fmt.Printf("\nExample:\n")
		fmt.Printf("  %s latlongrid  53.4  9.9  53.6  10.1  0.01\n", progName)
		fmt.Printf("\nRemarks:\n")
		fmt.Printf("  latmin lonmin = lower left / south west start point in decimal degrees\n")
		fmt.Printf("  latmax lonmax = upper right / north east end point in decimal degrees\n")
		fmt.Printf("  distance = grid distance in decimal degrees\n")
		fmt.Printf("  fractional digits of coord labels = fractional digits of grid distance\n")
		fmt.Printf("\nHints:\n")
		fmt.Printf("  useful to create a geographic lat/lon grid\n")
		fmt.Printf("\n")
		os.Exit(1)
	}

	latMin, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		log.Fatalf("\nerror: latmin (%s) not a float value\n", os.Args[2])
	}
	lonMin, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		log.Fatalf("\nerror: lonmin (%s) not a float value\n", os.Args[3])
	}
	latMax, err := strconv.ParseFloat(os.Args[4], 64)
	if err != nil {
		log.Fatalf("\nerror: latmax (%s) not a float value\n", os.Args[4])
	}
	lonMax, err := strconv.ParseFloat(os.Args[5], 64)
	if err != nil {
		log.Fatalf("\nerror: lonmax (%s) not a float value\n", os.Args[5])
	}
	gridDistance, err := strconv.ParseFloat(os.Args[6], 64)
	if err != nil {
		log.Fatalf("\nerror: distance (%s) not a float value\n", os.Args[6])
	}

	// derive coord label format from 'fractional digits of grid distance'
	fractionalDigits := 0
	tmp := strings.SplitN(os.Args[6], ".", 2)
	if len(tmp) > 1 {
		fractionalDigits = len(tmp[1])
	}
	coordLabelFormat := fmt.Sprintf("%%.%df", fractionalDigits)

	// verify input data
	if latMin >= latMax || lonMin >= lonMax {
		log.Fatalf("\nerror: invalid coord parameters\n")
	}

	distanceIntegerPart, _ := strconv.Atoi(tmp[0])
	distanceFractionalPart := 0
	if len(tmp) > 1 {
		distanceFractionalPart, _ = strconv.Atoi(tmp[1])
	}
	if distanceIntegerPart == 0 && distanceFractionalPart == 0 {
		log.Fatalf("\nerror: invalid distance\n")
	}

	// latitude grid lines (order: longitude, latitude)
	lsLat := make(orb.LineString, 0, 2)
	fcLat := geojson.NewFeatureCollection()

	for latTemp := latMin; latTemp <= latMax; latTemp += gridDistance {
		lsLat = append(lsLat, orb.Point{lonMin, latTemp})
		lsLat = append(lsLat, orb.Point{lonMax, latTemp})
		feature := geojson.NewFeature(lsLat)
		feature.Properties = make(map[string]interface{})
		feature.Properties["name"] = fmt.Sprintf(coordLabelFormat, latTemp)
		fcLat.Append(feature)
		lsLat = nil // reuse LineString object
	}

	dataJSON, err := json.MarshalIndent(fcLat, "", "  ")
	if err != nil {
		log.Fatalf("\nerror <%v> at json.MarshalIndent()\n", err)
	}

	// write data ([]byte) to file
	filename := "latgrid.geojson"
	if err := os.WriteFile(filename, dataJSON, 0666); err != nil {
		log.Fatalf("\nerror <%v> at os.WriteFile(); file = <%v>\n", err, filename)
	}

	fmt.Printf("\nlatitude coordinate grid lines: %s\n", filename)

	// longitude grid lines (order: longitude, latitude)
	lsLon := make(orb.LineString, 0, 2)
	fcLon := geojson.NewFeatureCollection()

	for lonTemp := lonMin; lonTemp <= lonMax; lonTemp += gridDistance {
		lsLon = append(lsLon, orb.Point{lonTemp, latMin})
		lsLon = append(lsLon, orb.Point{lonTemp, latMax})
		feature := geojson.NewFeature(lsLon)
		feature.Properties = make(map[string]interface{})
		feature.Properties["name"] = fmt.Sprintf(coordLabelFormat, lonTemp)
		fcLon.Append(feature)
		lsLon = nil // reuse LineString object
	}

	dataJSON, err = json.MarshalIndent(fcLon, "", "  ")
	if err != nil {
		log.Fatalf("\nerror <%v> at json.MarshalIndent()\n", err)
	}

	// write data ([]byte) to file
	filename = "longrid.geojson"
	if err := os.WriteFile(filename, dataJSON, 0666); err != nil {
		log.Fatalf("\nerror <%v> at os.WriteFile(); file = <%v>\n", err, filename)
	}

	fmt.Printf("longitude coordinate grid lines: %s\n", filename)
}

/*
utmgrid calculates/creates a UTM grid and saves it as GeoJSON files
*/
func utmgrid() {
	if len(os.Args) != 5 {
		fmt.Printf("\nUsage:\n")
		fmt.Printf("  %s utmgrid  utmmin  utmmax  distance\n", progName)
		fmt.Printf("\nExample:\n")
		fmt.Printf("  %s utmgrid  \"32 North 390000 5730000\"  \"32 North 430000 5760000\"  10000\n", progName)
		fmt.Printf("\nRemarks:\n")
		fmt.Printf("  utmmin = lower left start point in UTM format\n")
		fmt.Printf("  utmmax = upper right end point in UTM format\n")
		fmt.Printf("  distance = grid distance / square sidelength in meter\n")
		fmt.Printf("  utm format = zone number, hemisphere, easting, northing\n")
		fmt.Printf("  the hemisphere is either North or South\n")
		fmt.Printf("  a zone letter is not necessary\n")
		fmt.Printf("\nHints:\n")
		fmt.Printf("  useful to create a geodetic utm grid\n")
		fmt.Printf("\n")
		os.Exit(1)
	}

	utmMinZoneNumber := 0
	utmMinHemisphere := ""
	utmMinEasting := 0.0
	utmMinNorthing := 0.0
	n, err := fmt.Sscanf(os.Args[2], "%d%s%f%f", &utmMinZoneNumber, &utmMinHemisphere, &utmMinEasting, &utmMinNorthing)
	if err != nil {
		log.Fatalf("\nerror <%v> at fmt.Sscanf(); value = <%v>\n", err, os.Args[2])
	}
	if n != 4 {
		log.Fatalf("\nnumber of items unsufficient; expected = <%d>, parsed = <%d>; value = <%v>\n", 4, n, os.Args[2])
	}

	utmMaxZoneNumber := 0
	utmMaxHemisphere := ""
	utmMaxEasting := 0.0
	utmMaxNorthing := 0.0
	n, err = fmt.Sscanf(os.Args[3], "%d%s%f%f", &utmMaxZoneNumber, &utmMaxHemisphere, &utmMaxEasting, &utmMaxNorthing)
	if err != nil {
		log.Fatalf("\nerror <%v> at fmt.Sscanf(); value = <%v>\n", err, os.Args[3])
	}
	if n != 4 {
		log.Fatalf("\nnumber of items unsufficient; expected = <%d>, parsed = <%d>; value = <%v>\n", 4, n, os.Args[3])
	}

	// verify input data
	if utmMinZoneNumber != utmMaxZoneNumber {
		log.Fatalf("\nerror: utm zone numbers (%d / %d) not identical\n", utmMinZoneNumber, utmMaxZoneNumber)
	}
	if utmMinHemisphere != utmMaxHemisphere {
		log.Fatalf("\nerror: utm hemispheres (%s / %s) not identical\n", utmMinHemisphere, utmMaxHemisphere)
	}
	var northern bool
	if strings.ToLower(utmMinHemisphere) == "north" {
		northern = true
	} else if strings.ToLower(utmMinHemisphere) == "south" {
		northern = false
	} else {
		log.Fatalf("\nerror: utm hemisphere must be either North or South\n")
	}

	gridDistance, err := strconv.ParseFloat(os.Args[4], 64)
	if err != nil {
		log.Fatalf("\nerror <%v> at strconv.Atoi(); value = <%v>\n", err, os.Args[4])
	}

	// latitude / horizontal grid lines (order: longitude, latitude)
	lsLat := make(orb.LineString, 0, 2)
	fcLat := geojson.NewFeatureCollection()

	for utmNorthingTemp := utmMinNorthing; utmNorthingTemp <= utmMaxNorthing; utmNorthingTemp += gridDistance {
		latStart, lonStart, err := UTM.ToLatLon(utmMinEasting, utmNorthingTemp, utmMinZoneNumber, "", northern)
		if err != nil {
			log.Fatalf("\nerror <%v> at UTM.ToLatLon()\n", err)
		}
		latEnd, lonEnd, err := UTM.ToLatLon(utmMaxEasting, utmNorthingTemp, utmMinZoneNumber, "", northern)
		if err != nil {
			log.Fatalf("\nerror <%v> at UTM.ToLatLon()\n", err)
		}
		lsLat = append(lsLat, orb.Point{lonStart, latStart})
		lsLat = append(lsLat, orb.Point{lonEnd, latEnd})
		feature := geojson.NewFeature(lsLat)
		feature.Properties = make(map[string]interface{})
		feature.Properties["name"] = fmt.Sprintf("easting  %.0f", utmNorthingTemp)
		fcLat.Append(feature)
		lsLat = nil // reuse LineString object
	}

	dataJSON, err := json.MarshalIndent(fcLat, "", "  ")
	if err != nil {
		log.Fatalf("error <%v> at json.MarshalIndent()", err)
	}

	// write data ([]byte) to file
	filename := fmt.Sprintf("utmlatgrid%.0f.geojson", gridDistance)
	if err := os.WriteFile(filename, dataJSON, 0666); err != nil {
		log.Fatalf("error <%v> at os.WriteFile(); file = <%v>", err, filename)
	}

	fmt.Printf("\nutm latitude (horizontal) coordinate grid lines: %s\n", filename)

	// longitude / vertical grid lines (order: longitude, latitude)
	lsLon := make(orb.LineString, 0, 2)
	fcLon := geojson.NewFeatureCollection()

	for utmEastingTemp := utmMinEasting; utmEastingTemp <= utmMaxEasting; utmEastingTemp += gridDistance {
		latStart, lonStart, err := UTM.ToLatLon(utmEastingTemp, utmMinNorthing, utmMinZoneNumber, "", northern)
		if err != nil {
			log.Fatalf("\nerror <%v> at UTM.ToLatLon()\n", err)
		}
		latEnd, lonEnd, err := UTM.ToLatLon(utmEastingTemp, utmMaxNorthing, utmMinZoneNumber, "", northern)
		if err != nil {
			log.Fatalf("\nerror <%v> at UTM.ToLatLon()\n", err)
		}
		lsLon = append(lsLon, orb.Point{lonStart, latStart})
		lsLon = append(lsLon, orb.Point{lonEnd, latEnd})
		feature := geojson.NewFeature(lsLon)
		feature.Properties = make(map[string]interface{})
		feature.Properties["name"] = fmt.Sprintf("%.0f  northing", utmEastingTemp)
		fcLon.Append(feature)
		lsLon = nil // reuse LineString object
	}

	dataJSON, err = json.MarshalIndent(fcLon, "", "  ")
	if err != nil {
		log.Fatalf("error <%v> at json.MarshalIndent()", err)
	}

	// write data ([]byte) to file
	filename = fmt.Sprintf("utmlongrid%.0f.geojson", gridDistance)
	if err := os.WriteFile(filename, dataJSON, 0666); err != nil {
		log.Fatalf("error <%v> at os.WriteFile(); file = <%v>", err, filename)
	}

	fmt.Printf("utm longitude (vertical) coordinate grid lines: %s\n", filename)
}

/*
latlon2utm converts lat/lon to utm
*/
func latlon2utm() {
	if len(os.Args) != 4 {
		fmt.Printf("\nUsage:\n")
		fmt.Printf("  %s latlon2utm  lat  lon\n", progName)
		fmt.Printf("\nExample:\n")
		fmt.Printf("  %s latlon2utm  53.4012  9.9940\n", progName)
		fmt.Printf("\nHints:\n")
		fmt.Printf("  useful to convert lat/lon -> utm\n")
		fmt.Printf("\n")
		os.Exit(1)
	}

	lat, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		log.Fatalf("\nerror <%v> at strconv.ParseFloat(); value = <%v>\n", err, os.Args[2])
	}

	lon, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		log.Fatalf("\nerror <%v> at strconv.ParseFloat(); value = <%v>\n", err, os.Args[3])
	}

	easting, northing, zoneNumber, zoneLetter, err := UTM.FromLatLon(lat, lon, false)
	if err != nil {
		log.Fatalf("\nerror <%v> at UTM.FromLatLon()\n", err)
	}

	fmt.Printf("\nLat Lon = %s\n", formatLatLon(lat, lon))
	fmt.Printf("UTM     = %s\n", formatUTM(easting, northing, zoneNumber, zoneLetter))
}

/*
utm2latlon converts utm to lat/lon
*/
func utm2latlon() {
	if len(os.Args) != 6 {
		fmt.Printf("\nUsage:\n")
		fmt.Printf("  %s utm2latlon  zonenumber  hemisphere  easting  northing\n", progName)
		fmt.Printf("\nExample:\n")
		fmt.Printf("  %s utm2latlon  32  North  399384  5757242\n", progName)
		fmt.Printf("\nHints:\n")
		fmt.Printf("  useful to convert utm -> lat/lon\n")
		fmt.Printf("\n")
		os.Exit(1)
	}

	zoneNumber, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("\nerror <%v> at strconv.Atoi(); value = <%v>\n", err, os.Args[2])
	}

	hemisphere := os.Args[3]
	var northern bool
	if strings.ToLower(hemisphere) == "north" {
		northern = true
	} else if strings.ToLower(hemisphere) == "south" {
		northern = false
	} else {
		log.Fatalf("\nerror: hemisphere must be North or South\n")
	}

	easting, err := strconv.ParseFloat(os.Args[4], 64)
	if err != nil {
		log.Fatalf("\nerror <%v> at strconv.ParseFloat(); value = <%v>\n", err, os.Args[4])
	}

	northing, err := strconv.ParseFloat(os.Args[5], 64)
	if err != nil {
		log.Fatalf("\nerror <%v> at strconv.ParseFloat(); value = <%v>\n", err, os.Args[5])
	}

	lat, lon, err := UTM.ToLatLon(easting, northing, zoneNumber, "", northern)
	if err != nil {
		log.Fatalf("\nerror <%v> at UTM.ToLatLon()\n", err)
	}

	fmt.Printf("\nUTM     = %s\n", formatUTM(easting, northing, zoneNumber, hemisphere))
	fmt.Printf("Lat Lon = %s\n", formatLatLon(lat, lon))
}

/*
formatUTM formats UTM data as string
*/
func formatUTM(easting float64, northing float64, zoneNumber int, zoneLetterOrHemisphere string) string {
	result := fmt.Sprintf("%d %s %.0f %.0f", zoneNumber, zoneLetterOrHemisphere, easting, northing)
	return result
}

/*
formatLatLon formats lat/lon data as string
*/
func formatLatLon(lat float64, lon float64) string {
	result := fmt.Sprintf("%.5f %.5f", lat, lon)
	return result
}

/*
latlonline creates a geographic line and saves it as GeoJSON files
*/
func latlonline() {
	if len(os.Args) != 8 {
		fmt.Printf("\nUsage:\n")
		fmt.Printf("  %s latlonline  latstart  lonstart  latend  lonend  linelabel  filename\n", progName)
		fmt.Printf("\nExample:\n")
		fmt.Printf("  %s latlonline  51.98130  7.51479  51.99928  7.51479 \"beeline\" beeline\n", progName)
		fmt.Printf("\nRemarks:\n")
		fmt.Printf("  latstart lonstart = line start point in decimal degrees\n")
		fmt.Printf("  latend lonend = line end point in decimal degrees\n")
		fmt.Printf("  linelabel = name of line\n")
		fmt.Printf("  filename = name of file (extension .geojson will be added)\n")
		fmt.Printf("\nHints:\n")
		fmt.Printf("  useful to create a beeline between two given points\n")
		fmt.Printf("\n")
		os.Exit(1)
	}

	latStart, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		log.Fatalf("\nerror: latstart (%s) not a float value\n", os.Args[2])
	}
	lonStart, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		log.Fatalf("\nerror: lonstart (%s) not a float value\n", os.Args[3])
	}
	latEnd, err := strconv.ParseFloat(os.Args[4], 64)
	if err != nil {
		log.Fatalf("\nerror: latend (%s) not a float value\n", os.Args[4])
	}
	lonEnd, err := strconv.ParseFloat(os.Args[5], 64)
	if err != nil {
		log.Fatalf("\nerror: lonend (%s) not a float value\n", os.Args[5])
	}
	label := os.Args[6]

	// construct geographic line in geojson format
	lsLine := make(orb.LineString, 0, 2)
	fcLine := geojson.NewFeatureCollection()
	lsLine = append(lsLine, orb.Point{lonStart, latStart})
	lsLine = append(lsLine, orb.Point{lonEnd, latEnd})
	feature := geojson.NewFeature(lsLine)
	feature.Properties = make(map[string]interface{})
	feature.Properties["name"] = label
	fcLine.Append(feature)

	dataJSON, err := json.MarshalIndent(fcLine, "", "  ")
	if err != nil {
		log.Fatalf("\nerror <%v> at json.MarshalIndent()\n", err)
	}

	// write data ([]byte) to file
	filename := os.Args[7] + ".geojson"
	if err := os.WriteFile(filename, dataJSON, 0666); err != nil {
		log.Fatalf("\nerror <%v> at os.WriteFile(); file = <%v>\n", err, filename)
	}

	fmt.Printf("\ngeographic line (filename): %s\n", filename)
	fmt.Printf("start point (lat lon): %f %f\n", latStart, lonStart)
	fmt.Printf("end point (lat lon): %f %f\n", latEnd, lonEnd)
}

/*
utmline creates a geographic line and saves it as GeoJSON files
*/
func utmline() {
	if len(os.Args) != 6 {
		fmt.Printf("\nUsage:\n")
		fmt.Printf("  %s utmline  utmstart  utmend  linelabel  filename\n", progName)
		fmt.Printf("\nExample:\n")
		fmt.Printf("  %s utmline  \"32 U 0400000 5319440\"  \"32 U 0401000 5319440\"  \"1000 Meter\"  scalebar-1000\n", progName)
		fmt.Printf("\nRemarks:\n")
		fmt.Printf("  utmstart = line start point as utm coordinate\n")
		fmt.Printf("  utmend = line end point as utm coordinate\n")
		fmt.Printf("  linelabel = name of line\n")
		fmt.Printf("  filename = name of file (extension .geojson will be added)\n")
		fmt.Printf("\nHints:\n")
		fmt.Printf("  useful to create a scalebar (utm)\n")
		fmt.Printf("\n")
		os.Exit(1)
	}

	utmStartZoneNumber := 0
	utmStartZoneLetter := ""
	utmStartEasting := 0.0
	utmStartNorthing := 0.0
	n, err := fmt.Sscanf(os.Args[2], "%d%s%f%f", &utmStartZoneNumber, &utmStartZoneLetter, &utmStartEasting, &utmStartNorthing)
	if err != nil {
		log.Fatalf("\nerror <%v> at fmt.Sscanf(); value = <%v>\n", err, os.Args[2])
	}
	if n != 4 {
		log.Fatalf("\nnumber of items unsufficient; expected = <%d>, parsed = <%d>; value = <%v>\n", 4, n, os.Args[2])
	}

	utmEndZoneNumber := 0
	utmEndZoneLetter := ""
	utmEndEasting := 0.0
	utmEndNorthing := 0.0
	n, err = fmt.Sscanf(os.Args[3], "%d%s%f%f", &utmEndZoneNumber, &utmEndZoneLetter, &utmEndEasting, &utmEndNorthing)
	if err != nil {
		log.Fatalf("\nerror <%v> at fmt.Sscanf(); value = <%v>\n", err, os.Args[3])
	}
	if n != 4 {
		log.Fatalf("\nnumber of items unsufficient; expected = <%d>, parsed = <%d>; value = <%v>\n", 4, n, os.Args[3])
	}

	label := os.Args[4]

	// verify input data
	if utmStartZoneNumber != utmEndZoneNumber {
		log.Fatalf("\nerror: utm zone numbers (%d / %d) not identical\n", utmStartZoneNumber, utmEndZoneNumber)
	}
	if utmStartZoneLetter != utmEndZoneLetter {
		log.Fatalf("\nerror: utm zone letters (%s / %s) not identical\n", utmStartZoneLetter, utmStartZoneLetter)
	}

	latStart, lonStart, err := UTM.ToLatLon(utmStartEasting, utmStartNorthing, utmStartZoneNumber, utmStartZoneLetter)
	if err != nil {
		log.Fatalf("\nerror <%v> at UTM.ToLatLon()\n", err)
	}
	latEnd, lonEnd, err := UTM.ToLatLon(utmEndEasting, utmEndNorthing, utmEndZoneNumber, utmEndZoneLetter)
	if err != nil {
		log.Fatalf("\nerror <%v> at UTM.ToLatLon()\n", err)
	}

	// construct geographic line in geojson format
	lsLine := make(orb.LineString, 0, 2)
	fcLine := geojson.NewFeatureCollection()
	lsLine = append(lsLine, orb.Point{lonStart, latStart})
	lsLine = append(lsLine, orb.Point{lonEnd, latEnd})
	feature := geojson.NewFeature(lsLine)
	feature.Properties = make(map[string]interface{})
	feature.Properties["name"] = label
	fcLine.Append(feature)

	dataJSON, err := json.MarshalIndent(fcLine, "", "  ")
	if err != nil {
		log.Fatalf("\nerror <%v> at json.MarshalIndent()\n", err)
	}

	// write data ([]byte) to file
	filename := os.Args[5] + ".geojson"
	if err := os.WriteFile(filename, dataJSON, 0666); err != nil {
		log.Fatalf("\nerror <%v> at os.WriteFile(); file = <%v>\n", err, filename)
	}

	fmt.Printf("\ngeographic line (filename): %s\n", filename)
	fmt.Printf("start point (lat lon): %f %f\n", latStart, lonStart)
	fmt.Printf("end point (lat lon): %f %f\n", latEnd, lonEnd)
}

/*
bearingline calculates/creates a geographic line and saves it as GeoJSON files
*/
func bearingline() {
	if len(os.Args) != 8 {
		fmt.Printf("\nUsage:\n")
		fmt.Printf("  %s bearingline  lat  lon  angle  length  linelabel  filename\n", progName)
		fmt.Printf("\nExample:\n")
		fmt.Printf("  %s bearingline  51.98130  7.51479  90.0  1000.0  \"1000 Meter\"  scalebar-1000\n", progName)
		fmt.Printf("\nRemarks:\n")
		fmt.Printf("  lat = line start point in decimal degrees\n")
		fmt.Printf("  lon = line start point in decimal degrees\n")
		fmt.Printf("  angle = angle in degrees\n")
		fmt.Printf("  length = line length in meters\n")
		fmt.Printf("  linelabel = name of line\n")
		fmt.Printf("  filename = name of file (extension .geojson will be added)\n")
		fmt.Printf("\nHints:\n")
		fmt.Printf("  useful to create a scalebar (webmercator)\n")
		fmt.Printf("  useful to create a declination line\n")
		fmt.Printf("\n")
		os.Exit(1)
	}

	latStart, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		log.Fatalf("\nerror: lat (%s) not a float value\n", os.Args[2])
	}
	lonStart, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		log.Fatalf("\nerror: lon (%s) not a float value\n", os.Args[3])
	}
	angle, err := strconv.ParseFloat(os.Args[4], 64)
	if err != nil {
		log.Fatalf("\nerror: angle (%s) not a float value\n", os.Args[4])
	}
	length, err := strconv.ParseFloat(os.Args[5], 64)
	if err != nil {
		log.Fatalf("\nerror: length (%s) not a float value\n", os.Args[5])
	}
	label := os.Args[6]

	// calculate line end point (on surface of an ellipsoid)
	// create Ellipsoid object with WGS84-ellipsoid, angle units are degrees, distance units are meter
	geo1 := ellipsoid.Init("WGS84", ellipsoid.Degrees, ellipsoid.Meter, ellipsoid.LongitudeIsSymmetric, ellipsoid.BearingIsSymmetric)
	latEnd, lonEnd := geo1.At(latStart, lonStart, length, angle)

	// construct geographic line in geojson format
	lsLine := make(orb.LineString, 0, 2)
	fcLine := geojson.NewFeatureCollection()
	lsLine = append(lsLine, orb.Point{lonStart, latStart})
	lsLine = append(lsLine, orb.Point{lonEnd, latEnd})
	feature := geojson.NewFeature(lsLine)
	feature.Properties = make(map[string]interface{})
	feature.Properties["name"] = label
	fcLine.Append(feature)

	dataJSON, err := json.MarshalIndent(fcLine, "", "  ")
	if err != nil {
		log.Fatalf("\nerror <%v> at json.MarshalIndent()\n", err)
	}

	// write data ([]byte) to file
	filename := os.Args[7] + ".geojson"
	if err := os.WriteFile(filename, dataJSON, 0666); err != nil {
		log.Fatalf("\nerror <%v> at os.WriteFile(); file = <%v>\n", err, filename)
	}

	fmt.Printf("\ngeographic line (filename): %s\n", filename)
	fmt.Printf("start point (lat lon): %f %f\n", latStart, lonStart)
	fmt.Printf("end point (lat lon): %f %f\n", latEnd, lonEnd)
}

/*
runlua runs an user supplied lua script
*/
func runlua() {
	if len(os.Args) != 3 {
		fmt.Printf("\nUsage:\n")
		fmt.Printf("  %s runlua  filename\n", progName)
		fmt.Printf("\nExample:\n")
		fmt.Printf("  %s runlua  utmlabels.lua\n", progName)
		fmt.Printf("\nRemarks:\n")
		fmt.Printf("  filename = user supplied lua script\n")
		fmt.Printf("\nHints:\n")
		fmt.Printf("  useful to generate grid labels\n")
		fmt.Printf("\n")
		os.Exit(1)
	}

	L := lua.NewState()
	defer L.Close()

	luaScript := os.Args[2]
	fmt.Printf("\nRunning '%s' with script '%s' ...\n", lua.LuaVersion, luaScript)
	if err := L.DoFile(luaScript); err != nil {
		log.Fatalf("\nerror <%v> at L.DoFile(); file = <%v>", err, luaScript) // nolint
	}
}

/*
dumpData dumps an arbitrary data object
*/
func dumpData(writer io.Writer, objectname string, object interface{}) { // nolint
	if _, err := fmt.Fprintf(writer, "---------- %s ----------\n%s\n", objectname, spew.Sdump(object)); err != nil {
		log.Fatalf("Fehler <%v> bei fmt.Fprintf", err)
	}
}
