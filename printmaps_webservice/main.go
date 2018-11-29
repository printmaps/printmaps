/*
Purpose:
- Printmaps webservice (with JSON API)

Description:
- Webservice to build large printable maps based on OSM data.

Releases:
- 0.1.0 - 2017/05/23 : initial release (beta version)
- 0.1.1 - 2017/05/26 : improvements
- 0.1.2 - 2017/07/04 : problem with upload filename fixed
- 0.2.0 - 2018/04/22 : support for full planet implemented
- 0.3.0 - 2018/11/29 : service URL changed to beta2 (incompatible with beta)
                       update data issue fixed

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

Primary workflows (abstracted):
- step 1: client sends 'create' request (including map meta data)
  server verifies meta data
  server creates 'id' and 'directory'
  server resonses with 'id' and accepted 'meta data'
- step 2: client sends 'build order' request (identified by 'id')
  server verifies 'build order'
  server places 'build order' for (parallel running) build service
  server responses with 'accepted'
- build service (running as parallel process) builds the map:
  build service fetches 'build order'
  build service builds the map
  build service stores the map in 'directory'
  build service updates 'map state'
- step 3: client requests 'map state' (identified by 'id')
  server responses with 'map state'
- step 4 (if 'map state' includes 'build ok'): client requests 'download'
  server responses with 'map'

Optional workflows (abstracted):
- meta data request
  client requests 'meta data' (identified by 'id')
  server responses with 'meta data'
- meta data update
  client sends updated 'meta data' (identified by 'id')
  server responses with updated 'meta data'
- delete map request
  client request 'delete' map (meta, state, map data) (identified by 'id')
  server deletes all artifacts
  server responses with 'no content'
- service capabilities
  client requests 'service capabilities'
  server responses with 'service capabilities'
- map data capabilities (area, region)
  client requests 'map data capabilities'
  server responses with 'map data capabilities'
- service usage
  client requests 'server usage'
  server responses with 'server usage' (html)

Contact (eMail):
- printmaps.service@gmail.com

Remarks:
- Cross compilation for Linux: env GOOS=linux GOARCH=amd64 go build -v

Logging:
- The log file is intended for reading by humans.
- It only contains service state and error informations.

ToDo:
- pdf+svg support
- metrics

Links:
- http://www.printmaps-osm.de
- http://jsonapi.org
*/

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	pip "github.com/JamesMilnerUK/pip-go"
	"github.com/julienschmidt/httprouter"
	yaml "gopkg.in/yaml.v2"
)

// Config defines all program settings
type Config struct {
	Logfile         string
	Workdir         string
	Addr            string
	Capafile        string
	Polyfile        string
	Maintenancefile string
	Maintenancemode bool
}

var config Config

// general program info
var (
	progName    = os.Args[0]
	progVersion = "0.3.0"
	progDate    = "2018/11/29"
	progPurpose = "Printmaps Webservice"
	progInfo    = "Webservice to build large printable maps based on OSM data."
)

// area (polygon) and bounding box describing the available map data
var (
	pPolygon            pip.Polygon
	pPolygonBoundingBox pip.BoundingBox
)

// ConfigMapformat describes the map format
type ConfigMapformat struct {
	Type           string
	MinPrintWidth  float64
	MaxPrintWidth  float64
	MinPrintHeigth float64
	MaxPrintHeigth float64
}

// ConfigMapscale describes the map scale (details)
type ConfigMapscale struct {
	MinScale int
	MaxScale int
}

// ConfigMapdata describes the map data
type ConfigMapdata struct {
	Description  string
	MinLatitude  float64
	MaxLatitude  float64
	MinLongitude float64
	MaxLongitude float64
}

// ConfigStyle describes the map style
type ConfigStyle struct {
	Name             string
	ShortDescription string
	LongDescription  string
	Release          string
	Date             string
	Link             string
	Copyright        string
	Layers           string
}

// PrintmapsFeature decribes the capabilities of the service
type PrintmapsFeature struct {
	ConfigMapdata    ConfigMapdata
	ConfigMapformats []ConfigMapformat
	ConfigMapscale   ConfigMapscale
	ConfigStyles     []ConfigStyle
}

// general vars
var (
	pmFeature PrintmapsFeature // capabilities of this service
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

	configfile := "printmaps_webservice.yaml"
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

	log.Printf("config addr = %s", config.Addr)
	log.Printf("config capafile  = %s", config.Capafile)
	log.Printf("config polyfile = %s", config.Polyfile)
	log.Printf("config logfile = %s", config.Logfile)
	log.Printf("config maintenancefile = %s", config.Maintenancefile)
	log.Printf("config maintenancemode = %t", config.Maintenancemode)

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

	// full planet osm data (world) : config.Polyfile empty
	if config.Polyfile != "" {
		// read poly file (describing the area with map data) and determine the bounding box
		if err := readPolyfile(config.Polyfile, &pPolygon); err != nil {
			log.Fatalf("fatal error <%v> at readPolyfile(), file = <%v>", err, config.Polyfile)
		}
		pPolygonBoundingBox = pip.GetBoundingBox(pPolygon)
	}

	// read capabilities file (describing the features of this service)
	if err := readCapafile(config.Capafile, &pmFeature); err != nil {
		log.Fatalf("fatal error <%v> at readCapafile(), file = <%v>", err, config.Capafile)
	}

	if config.Polyfile != "" {
		// modify lat/lon values
		pmFeature.ConfigMapdata.MinLatitude = pPolygonBoundingBox.BottomLeft.Y
		pmFeature.ConfigMapdata.MaxLatitude = pPolygonBoundingBox.TopRight.Y
		pmFeature.ConfigMapdata.MinLongitude = pPolygonBoundingBox.BottomLeft.X
		pmFeature.ConfigMapdata.MaxLongitude = pPolygonBoundingBox.TopRight.X
	}

	log.Printf("MinLatitude = %f", pmFeature.ConfigMapdata.MinLatitude)
	log.Printf("MaxLatitude = %f", pmFeature.ConfigMapdata.MaxLatitude)
	log.Printf("MinLongitude = %f", pmFeature.ConfigMapdata.MinLongitude)
	log.Printf("MaxLongitude = %f", pmFeature.ConfigMapdata.MaxLongitude)

	router := httprouter.New()

	if config.Maintenancemode == false {
		// production mode
		// GET (fetch resource)
		router.GET("/api/beta2/maps/metadata/:id", middlewareHandler(fetchMetadata))
		router.GET("/api/beta2/maps/mapstate/:id", middlewareHandler(fetchMapstate))
		router.GET("/api/beta2/maps/mapfile/:id", middlewareHandler(fetchMapfile))

		// POST (create resource)
		router.POST("/api/beta2/maps/metadata", middlewareHandler(createMetadata))
		router.POST("/api/beta2/maps/mapfile", middlewareHandler(createMapfile))

		// PATCH (update resource)
		router.PATCH("/api/beta2/maps/metadata", middlewareHandler(updateMetadata))

		// DELETE (delete resource)
		router.DELETE("/api/beta2/maps/:id", middlewareHandler(deleteMap))

		// service / mapdata capabilities
		router.GET("/api/beta2/maps/capabilities/service", middlewareHandler(revealCapaService))
		router.GET("/api/beta2/maps/capabilities/mapdata", middlewareHandler(revealCapaMapdata))

		// upload user data file
		router.POST("/api/beta2/maps/upload/:id", middlewareHandler(uploadUserdata))
	} else {
		// maintenance mode (catches all requests)
		log.Printf("--> MAINTENANCE MODE ACTIVATED <--")
		log.Printf("--> service responses to each request with '503 Service Unavailable' and the maintenance file <--")
		router.GET("/*name", showMaintenance)
	}

	// subscribe to signals
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, syscall.SIGINT)  // kill -SIGINT pid -> interrupt
	signal.Notify(stopChan, syscall.SIGTERM) // kill -SIGTERM pid -> terminated

	pmWebservice := &http.Server{Addr: config.Addr, Handler: router}
	go func() {
		log.Printf("Listen for requests on port %s ...", config.Addr)
		if err := pmWebservice.ListenAndServe(); err != nil {
			log.Fatalf("fatal error <%v> at pmWebservice.ListenAndServe()", err)
		}
	}()

	// wait for signals
	sig := <-stopChan
	log.Printf("Signal <%v> received, shutting down %s (%s) ...", sig, progName, progPurpose)

	// shut down gracefully (wait max 5 seconds before halting)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pmWebservice.Shutdown(ctx); err != nil {
		log.Fatalf("fatal error <%v> at pmWebservice.Shutdown()", err)
	}

	log.Printf("%s (%s) gracefully shut down.", progName, progPurpose)
}

/*
middlewareHandler is a wrapper to catch all client requests
*/
func middlewareHandler(nextFunction httprouter.Handle) httprouter.Handle {

	return func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {

		// dump request before consuming
		dumpedRequest := dumpRequest(request)

		// create response recorder and delegate request to the given handle
		responseRecorder := httptest.NewRecorder()
		nextFunction(responseRecorder, request, params)

		// log responses with status code 500 ("internal server error")
		if responseRecorder.Code == http.StatusInternalServerError {
			log.Printf("Request %s", dumpedRequest)
			log.Printf("Response %s", dumpResponse(responseRecorder))
		}

		// copy everything from response recorder to actual response writer
		for key, value := range responseRecorder.HeaderMap {
			writer.Header()[key] = value
		}
		writer.WriteHeader(responseRecorder.Code)
		responseRecorder.Body.WriteTo(writer)
	}
}

/*
dumpRequest dumps a http request
*/
func dumpRequest(request *http.Request) string {

	dump, err := httputil.DumpRequest(request, true)
	if err != nil {
		return fmt.Sprintf("error <%v> at httputil.DumpRequest()", err)
	}
	return string(dump)
}

/*
dumpResponse dumps a (recorded) http response
*/
func dumpResponse(responseRecorder *httptest.ResponseRecorder) string {

	dump := fmt.Sprintf("%v (%s)\n", responseRecorder.Code, http.StatusText(responseRecorder.Code))
	for key, value := range responseRecorder.HeaderMap {
		dump += fmt.Sprintf("%s %s\n", key, value)
	}
	dump += fmt.Sprintf("\n%s", responseRecorder.Body.String())
	return dump
}

/*
showMaintenance shows the maintenance page and sends status code 503 (Service Unavailable)
*/
func showMaintenance(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {

	content, err := ioutil.ReadFile(config.Maintenancefile)
	if err != nil {
		log.Printf("error <%v> at ioutil.ReadFile(), file = <%s>", err, config.Maintenancefile)
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.Header().Set("Content-Length", strconv.Itoa(len(content)))
	writer.WriteHeader(http.StatusServiceUnavailable)
	writer.Write(content)
}

/*
	// "NotFound" handler
	router.NotFound = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		logRequest(request)
		http.Error(writer, apiUsage, 404)
	})

	// "MethodNotAllowed" handler
	router.MethodNotAllowed = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		logRequest(request)
		http.Error(writer, apiUsage, 405)
	})
*/
