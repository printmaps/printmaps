// Fetch handler

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

/*
fetchMetadata fetches the meta data for a given map ID
*/
func fetchMetadata(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {

	var pmErrorList PrintmapsErrorList
	var pmData PrintmapsData

	id := params.ByName("id")

	if err := readMetadata(&pmData, id); err != nil {
		if os.IsNotExist(err) {
			appendError(&pmErrorList, "4002", "requested ID not found: "+id, id)
		} else {
			message := fmt.Sprintf("error <%v> at readMetadata(), id = <%s>", err, id)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}
	}

	if len(pmErrorList.Errors) == 0 {
		content, err := json.MarshalIndent(pmData, indentPrefix, indexString)
		if err != nil {
			message := fmt.Sprintf("error <%v> at json.MarshalIndent()", err)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}

		writer.Header().Set("Content-Type", JSONAPIMediaType)
		writer.Header().Set("Content-Length", strconv.Itoa(len(content)))
		writer.WriteHeader(http.StatusOK)
		writer.Write(content)
	} else {
		// request not ok, response with error list
		content, err := json.MarshalIndent(pmErrorList, indentPrefix, indexString)
		if err != nil {
			message := fmt.Sprintf("error <%v> at json.MarshalIndent()", err)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}

		writer.Header().Set("Content-Type", JSONAPIMediaType)
		writer.Header().Set("Content-Length", strconv.Itoa(len(content)))
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(content)
	}
}

/*
fetchMapstate fetches the current state of the map creation process
*/
func fetchMapstate(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {

	var pmErrorList PrintmapsErrorList
	var pmState PrintmapsState

	id := params.ByName("id")

	if err := readMapstate(&pmState, id); err != nil {
		if os.IsNotExist(err) {
			appendError(&pmErrorList, "4002", "requested ID not found: "+id, id)
		} else {
			message := fmt.Sprintf("error <%v> at readMapstate(), id = <%s>", err, id)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}
	}

	if len(pmErrorList.Errors) == 0 {
		content, err := json.MarshalIndent(pmState, indentPrefix, indexString)
		if err != nil {
			message := fmt.Sprintf("error <%v> at json.MarshalIndent()", err)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}

		writer.Header().Set("Content-Type", JSONAPIMediaType)
		writer.Header().Set("Content-Length", strconv.Itoa(len(content)))
		writer.WriteHeader(http.StatusOK)
		writer.Write(content)
	} else {
		// request not ok, response with error list
		content, err := json.MarshalIndent(pmErrorList, indentPrefix, indexString)
		if err != nil {
			message := fmt.Sprintf("error <%v> at json.MarshalIndent()", err)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}

		writer.Header().Set("Content-Type", JSONAPIMediaType)
		writer.Header().Set("Content-Length", strconv.Itoa(len(content)))
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(content)
	}
}

/*
fetchMapfile send the map file with the give map ID to the client
*/
func fetchMapfile(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {

	var pmErrorList PrintmapsErrorList
	var pmData PrintmapsData
	var pmState PrintmapsState

	id := params.ByName("id")

	if err := readMetadata(&pmData, id); err != nil {
		if os.IsNotExist(err) {
			appendError(&pmErrorList, "4002", "requested ID not found: "+id, id)
		} else {
			message := fmt.Sprintf("error <%v> at readMetadata(), id = <%s>", err, id)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}
	}

	if len(pmErrorList.Errors) == 0 {
		if err := readMapstate(&pmState, id); err != nil {
			if os.IsNotExist(err) {
				appendError(&pmErrorList, "4002", "requested ID not found: "+id, id)
			} else {
				message := fmt.Sprintf("error <%v> at readMapstate(), id = <%s>", err, id)
				http.Error(writer, message, http.StatusInternalServerError)
				log.Printf("Response %d - %s", http.StatusInternalServerError, message)
				return
			}
		}
	}

	if len(pmErrorList.Errors) == 0 {
		// verify state
		if pmState.Data.Attributes.MapOrderSubmitted == "" {
			appendError(&pmErrorList, "6001", "please submit map build order first", id)
		} else if pmState.Data.Attributes.MapBuildStarted == "" {
			appendError(&pmErrorList, "6002", "please repeat download request later", id)
		} else if pmState.Data.Attributes.MapBuildCompleted == "" {
			appendError(&pmErrorList, "6003", "please repeat download request later", id)
		}
	}

	if len(pmErrorList.Errors) == 0 {
		// verify build completion (successful == yes/no)
		if pmState.Data.Attributes.MapBuildSuccessful == "no" {
			appendError(&pmErrorList, "6004", pmState.Data.Attributes.MapBuildMessage, id)
		}
	}

	if len(pmErrorList.Errors) == 0 {
		// request ok, response with mapfile
		filename := filepath.Join(PathWorkdir, PathMaps, id, FileMapfile)
		http.ServeFile(writer, request, filename)
		log.Printf("Map <%s> send to client", filename)
	} else {
		// request not ok, response with error list
		content, err := json.MarshalIndent(pmErrorList, indentPrefix, indexString)
		if err != nil {
			message := fmt.Sprintf("error <%v> at json.MarshalIndent()", err)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}

		writer.Header().Set("Content-Type", JSONAPIMediaType)
		writer.Header().Set("Content-Length", strconv.Itoa(len(content)))
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(content)
	}
}
