// Delete handler

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/printmaps/printmaps/pd"
)

/*
deleteMap deletes all data for a given map ID.
*/
func deleteMap(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
	var pmData pd.PrintmapsData
	var pmErrorList pd.PrintmapsErrorList

	id := params.ByName("id")

	// verify ID
	_, err := uuid.FromString(id)
	if err != nil {
		appendError(&pmErrorList, "4001", "error = "+err.Error(), "")
	}

	// map directory must exist
	if len(pmErrorList.Errors) == 0 {
		if !pd.IsExistMapDirectory(id) {
			appendError(&pmErrorList, "4002", "requested ID not found: "+id, id)
		}
	}

	if len(pmErrorList.Errors) == 0 {
		if err := pd.ReadMetadata(&pmData, id); err != nil {
			if os.IsNotExist(err) {
				appendError(&pmErrorList, "4002", "requested ID not found: "+id, id)
			} else {
				message := fmt.Sprintf("error <%v> at readMetadata(), id = <%s>", err, id)
				http.Error(writer, message, http.StatusInternalServerError)
				log.Printf("Response %d - %s", http.StatusInternalServerError, message)
				return
			}
		}
	}

	if len(pmErrorList.Errors) == 0 {
		// delete map directory
		path := filepath.Join(pd.PathWorkdir, pd.PathMaps, id)
		if err := os.RemoveAll(path); err != nil {
			message := fmt.Sprintf("error <%v> at os.RemoveAll(), path = <%s>", err, path)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}

		writer.WriteHeader(http.StatusNoContent)
	} else {
		// request not ok, response with error list
		content, err := json.MarshalIndent(pmErrorList, pd.IndentPrefix, pd.IndexString)
		if err != nil {
			message := fmt.Sprintf("error <%v> at json.MarshalIndent()", err)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}

		writer.Header().Set("Content-Type", pd.JSONAPIMediaType)
		writer.Header().Set("Content-Length", strconv.Itoa(len(content)))
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(content)
	}
}
