// Data structures, global constants, globals variables

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// general vars
var (
	PathWorkdir string // base path of working directory (set as start option)
)

// general constants
const (
	PathMaps     = "maps"          // path of maps (relative to base path)
	PathOrders   = "orders"        // path of orders (relative to base path)
	FileMetadata = "metadata.json" // file holds meta data
	FileMapstate = "mapstate.json" // file holds map state
	FileMapfile  = "printmap.zip"  // file holds map data
)

// JSON identation constants
const (
	indentPrefix = ""
	indexString  = "    "
)

// JSONAPI constants
const (
	JSONAPIMediaType = "application/vnd.api+json; charset=utf-8"
)

// PrintmapsErrorList is used for the Printmaps error list (error response)
type PrintmapsErrorList struct {
	Errors []PrintmapsError
}

// PrintmapsError is used for a single Printmaps error
type PrintmapsError struct {
	ID     string
	Status string
	Code   string
	Source struct {
		Pointer   string `json:",omitempty"`
		Parameter string `json:",omitempty"`
	}
	Title  string `json:",omitempty"`
	Detail string `json:",omitempty"`
}

// PrintmapsData is used for the Printmaps data (response object)
type PrintmapsData struct {
	Data struct {
		Type       string
		ID         string
		Attributes Metadata
	}
}

// UserData describes an user defined data object (e.g. track, waypoint, ...)
type UserData struct {
	Style string `yaml:"Style"`
	SRS   string `yaml:"SRS"`
	Type  string `yaml:"Type"`
	File  string `yaml:"File"`
	Layer string `yaml:"Layer"`
}

// UserItem describes an user defined map object (e.g. point, polygon, ...)
type UserItem struct {
	Style         string `yaml:"Style"`
	WellKnownText string `yaml:"WellKnownText"`
}

// UserScalebar describes an user defined scalebar object
type UserScalebar struct {
	Style        string  `yaml:"Style"`
	NatureLength float64 `yaml:"NatureLength"`
	XPos         float64 `yaml:"XPos"`
	YPos         float64 `yaml:"YPos"`
}

// Metadata is used for the description of the map (what to build)
type Metadata struct {
	// essential map attributes (required)
	Fileformat  string  `yaml:"Fileformat"`
	Scale       int     `yaml:"Scale"`
	PrintWidth  float64 `yaml:"PrintWidth"`
	PrintHeight float64 `yaml:"PrintHeight"`
	Latitude    float64 `yaml:"Latitude"`
	Longitude   float64 `yaml:"Longitude"`
	Style       string  `yaml:"Style"`

	// advanced map attributes (optional)
	HideLayers string `yaml:"HideLayers"`

	// user defined data objects (optional)
	UserData []UserData `yaml:"UserData"`

	// user defined map elements (optional)
	UserItems []UserItem `yaml:"UserItems"`

	// user defined scalebar (optional)
	UserScalebar UserScalebar `yaml:"UserScalebar"`

	// uploaded user files (read-only value)
	UserFiles string `json:"UserFiles,omitempty" yaml:"-"`
}

// PrintmapsState is used for the Printmaps process state (response object)
type PrintmapsState struct {
	Data struct {
		Type       string
		ID         string
		Attributes Mapstate
	}
}

// BoxMillimeter represents the dimensions of the map in millimeters
type BoxMillimeter struct {
	Width  float64
	Height float64
}

// BoxPixel represents the dimensions of the map in pixels
type BoxPixel struct {
	Width  int
	Height int
}

// BoxEPSG3857 represents a spatial envelope (bounding box)
type BoxEPSG3857 struct {
	XMin float64
	YMin float64
	XMax float64
	YMax float64
}

// BoxEPSG4326 represents a spatial envelope (bounding box)
type BoxEPSG4326 struct {
	LonMin float64
	LatMin float64
	LonMax float64
	LatMax float64
}

// Mapstate is used to represent the current state of a map creation process
type Mapstate struct {
	MapMetadataWritten    string
	MapOrderSubmitted     string
	MapBuildStarted       string
	MapBuildCompleted     string
	MapBuildSuccessful    string
	MapBuildMessage       string
	MapBuildBoxMillimeter BoxMillimeter
	MapBuildBoxPixel      BoxPixel
	MapBuildBoxEPSG3857   BoxEPSG3857
	MapBuildBoxEPSG4326   BoxEPSG4326
}

/*
writeMetadata writes the map meta data to a file
*/
func writeMetadata(pmData PrintmapsData) error {

	// create directory if necessary
	path := filepath.Join(PathWorkdir, PathMaps, pmData.Data.ID)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			log.Printf("error <%v> at os.MkdirAll(), path = <%s>", err, path)
			return err
		}
	}

	data, err := json.MarshalIndent(pmData, indentPrefix, indexString)
	if err != nil {
		log.Printf("error <%v> at json.MarshalIndent()", err)
		return err
	}

	file := filepath.Join(PathWorkdir, PathMaps, pmData.Data.ID, FileMetadata)
	return ioutil.WriteFile(file, data, 0666)
}

/*
readMetadata reads the map meta data
*/
func readMetadata(pmData *PrintmapsData, id string) error {

	file := filepath.Join(PathWorkdir, PathMaps, id, FileMetadata)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		// log.Printf("error <%v> at ioutil.ReadFile(), file = <%s>", err, file)
		return err
	}

	err = json.Unmarshal(data, pmData)
	if err != nil {
		log.Printf("error <%v> at json.Unmarshal()", err)
		return err
	}

	// list of uploaded user files

	path := filepath.Join(PathWorkdir, PathMaps, id)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Printf("error <%v> at ioutil.ReadDir(), path = <%v>", err, path)
		return err
	}

	fileList := ""
	format := "%s,%d"
	for _, fileInfo := range files {
		if fileInfo.IsDir() == false {
			if !(fileInfo.Name() == FileMetadata || fileInfo.Name() == FileMapstate || fileInfo.Name() == FileMapfile) {
				fileList = fileList + fmt.Sprintf(format, fileInfo.Name(), fileInfo.Size())
				format = ",%s,%d"
			}
		}
	}
	if fileList != "" {
		pmData.Data.Attributes.UserFiles = fileList
	}

	return nil
}

/*
writeMapstate writes (updates) the state of the map creation process
*/
func writeMapstate(pmState PrintmapsState) error {

	// create directory if necessary
	path := filepath.Join(PathWorkdir, PathMaps, pmState.Data.ID)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			log.Printf("error <%v> at os.MkdirAll(), path = <%s>", err, path)
			return err
		}
	}

	data, err := json.MarshalIndent(pmState, indentPrefix, indexString)
	if err != nil {
		log.Printf("error <%v> at json.MarshalIndent()", err)
		return err
	}

	file := filepath.Join(PathWorkdir, PathMaps, pmState.Data.ID, FileMapstate)
	return ioutil.WriteFile(file, data, 0666)
}

/*
readMapstate reads the state of the map creation process
*/
func readMapstate(pmState *PrintmapsState, id string) error {

	file := filepath.Join(PathWorkdir, PathMaps, id, FileMapstate)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		// log.Printf("error <%v> at ioutil.ReadFile(), file = <%s>", err, file)
		return err
	}

	err = json.Unmarshal(data, pmState)
	if err != nil {
		log.Printf("error <%v> at json.Unmarshal()", err)
		return err
	}

	return nil
}

/*
createDirectories creates the necessary directories
*/
func createDirectories() {

	// create 'maps' directory if necessary
	path := filepath.Join(PathWorkdir, PathMaps)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			log.Fatalf("fatal error <%v> at os.MkdirAll(), path = <%s>", err, path)
		}
	}

	// create 'orders' directory if necessary
	path = filepath.Join(PathWorkdir, PathOrders)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			log.Fatalf("fatal error <%v> at os.MkdirAll(), path = <%s>", err, path)
		}
	}
}
