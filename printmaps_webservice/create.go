// Create handler

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	uuid "github.com/satori/go.uuid"
)

/*
createMetadata creates the meta data for a new map
*/
func createMetadata(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {

	var pmErrorList PrintmapsErrorList
	var pmData PrintmapsData
	var pmState PrintmapsState

	verifyContentType(request, &pmErrorList)
	verifyAccept(request, &pmErrorList)

	// process body
	if err := json.NewDecoder(request.Body).Decode(&pmData); err != nil {
		appendError(&pmErrorList, "2001", "error = "+err.Error(), "")
	} else {
		verifyMetadata(pmData, &pmErrorList)
	}

	if len(pmErrorList.Errors) == 0 {
		// request ok, response with (new) ID and data, persist data
		universallyUniqueIdentifier, err := uuid.NewV4()
		if err != nil {
			message := fmt.Sprintf("error <%v> at uuid.NewV4()", err)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}
		pmData.Data.ID = universallyUniqueIdentifier.String()

		if err := writeMetadata(pmData); err != nil {
			message := fmt.Sprintf("error <%v> at writeMetadata()", err)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}

		content, err := json.MarshalIndent(pmData, indentPrefix, indexString)
		if err != nil {
			message := fmt.Sprintf("error <%v> at son.MarshalIndent()", err)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}

		// write state
		pmState.Data.Type = "maps"
		pmState.Data.ID = pmData.Data.ID
		pmState.Data.Attributes.MapMetadataWritten = time.Now().Format(time.RFC3339)
		if err = writeMapstate(pmState); err != nil {
			message := fmt.Sprintf("error <%v> at updateMapstate()", err)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}

		writer.Header().Set("Content-Type", JSONAPIMediaType)
		writer.Header().Set("Content-Length", strconv.Itoa(len(content)))
		writer.WriteHeader(http.StatusCreated)
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
createMapfile creates a (asynchronous) build order for the map defined in the metadata
*/
func createMapfile(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {

	var pmErrorList PrintmapsErrorList
	var pmDataPost PrintmapsData
	var pmData PrintmapsData
	var pmState PrintmapsState

	verifyContentType(request, &pmErrorList)
	verifyAccept(request, &pmErrorList)

	// process body (with map ID)
	err := json.NewDecoder(request.Body).Decode(&pmDataPost)
	if err != nil {
		appendError(&pmErrorList, "2001", "error = "+err.Error(), "")
	}

	id := pmDataPost.Data.ID

	if len(pmErrorList.Errors) == 0 {
		// request ok, read meta data from file
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
		// verify required data
		verifyRequiredMetadata(pmData, &pmErrorList)
		if len(pmErrorList.Errors) == 0 {
			// everything is ok, create build order
			if err := createMapOrder(pmData); err != nil {
				message := fmt.Sprintf("error <%v> at createMapOrder(), id = <%s>", err, id)
				http.Error(writer, message, http.StatusInternalServerError)
				log.Printf("Response %d - %s", http.StatusInternalServerError, message)
				return
			}

			// read state
			if err := readMapstate(&pmState, id); err != nil {
				if !os.IsNotExist(err) {
					message := fmt.Sprintf("error <%v> at readMapstate(), id = <%s>", err, id)
					http.Error(writer, message, http.StatusInternalServerError)
					log.Printf("Response %d - %s", http.StatusInternalServerError, message)
					return
				}
			}

			// write (update) state
			pmState.Data.Attributes.MapOrderSubmitted = time.Now().Format(time.RFC3339)
			pmState.Data.Attributes.MapBuildStarted = ""
			pmState.Data.Attributes.MapBuildCompleted = ""
			pmState.Data.Attributes.MapBuildSuccessful = ""
			pmState.Data.Attributes.MapBuildMessage = ""
			pmState.Data.Attributes.MapBuildBoxMillimeter = BoxMillimeter{}
			pmState.Data.Attributes.MapBuildBoxPixel = BoxPixel{}
			pmState.Data.Attributes.MapBuildBoxProjection = BoxProjection{}
			pmState.Data.Attributes.MapBuildBoxWGS84 = BoxWGS84{}
			if err = writeMapstate(pmState); err != nil {
				message := fmt.Sprintf("error <%v> at updateMapstate()", err)
				http.Error(writer, message, http.StatusInternalServerError)
				log.Printf("Response %d - %s", http.StatusInternalServerError, message)
				return
			}
		}
	}

	if len(pmErrorList.Errors) == 0 {
		// request ok, response with data
		content, err := json.MarshalIndent(pmData, indentPrefix, indexString)
		if err != nil {
			message := fmt.Sprintf("error <%v> at json.MarshalIndent()", err)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}

		writer.Header().Set("Content-Type", JSONAPIMediaType)
		writer.Header().Set("Content-Length", strconv.Itoa(len(content)))
		writer.WriteHeader(http.StatusAccepted)
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
