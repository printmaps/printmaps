// Capabilities handler

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/printmaps/printmaps/internal/pd"
)

/*
revealCapaService reveals the capabilities of this service
*/
func revealCapaService(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {

	content, err := json.MarshalIndent(pmFeature, pd.IndentPrefix, pd.IndexString)
	if err != nil {
		message := fmt.Sprintf("error <%v> at json.MarshalIndent()", err)
		http.Error(writer, message, http.StatusInternalServerError)
		log.Printf("Response %d - %s", http.StatusInternalServerError, message)
		return
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Header().Set("Content-Length", strconv.Itoa(len(content)))
	writer.WriteHeader(http.StatusOK)
	writer.Write(content)
}

/*
revealCapaMapdata reveals the capabilities of the mapdata
*/
func revealCapaMapdata(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {

	content, err := json.MarshalIndent(pPolygon, pd.IndentPrefix, pd.IndexString)
	if err != nil {
		message := fmt.Sprintf("error <%v> at json.MarshalIndent()", err)
		http.Error(writer, message, http.StatusInternalServerError)
		log.Printf("Response %d - %s", http.StatusInternalServerError, message)
		return
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Header().Set("Content-Length", strconv.Itoa(len(content)))
	writer.WriteHeader(http.StatusOK)
	writer.Write(content)
}
