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
	"time"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	yaml "gopkg.in/yaml.v2"
)

// general program info
var (
	progName    = os.Args[0]
	progVersion = "0.2.0"
	progDate    = "2018/10/24"
	progPurpose = "Printmaps Command Line Interface Client"
	progInfo    = "Creates large-sized maps in print quality."
)

// MapConfig represents the map configuration
type MapConfig struct {
	ServiceURL  string   `yaml:"ServiceURL"`
	Metadata    Metadata `yaml:"Metadata,inline"`
	UploadFiles []string `yaml:"UserFiles"`
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
var netClient = &http.Client{
	Timeout: time.Second * 180,
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
	// fmt.Printf("mapConfig = %#v\n", mapConfig)

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
		checkMapDefinitionFile()
		fetch(action)
	} else if action == "template" {
		template()
	} else if action == "rectangle" {
		rectangle()
	} else if action == "coordgrid" {
		coordgrid()
	} else {
		fmt.Printf("action <%v> not supported\n", action)
	}

	fmt.Printf("\n")
}

/*
checkMapDefinitionFile checks if the map definition file exists
*/
func checkMapDefinitionFile() {

	if _, err := os.Stat(mapDefinitionFile); os.IsNotExist(err) {
		fmt.Printf("\nWARNING - PRECONDITION FAILED:\n")
		fmt.Printf("- the map definition file <%s> doesn't exists\n", mapDefinitionFile)
		fmt.Printf("- apply the 'template' action to create the file\n\n")
	}
}

/*
checkMapIDFile checks if the map id file exists
*/
func checkMapIDFile() {

	if _, err := os.Stat(mapIDFile); os.IsNotExist(err) {
		fmt.Printf("\nWARNING - PRECONDITION FAILED:\n")
		fmt.Printf("- the map ID file <%s> doesn't exists\n", mapIDFile)
		fmt.Printf("- apply the 'create' action to create the file\n\n")
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
	fmt.Printf("  Helper    : template, rectangle, coordgrid\n")

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
	fmt.Printf("  rectangle    : calculates wkt rectangle from base values\n")
	fmt.Printf("  coordgrid    : calculates coordinate grid in GeoJSON format\n")

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
		fmt.Printf("nothing to do ... map ID file '%s' exists\n", mapIDFile)
		return
	}

	pmData := PrintmapsData{}
	pmData.Data.Type = "maps"
	pmData.Data.ID = mapID
	pmData.Data.Attributes = mapConfig.Metadata

	requestURL := mapConfig.ServiceURL + "metadata"

	data, err := json.MarshalIndent(pmData, indentPrefix, indexString)
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

	pmDataResponse := PrintmapsData{}
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

	pmData := PrintmapsData{}
	pmData.Data.Type = "maps"
	pmData.Data.ID = mapID
	pmData.Data.Attributes = mapConfig.Metadata

	requestURL := mapConfig.ServiceURL + "metadata"

	data, err := json.MarshalIndent(pmData, indentPrefix, indexString)
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

	filename := "printmap.zip"
	requestURL := mapConfig.ServiceURL + "mapfile/" + mapID

	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("error <%v> at os.Create(), file = <%s>", err, filename)
	}
	defer file.Close()

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

	printResponse(resp, false)

	filesize := float64(resp.ContentLength) / (1024.0 * 1024.0)
	fmt.Printf("downloading file '%s' (%.1f MB) ... ", filename, filesize)

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
		fmt.Printf("nothing to do ... map ID empty\n")
		return
	}

	// delete server data
	fmt.Printf("deleting server data ...\n")

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
	fmt.Printf("\nremoving local map ID file '%s' ... ", mapIDFile)
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

	fmt.Printf("creating map definition file '%s' ... ", mapDefinitionFile)
	if _, err := os.Stat(mapDefinitionFile); err == nil {
		fmt.Printf("nothing done ... map definition file '%s' already exists\n", mapDefinitionFile)
	} else {
		err := ioutil.WriteFile(mapDefinitionFile, []byte(mapTemplate), 0666)
		if err != nil {
			log.Fatalf("error <%v> at ioutil.WriteFile(), file = <%s>", err, mapDefinitionFile)
		}
		fmt.Printf("done\n")
	}
}

/*
rectangle calculates well-known-text rectangles from base values
*/
func rectangle() {

	if len(os.Args) != 7 {
		fmt.Printf("Usage:\n")
		fmt.Printf("  %s rectangle  x    y    width  height size\n", progName)
		fmt.Printf("\nExample:\n")
		fmt.Printf("  %s rectangle  0.0  0.0  420.0  594.0  20.0\n", progName)
		return
	}

	x, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		fmt.Printf("\nerror: x (%s) not a float value\n", os.Args[2])
		return
	}
	y, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		fmt.Printf("\nerror: y (%s) not a float value\n", os.Args[3])
		return
	}
	width, err := strconv.ParseFloat(os.Args[4], 64)
	if err != nil {
		fmt.Printf("\nerror: width (%s) not a float value\n", os.Args[4])
		return
	}
	height, err := strconv.ParseFloat(os.Args[5], 64)
	if err != nil {
		fmt.Printf("\nerror: height (%s) not a float value\n", os.Args[5])
		return
	}
	size, err := strconv.ParseFloat(os.Args[6], 64)
	if err != nil {
		fmt.Printf("\nerror: (frame) size (%s) not a float value\n", os.Args[6])
		return
	}

	outerLowerLeftX := x
	outerLowerLeftY := y
	outerUpperLeftX := x
	outerUpperLeftY := y + height
	outerUpperRightX := x + width
	outerUpperRightY := outerUpperLeftY
	outerLowerRightX := outerUpperRightX
	outerLowerRightY := outerLowerLeftY

	innerLowerLeftX := x + size
	innerLowerLeftY := y + size
	innerUpperLeftX := x + size
	innerUpperLeftY := y + height - size
	innerUpperRightX := x + width - size
	innerUpperRightY := innerUpperLeftY
	innerLowerRightX := innerUpperRightX
	innerLowerRightY := innerLowerLeftY

	// as polygon with hole
	fmt.Printf("\nwell-known-text (wkt) as rectangle with hole (frame) (probably what you want) ...\n")
	fmt.Printf("POLYGON((%.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f), (%.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f))\n",
		outerLowerLeftX, outerLowerLeftY,
		outerUpperLeftX, outerUpperLeftY,
		outerUpperRightX, outerUpperRightY,
		outerLowerRightX, outerLowerRightY,
		outerLowerLeftX, outerLowerLeftY,
		innerLowerLeftX, innerLowerLeftY,
		innerUpperLeftX, innerUpperLeftY,
		innerUpperRightX, innerUpperRightY,
		innerLowerRightX, innerLowerRightY,
		innerLowerLeftX, innerLowerLeftY)

	// as two rectangles
	fmt.Printf("\nwell-known-text (wkt) as rectangle outlines (outer and inner) ...\n")
	fmt.Printf("LINESTRING(%.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f)\n",
		outerLowerLeftX, outerLowerLeftY,
		outerUpperLeftX, outerUpperLeftY,
		outerUpperRightX, outerUpperRightY,
		outerLowerRightX, outerLowerRightY,
		outerLowerLeftX, outerLowerLeftY)
	fmt.Printf("LINESTRING(%.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f, %.1f %.1f)\n",
		innerLowerLeftX, innerLowerLeftY,
		innerUpperLeftX, innerUpperLeftY,
		innerUpperRightX, innerUpperRightY,
		innerLowerRightX, innerLowerRightY,
		innerLowerLeftX, innerLowerLeftY)
}

/*
coordgrid calculates a coordinate grid and saves it as GeoJSON files
*/
func coordgrid() {

	if len(os.Args) != 7 {
		fmt.Printf("Usage:\n")
		fmt.Printf("  %s coordgrid  latmin  lonmin  latmax  lonmax  distance\n", progName)
		fmt.Printf("\nExample:\n")
		fmt.Printf("  %s coordgrid  53.4    9.9     53.6    10.1    0.01\n", progName)
		fmt.Printf("\nRemark:\n")
		fmt.Printf("  fractional digits of coord labels = fractional digits of grid distance\n")
		return
	}

	latMin, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		fmt.Printf("\nerror: latmin (%s) not a float value\n", os.Args[2])
		return
	}
	lonMin, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		fmt.Printf("\nerror: lonmin (%s) not a float value\n", os.Args[3])
		return
	}
	latMax, err := strconv.ParseFloat(os.Args[4], 64)
	if err != nil {
		fmt.Printf("\nerror: latmax (%s) not a float value\n", os.Args[4])
		return
	}
	lonMax, err := strconv.ParseFloat(os.Args[5], 64)
	if err != nil {
		fmt.Printf("\nerror: lonmax (%s) not a float value\n", os.Args[5])
		return
	}
	gridDistance, err := strconv.ParseFloat(os.Args[6], 64)
	if err != nil {
		fmt.Printf("\nerror: distance (%s) not a float value\n", os.Args[6])
		return
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
		fmt.Printf("\nerror: invalid coord parameters\n")
		return
	}

	distanceIntegerPart, _ := strconv.Atoi(tmp[0])
	distanceFractionalPart := 0
	if len(tmp) > 1 {
		distanceFractionalPart, _ = strconv.Atoi(tmp[1])
	}
	if distanceIntegerPart == 0 && distanceFractionalPart == 0 {
		fmt.Printf("\nerror: invalid distance\n")
		return
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
		log.Fatalf("error <%v> at json.MarshalIndent()", err)
	}

	// write data ([]byte) to file
	filename := "latgrid.geojson"
	if err := ioutil.WriteFile(filename, dataJSON, 0666); err != nil {
		log.Fatalf("error <%v> at ioutil.WriteFile(); file = <%v>", err, filename)
	}

	fmt.Printf("\nlatitude coordinate grid lines: %s\n", filename)

	// longitude grid lines (order: longitude, latitude)
	lsLon := make(orb.LineString, 0, 2)
	fcLon := geojson.NewFeatureCollection()

	for lonTemp := lonMin; lonTemp <= lonMax; lonTemp += gridDistance {
		// mls = append(mls, orb.LineString{{lonTemp, latMin}, {lonTemp, latMax}})
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
		log.Fatalf("error <%v> at json.MarshalIndent()", err)
	}

	// write data ([]byte) to file
	filename = "longrid.geojson"
	if err := ioutil.WriteFile(filename, dataJSON, 0666); err != nil {
		log.Fatalf("error <%v> at ioutil.WriteFile(); file = <%v>", err, filename)
	}

	fmt.Printf("longitude coordinate grid lines: %s\n", filename)
}

var mapTemplate = `# map definition file
# -------------------
# - do not use tabs or unnecessary white spaces
# - see also http://mapnik.org/mapnik-reference
#
# basic symbolizers:
# - LinePatternSymbolizer
# - LineSymbolizer
# - MarkersSymbolizer
# - PointSymbolizer
# - PolygonPatternSymbolizer
# - PolygonSymbolizer
# - TextSymbolizer
#
# advanced symbolizers:
# - BuildingSymbolizer
# - RasterSymbolizer
# - ShieldSymbolizer
#
# purpose : 
# author  : 
# release : 

# service configuration
# ---------------------

# URL of webservice
ServiceURL: http://printmaps-osm.de:8181/api/beta/maps/

# proxy configuration (not to be done here)
# - set the environment variable $HTTP_PROXY on your local system 
# - e.g. export HTTP_PROXY=http://user:password@proxy.server:port

# essential map attributes (required)
# -----------------------------------

# file format (e.g. png)
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

# advanced map attributes (optional)
# ----------------------------------

# layers to hide (see service capabilities for possible values)
# e.g. hide admin borders: admin-low-zoom,admin-mid-zoom,admin-high-zoom,admin-text
# e.g. hide nature reserve borders: nature-reserve-boundaries,nature-reserve-text
# e.g. hide tourism borders (theme park, zoo): tourism-boundary
# e.g. hide highway shields: roads-text-ref-low-zoom,roads-text-ref
HideLayers: admin-low-zoom,admin-mid-zoom,admin-high-zoom,admin-text

# user defined data objects (optional)
# ------------------------------------
# srs: spatial reference system (is always '+init=epsg:4326' for gpx and kml)
# type: type of data source (ogr, shape, gdal, csv)
# layer: data layer to extract (only required for ogr)

UserData:

#- Style: <LineSymbolizer stroke='firebrick' stroke-width='8' stroke-linecap='round' />
#  SRS: '+init=epsg:4326'
#  Type: ogr
#  File: mytrack.gpx
#  Layer: tracks

# user defined map elements (optional)
# ------------------------------------
# well-known-text:
#   POINT, LINESTRING, POLYGON, MULTIPOINT, MULTILINESTRING, MULTIPOLYGON
#   all values are in millimeter (reference X0 Y0: lower left map corner)
# font sets:
#   fontset-0: Noto Fonts normal
#   fontset-1: Noto Fonts italic
#   fontset-2: Noto Fonts bold

UserItems:

# frame
- Style: <PolygonSymbolizer fill='white' fill-opacity='1.0' /> 
  WellKnownText: POLYGON((0.0 0.0, 0.0 594.0, 420.0 594.0, 420.0 0.0, 0.0 0.0), (20.0 20.0, 20.0 574.0, 400.0 574.0, 400.0 20.0, 20.0 20.0))

# border
- Style: <LineSymbolizer stroke='black' stroke-width='3' stroke-linecap='square' />
  WellKnownText: LINESTRING(20.0 20.0, 20.0 574.0, 400.0 574.0, 400.0 20.0, 20.0 20.0)

# title
- Style: <TextSymbolizer fontset-name='fontset-2' size='150' fill='firebrick' opacity='0.3' allow-overlap='true'>'H A M B U R G'</TextSymbolizer>
  WellKnownText: POINT(210.0 510.0)

# copyright
- Style: <TextSymbolizer fontset-name='fontset-0' size='18' fill='firebrick' orientation='90' allow-overlap='true'>'© OpenStreetMap contributors'</TextSymbolizer>
  WellKnownText: POINT(10.0 297)

# scalebar label
- Style: <TextSymbolizer fontset-name='fontset-2' size='16' fill='firebrick' allow-overlap='true'>'1000 Meter'</TextSymbolizer>
  WellKnownText: POINT(42.0 36.0)

# user defined scalebar (optional)
# --------------------------------
# nature length in meter
# X and Y in millimeter

UserScalebar:
  Style: <LineSymbolizer stroke='firebrick' stroke-width='8' stroke-linecap='butt' />
  NatureLength: 1000.0
  XPos: 30.0
  YPos: 30.0 

# user files to upload
# --------------------

UserFiles:
#- mytrack.gpx
`
