// Update handler

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

/*
updateMetadata updates (patches) the meta data for a given map ID
*/
func updateMetadata(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {

	var pmErrorList PrintmapsErrorList
	var pmDataPost PrintmapsData
	var pmData PrintmapsData
	var pmState PrintmapsState

	verifyContentType(request, &pmErrorList)
	verifyAccept(request, &pmErrorList)

	bodyBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		message := fmt.Sprintf("error <%v> at ioutil.ReadAll()", err)
		http.Error(writer, message, http.StatusInternalServerError)
		log.Printf("Response %d - %s", http.StatusInternalServerError, message)
		return
	}

	if err = json.Unmarshal(bodyBytes, &pmDataPost); err != nil {
		appendError(&pmErrorList, "2001", "error = "+err.Error(), "")
	}

	id := pmDataPost.Data.ID
	userFiles := ""

	// the update data set must contains all map elements (changed + unchanged)
	if len(pmErrorList.Errors) == 0 {
		if err = json.Unmarshal(bodyBytes, &pmData); err != nil {
			appendError(&pmErrorList, "2001", "error = "+err.Error(), id)
		} else {
			verifyMetadata(pmData, &pmErrorList)
		}
	}

	if len(pmErrorList.Errors) == 0 {
		// request ok, response with updated data, persist data
		if err := writeMetadata(pmData); err != nil {
			message := fmt.Sprintf("error <%v> at writeMetadata()", err)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}

		pmData.Data.Attributes.UserFiles = userFiles
		content, err := json.MarshalIndent(pmData, indentPrefix, indexString)
		if err != nil {
			message := fmt.Sprintf("error <%v> at json.MarshalIndent()", err)
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
		pmState.Data.Attributes.MapMetadataWritten = time.Now().Format(time.RFC3339)
		pmState.Data.Attributes.MapOrderSubmitted = ""
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
