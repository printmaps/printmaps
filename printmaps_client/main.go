/*
Purpose:
- Printmaps Command Line Client

Description:
- Creates large-sized maps in print quality.

Releases:
- 0.1.0 - 2017/05/23 : beta 1
- 0.1.1 - 2017/05/26 : general improvements
- 0.1.2 - 2017/05/28 : template improved
- 0.1.3 - 2017/06/01 : textual corrections
- 0.1.4 - 2017/06/27 : template modified
- 0.1.5 - 2017/07/04 : problem with upload filepath fixed
- 0.2.0 - 2018/10/24 : new helper 'coordgrid'
- 0.3.0 - 2018/12/04 : helper 'coordgrid' renamed to 'latlongrid'
					   helper 'rectangle' simplified
					   new helper 'utmgrid', utm2latlon', latlon2utm'
					   new helper 'bearingline', 'latlonline', 'utmline'
					   new helper 'passepartout', 'cropmarks'
					   map projection setting added
					   new helper 'runlua'
- 0.3.1 - 2018/12/10 : refactoring (data.go as package)
- 0.3.2 - 2019/01/21 : template modified
- 0.3.3 - 2019/01/22 : logic error fixed
- 0.3.4 - 2019/02/07 : client timeout setting removed

Author:
- Klaus Tockloth

Copyright and license:
- Copyright (c) 2017-2019 Klaus Tockloth
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
- http://www.printmaps-osm.de
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	progVersion = "0.3.4"
	progDate    = "2019/02/07"
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

	if len(os.Args) == 1 {
		printUsage()
	}

	// read map definition
	if _, err := os.Stat(mapDefinitionFile); err == nil {
		source, err := ioutil.ReadFile(mapDefinitionFile)
		if err != nil {
			log.Fatalf("error <%v> at ioutil.ReadFile(), file = <%s>", err, mapDefinitionFile)
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
		filedata, err := ioutil.ReadFile(mapIDFile)
		if err != nil {
			log.Fatalf("error <%v> at ioutil.ReadFile(), file = <%s>", err, mapIDFile)
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
	} else if action == "template" {
		template()
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
		fmt.Printf("- apply the 'template' action to create the file\n\n")
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
	fmt.Printf("  Name    : %s\n", progName)
	fmt.Printf("  Release : %s - %s\n", progVersion, progDate)
	fmt.Printf("  Purpose : %s\n", progPurpose)
	fmt.Printf("  Info    : %s\n", progInfo)

	fmt.Printf("\nUsage:\n")
	fmt.Printf("  %s <action>\n", progName)

	fmt.Printf("\nExample:\n")
	fmt.Printf("  %s create\n", progName)

	fmt.Printf("\nActions:\n")
	fmt.Printf("  Primary   : create, update, upload, order, state, download\n")
	fmt.Printf("  Secondary : data, delete, capabilities\n")
	fmt.Printf("  Helper    : template\n")
	fmt.Printf("  Helper    : passepartout, rectangle, cropmarks\n")
	fmt.Printf("  Helper    : latlongrid, utmgrid\n")
	fmt.Printf("  Helper    : latlon2utm, utm2latlon\n")
	fmt.Printf("  Helper    : bearingline, latlonline, utmline\n")
	fmt.Printf("  Helper    : runlua\n")

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
	fmt.Printf("  template     : creates a template file for building a map\n")
	fmt.Printf("  passepartout : calculates wkt passe-partout from base values\n")
	fmt.Printf("  rectangle    : calculates wkt rectangle from base values\n")
	fmt.Printf("  cropmarks    : calculates wkt crop marks from base values\n")
	fmt.Printf("  latlongrid   : creates lat/lon grid in geojson format\n")
	fmt.Printf("  utmgrid      : careates utm grid in geojson format\n")
	fmt.Printf("  latlon2utm   : converts coordinates from lat/lon to utm\n")
	fmt.Printf("  utm2latlon   : converts coordinates from utm to lat/lon\n")
	fmt.Printf("  bearingline  : creates geographic line in geojson format\n")
	fmt.Printf("  latlonline   : creates geographic line in geojson format\n")
	fmt.Printf("  utmline      : creates geographic line in geojson format\n")
	fmt.Printf("  runlua       : run lua 5.1 script for generating labels\n")

	fmt.Printf("\nHow to start:\n")
	fmt.Printf("  - Start with creating a new directory on your local system.\n")
	fmt.Printf("  - Change into this directory and run the 'template' action.\n")
	fmt.Printf("  - This creates the default map definition file '%s'.\n", mapDefinitionFile)
	fmt.Printf("  - You have now a full working example for building a map.\n")
	fmt.Printf("  - Build the map in order to get familiar with this client.\n")
	fmt.Printf("  - Run the actions 'create', 'order', 'state' and 'download'.\n")
	fmt.Printf("  - Unzip the file and view it with an appropriate application.\n")
	fmt.Printf("  - Modify the map definition file '%s' to your needs.\n", mapDefinitionFile)
	fmt.Printf("\n")

	os.Exit(1)
}

/*
create creates a new map
*/
func create() {

	if mapID != "" {
		fmt.Printf("\nnothing to do ... map ID file '%s' exists\n", mapIDFile)
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
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error <%v> at ioutil.ReadAll()", err)
	}

	pmDataResponse := pd.PrintmapsData{}
	err = json.Unmarshal(data, &pmDataResponse)
	if err != nil {
		log.Fatalf("error <%v> at json.Unmarshal()", err)
	}

	err = ioutil.WriteFile(mapIDFile, []byte(pmDataResponse.Data.ID), 0666)
	if err != nil {
		log.Fatalf("error <%v> at ioutil.WriteFile()", err)
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
		log.Fatalf("error <%v> at os.Create(), file = <%s>", err, filename)
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
		log.Fatalf("error <%v> at os.Remove()", mapIDFile)
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

	if body == false {
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

	if body == false {
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
template creates a map definition template for building a map
*/
func template() {

	fmt.Printf("\ncreating map definition file '%s' ...\n", mapDefinitionFile)
	if _, err := os.Stat(mapDefinitionFile); err == nil {
		fmt.Printf("nothing done, map definition file '%s' already exists\n", mapDefinitionFile)
	} else {
		err := ioutil.WriteFile(mapDefinitionFile, []byte(mapTemplate), 0666)
		if err != nil {
			log.Fatalf("error <%v> at ioutil.WriteFile(), file = <%s>", err, mapDefinitionFile)
		}
		fmt.Printf("done\n")
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
		fmt.Printf("  height = map width in millimeters\n")
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
	if err := ioutil.WriteFile(filename, dataJSON, 0666); err != nil {
		log.Fatalf("\nerror <%v> at ioutil.WriteFile(); file = <%v>\n", err, filename)
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
	if err := ioutil.WriteFile(filename, dataJSON, 0666); err != nil {
		log.Fatalf("\nerror <%v> at ioutil.WriteFile(); file = <%v>\n", err, filename)
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
	if err := ioutil.WriteFile(filename, dataJSON, 0666); err != nil {
		log.Fatalf("error <%v> at ioutil.WriteFile(); file = <%v>", err, filename)
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
	if err := ioutil.WriteFile(filename, dataJSON, 0666); err != nil {
		log.Fatalf("error <%v> at ioutil.WriteFile(); file = <%v>", err, filename)
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
	if err := ioutil.WriteFile(filename, dataJSON, 0666); err != nil {
		log.Fatalf("\nerror <%v> at ioutil.WriteFile(); file = <%v>\n", err, filename)
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
	if utmStartZoneNumber != utmStartZoneNumber {
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
	if err := ioutil.WriteFile(filename, dataJSON, 0666); err != nil {
		log.Fatalf("\nerror <%v> at ioutil.WriteFile(); file = <%v>\n", err, filename)
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
	if err := ioutil.WriteFile(filename, dataJSON, 0666); err != nil {
		log.Fatalf("\nerror <%v> at ioutil.WriteFile(); file = <%v>\n", err, filename)
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
		log.Fatalf("\nerror <%v> at L.DoFile(); file = <%v>", err, luaScript)
	}
}

/*
dumpData dumps an arbitrary data object
*/
func dumpData(writer io.Writer, objectname string, object interface{}) {

	if _, err := fmt.Fprintf(writer, "---------- %s ----------\n%s\n", objectname, spew.Sdump(object)); err != nil {
		log.Fatalf("Fehler <%v> bei fmt.Fprintf", err)
	}
}

var mapTemplate = `# map definition file
# -------------------
# general hint for this yaml config file:
# - do not use tabs or unnecessary white spaces
#
# useful links:
# - https://github.com/mapnik/mapnik/wiki/SymbologySupport
# - http://mapnik.org/mapnik-reference
#
# basic symbolizers:
# - LinePatternSymbolizer (https://github.com/mapnik/mapnik/wiki/LinePatternSymbolizer)
# - LineSymbolizer (https://github.com/mapnik/mapnik/wiki/LineSymbolizer)
# - MarkersSymbolizer (https://github.com/mapnik/mapnik/wiki/MarkersSymbolizer)
# - PointSymbolizer (https://github.com/mapnik/mapnik/wiki/PointSymbolizer)
# - PolygonPatternSymbolizer (https://github.com/mapnik/mapnik/wiki/PolygonPatternSymbolizer)
# - PolygonSymbolizer (https://github.com/mapnik/mapnik/wiki/PolygonSymbolizer)
# - TextSymbolizer (https://github.com/mapnik/mapnik/wiki/TextSymbolizer)
#
# advanced symbolizers:
# - BuildingSymbolizer (https://github.com/mapnik/mapnik/wiki/BuildingSymbolizer)
# - RasterSymbolizer (https://github.com/mapnik/mapnik/wiki/RasterSymbolizer)
# - ShieldSymbolizer (https://github.com/mapnik/mapnik/wiki/ShieldSymbolizer)
#
# purpose: 
# author : 
# release: 
#
# frame (2 + 18 = bleed + frame):
# printmaps passepartout 420.0 594.0 20.0 20.0 20.0 20.0
#
# crop marks:
# printmaps cropmarks 420.0 594.0 5.0
#
# scalebar:
# printmaps bearingline 53.49777 9.93321 90.0 1000.0 "1000 Meter" scalebar-1000

# service configuration
# ---------------------

# URL of webservice
ServiceURL: http://printmaps-osm.de:8282/api/beta2/maps/

# proxy configuration (not to be done here)
# - set the environment variable $HTTP_PROXY on your local system 
# - e.g. export HTTP_PROXY=http://user:password@proxy.server:port

# essential map attributes (required)
# -----------------------------------

# file format (png, pdf, svg)
Fileformat: png

# scale as in "1:10000" (e.g. 10000, 25000)
Scale: 20000

# width and height (millimeter, e.g. 609.6)
PrintWidth: 420.0
PrintHeight: 594.0

# center coordinates (decimal degrees, e.g. 51.9506)
Latitude: 53.5459
Longitude: 9.9836

# style / design (osm-carto, osm-carto-mono, osm-carto-ele20, schwarzplan, schwarzplan+, raster10)
# raster10 (no map data): useful for placing / styling the user map elements
# request the service capabilities to get a list of all available map styles
Style: osm-carto

# map projection, EPSG code as number (without prefix "EPSG:")
# e.g. 3857 (EPSG:3857 / WGS84 / Web Mercator) (used by Google/Bing/OpenStreetMap)
# e.g. 32632 (EPSG:32632 / WGS 84 / UTM Zone 32N)
# e.g. 27700 (EPSG:27700 / OSGB 1936 / British National Grid)
# e.g. 2056 (EPSG:2056 / CH1903+ / LV95)
Projection: 3857

# advanced map attributes (optional)
# ----------------------------------

# layers to hide (see service capabilities for possible values)
# e.g. hide admin borders: admin-low-zoom,admin-mid-zoom,admin-high-zoom,admin-text
# e.g. hide nature reserve borders: nature-reserve-boundaries,nature-reserve-text
# e.g. hide tourism borders (theme park, zoo): tourism-boundary
# e.g. hide highway shields: roads-text-ref-low-zoom,roads-text-ref
HideLayers: admin-low-zoom,admin-mid-zoom,admin-high-zoom,admin-text

# user defined objects (optional, draw order remains)
# ---------------------------------------------------
#
# data object defined by ...
# style: object style
# srs: spatial reference system (is always '+init=epsg:4326' for gpx and kml)
# type: type of data source (ogr, shape, gdal, csv)
# file: name of data objects file
# layer: data layer to extract (only required for ogr)
#
# item object defined by ...
# style: object style
# well-known-text: object definition
#
# well-known-text:
#   POINT, LINESTRING, POLYGON, MULTIPOINT, MULTILINESTRING, MULTIPOLYGON
#   all values are in millimeter (reference X0 Y0: lower left map corner)
#
# font sets:
#   fontset-0: Noto Fonts normal
#   fontset-1: Noto Fonts italic
#   fontset-2: Noto Fonts bold

UserObjects:

#- Style: <LineSymbolizer stroke='firebrick' stroke-width='8' stroke-linecap='round' />
#  SRS: '+init=epsg:4326'
#  Type: ogr
#  File: mytrack.gpx
#  Layer: tracks

# scale bar (use always stroke-linecap='butt')
- Style: |
         <LineSymbolizer stroke='dimgray' stroke-width='4.0' stroke-linecap='butt' />
         <TextSymbolizer fontset-name='fontset-2' size='12' fill='dimgray' halo-radius='1' halo-fill='rgba(255, 255, 255, 0.6)' placement='line' dy='-6' allow-overlap='true'>[name]</TextSymbolizer>
  SRS: '+init=epsg:4326'
  Type: ogr
  File: scalebar-1000.geojson
  Layer: OGRGeoJSON

# frame
- Style: <PolygonSymbolizer fill='white' fill-opacity='1.0' /> 
  WellKnownText: POLYGON((0.0 0.0, 0.0 594.0, 420.0 594.0, 420.0 0.0, 0.0 0.0), (20.0 20.0, 20.0 574.0, 400.0 574.0, 400.0 20.0, 20.0 20.0))

# border (around map area)
- Style: <LineSymbolizer stroke='dimgray' stroke-width='1.0' stroke-linecap='square' />
  WellKnownText: LINESTRING(20.0 20.0, 20.0 574.0, 400.0 574.0, 400.0 20.0, 20.0 20.0)

# crop marks (only the half line width is visible)
- Style: <LineSymbolizer stroke='dimgray' stroke-width='1.5' stroke-linecap='square' />
  WellKnownText: MULTILINESTRING((5.0 0.0, 0.0 0.0, 0.0 5.0), (5.0 594.0, 0.0 594.0, 0.0 589.0), (415.0 594.0, 420.0 594.0, 420.0 589.0), (415.0 0.0, 420.0 0.0, 420.0 5.0))

# title
- Style: <TextSymbolizer fontset-name='fontset-2' size='150' fill='dimgray' opacity='0.3' allow-overlap='true'>'H A M B U R G'</TextSymbolizer>
  WellKnownText: POINT(210.0 510.0)

# copyright
- Style: <TextSymbolizer fontset-name='fontset-0' size='12' fill='dimgray' orientation='90' allow-overlap='true'>'© OpenStreetMap contributors'</TextSymbolizer>
  WellKnownText: POINT(10.0 297)

# user files to upload
# --------------------

UserFiles:
#- mytrack.gpx
- scalebar-1000.geojson
`
