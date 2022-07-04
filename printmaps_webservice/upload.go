// Upload handler

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/printmaps/printmaps/pd"
)

/*
uploadUserdata allows the upload of an user data file (e.g. gpx file).
*/
func uploadUserdata(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
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

	userfileName := ""
	userfileSize := int64(-1)

	if len(pmErrorList.Errors) == 0 {
		// input file
		file, header, err := request.FormFile("file")
		if err != nil {
			fmt.Fprintln(writer, err)
			return
		}
		defer file.Close()
		_, userfileName = filepath.Split(header.Filename)

		filename := filepath.Join(pd.PathWorkdir, pd.PathMaps, pmData.Data.ID, userfileName)
		out, err := os.Create(filename)
		if err != nil {
			message := fmt.Sprintf("error <%v> at os.Create(), file = <%s>", err, filename)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}

		// write content from POST to file
		bytesWritten, err := io.Copy(out, file)
		out.Close()
		if err != nil {
			message := fmt.Sprintf("error <%v> at io.Copy(), file = <%s>", err, filename)
			http.Error(writer, message, http.StatusInternalServerError)
			log.Printf("Response %d - %s", http.StatusInternalServerError, message)
			return
		}
		userfileSize = bytesWritten

		filelimit := int64(224 * 1024 * 1024)
		removeUserfile := false
		if userfileSize > filelimit {
			log.Printf("user file <%s> (%d bytes) exceeds upload limit", filename, userfileSize)
			message := fmt.Sprintf("max upload size = %d bytes", filelimit)
			appendError(&pmErrorList, "7001", message, id)
			removeUserfile = true
		} else {
			// verify security of uploaded file
			err := verifyUploadedFile(filename)
			if err != nil {
				log.Printf("insecure user file <%s> rejected", err)
				appendError(&pmErrorList, "7002", "only data or image files are accepted", id)
				removeUserfile = true
			}
		}
		if removeUserfile {
			err := os.Remove(filename)
			if err != nil {
				log.Printf("unexpected error <%s> os.Remove(), file = <%s>", err, filename)
			}
		}
	}

	if len(pmErrorList.Errors) == 0 {
		// upload request ok (user data file created)
		writer.WriteHeader(http.StatusCreated)
		message := fmt.Sprintf("file <%s, %d bytes> successfully uploaded", userfileName, userfileSize)
		writer.Write([]byte(message))
		log.Printf("uploadUserdata(): %s", message)
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

/*
verifyUploadedFile verifies if a file is 'secure' (using linux file command).
*/
func verifyUploadedFile(filename string) error {
	command := fmt.Sprintf("file %s", filename)
	commandExitStatus, commandOutput, err := runCommand(command)
	if err != nil {
		log.Printf("error <%v> at runCommand()", err)
		log.Printf("command = <%v>", command)
		log.Printf("command exit status = <%d>", commandExitStatus)
		if len(commandOutput) > 0 {
			log.Printf("command output (stdout, stderr) =\n%s", string(commandOutput))
		}
		message := fmt.Sprintf("error <%v> at runCommand()", err)
		return errors.New(message)
	}
	if strings.Contains(string(commandOutput), "executable") {
		return errors.New(strings.TrimSpace(string(commandOutput)))
	}
	return nil
}
