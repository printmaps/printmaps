// Verify data

package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	pip "github.com/JamesMilnerUK/pip-go"
	"github.com/printmaps/printmaps/internal/pd"
)

/*
verifyContentType verifies the media type for header field "Content-Type"
*/
func verifyContentType(request *http.Request, pmErrorList *pd.PrintmapsErrorList) {

	mediaType := request.Header.Get("Content-Type")
	if mediaType != pd.JSONAPIMediaType {
		appendError(pmErrorList, "1001", "expected http header field = Content-Type: "+pd.JSONAPIMediaType, "")
	}
}

/*
verifyAccept verifies the media type for header field "Accept"
*/
func verifyAccept(request *http.Request, pmErrorList *pd.PrintmapsErrorList) {

	mediaType := request.Header.Get("Accept")
	if mediaType != pd.JSONAPIMediaType {
		appendError(pmErrorList, "1002", "expected http header field = Accept: "+pd.JSONAPIMediaType, "")
	}
}

/*
verifyMetadata verifies the map meta data
*/
func verifyMetadata(pmData pd.PrintmapsData, pmErrorList *pd.PrintmapsErrorList) {

	var message string
	var found bool

	if pmData.Data.Type != "maps" {
		appendError(pmErrorList, "3001", "valid value: maps", pmData.Data.ID)
	}

	// try to find the style
	if pmData.Data.Attributes.Style != "" {
		found = false
		for _, style := range pmFeature.ConfigStyles {
			if pmData.Data.Attributes.Style == style.Name {
				found = true
				break
			}
		}
		if found == false {
			var validStyles []string
			for _, style := range pmFeature.ConfigStyles {
				validStyles = append(validStyles, style.Name)
			}
			message = fmt.Sprintf("valid values: %s", strings.Join(validStyles, ", "))
			appendError(pmErrorList, "3008", message, pmData.Data.ID)
		}
	}

	// try to find the format
	var inputMapformat ConfigMapformat
	if pmData.Data.Attributes.Fileformat != "" {
		found = false
		for _, mapformat := range pmFeature.ConfigMapformats {
			if pmData.Data.Attributes.Fileformat == mapformat.Type {
				inputMapformat = mapformat
				found = true
				break
			}
		}
		if found == false {
			var validFormats []string
			for _, mapformat := range pmFeature.ConfigMapformats {
				validFormats = append(validFormats, mapformat.Type)
			}
			message = fmt.Sprintf("valid values: %s", strings.Join(validFormats, ", "))
			appendError(pmErrorList, "3002", message, pmData.Data.ID)
		}
	}

	if pmData.Data.Attributes.Scale != 0 {
		if pmData.Data.Attributes.Scale < pmFeature.ConfigMapscale.MinScale || pmData.Data.Attributes.Scale > pmFeature.ConfigMapscale.MaxScale {
			message = fmt.Sprintf("valid values: %d ... %d", pmFeature.ConfigMapscale.MinScale, pmFeature.ConfigMapscale.MaxScale)
			appendError(pmErrorList, "3003", message, pmData.Data.ID)
		}
	}

	if pmData.Data.Attributes.PrintWidth != 0 && inputMapformat.Type != "" {
		if pmData.Data.Attributes.PrintWidth < inputMapformat.MinPrintWidth || pmData.Data.Attributes.PrintWidth > inputMapformat.MaxPrintWidth {
			message = fmt.Sprintf("valid values: %.2f ... %.2f", inputMapformat.MinPrintWidth, inputMapformat.MaxPrintWidth)
			appendError(pmErrorList, "3004", message, pmData.Data.ID)
		}
	}

	if pmData.Data.Attributes.PrintHeight != 0 && inputMapformat.Type != "" {
		if pmData.Data.Attributes.PrintHeight < inputMapformat.MinPrintHeigth || pmData.Data.Attributes.PrintHeight > inputMapformat.MaxPrintHeigth {
			message = fmt.Sprintf("valid values: %.2f ... %.2f", inputMapformat.MinPrintHeigth, inputMapformat.MaxPrintHeigth)
			appendError(pmErrorList, "3005", message, pmData.Data.ID)
		}
	}

	if pmData.Data.Attributes.Latitude != 0.0 {
		// latMin := pPolygonBoundingBox.BottomLeft.Y
		// latMax := pPolygonBoundingBox.TopRight.Y
		latMin := pmFeature.ConfigMapdata.MinLatitude
		latMax := pmFeature.ConfigMapdata.MaxLatitude
		if pmData.Data.Attributes.Latitude < latMin || pmData.Data.Attributes.Latitude > latMax {
			message = fmt.Sprintf("valid values: %.2f ... %.2f", latMin, latMax)
			appendError(pmErrorList, "3006", message, pmData.Data.ID)
		}
	}

	if pmData.Data.Attributes.Longitude != 0.0 {
		// lonMin := pPolygonBoundingBox.BottomLeft.X
		// lonMax := pPolygonBoundingBox.TopRight.X
		lonMin := pmFeature.ConfigMapdata.MinLongitude
		lonMax := pmFeature.ConfigMapdata.MaxLongitude
		if pmData.Data.Attributes.Longitude < lonMin || pmData.Data.Attributes.Longitude > lonMax {
			message = fmt.Sprintf("valid values: %.2f ... %.2f", lonMin, lonMax)
			appendError(pmErrorList, "3007", message, pmData.Data.ID)
		}
	}

	// projection must be an integer
	if pmData.Data.Attributes.Projection != "" {
		_, err := strconv.Atoi(pmData.Data.Attributes.Projection)
		if err != nil {
			appendError(pmErrorList, "3014", "projection must be an integer", pmData.Data.ID)
		}
	}

	// full planet osm data (world) : config.Polyfile empty
	if config.Polyfile != "" {
		if pmData.Data.Attributes.Latitude != 0.0 || pmData.Data.Attributes.Longitude != 0.0 {
			var pP pip.Point
			pP.X = pmData.Data.Attributes.Longitude
			pP.Y = pmData.Data.Attributes.Latitude
			found = pip.PointInPolygon(pP, pPolygon)
			if found == false {
				appendError(pmErrorList, "3013", "no data available for the center position of the map", pmData.Data.ID)
			}
		}
	}
}

/*
verifyRequiredMetadata verifies (only) the existence of the required map meta data
*/
func verifyRequiredMetadata(pmData pd.PrintmapsData, pmErrorList *pd.PrintmapsErrorList) {

	var missingAttributes []string

	// required map meta data (content already validated)

	if pmData.Data.Attributes.Style == "" {
		missingAttributes = append(missingAttributes, "style")
	}
	if pmData.Data.Attributes.Fileformat == "" {
		missingAttributes = append(missingAttributes, "fileformat")
	}
	if pmData.Data.Attributes.Scale == 0 {
		missingAttributes = append(missingAttributes, "scale")
	}
	if pmData.Data.Attributes.PrintWidth == 0 {
		missingAttributes = append(missingAttributes, "printWidth")
	}
	if pmData.Data.Attributes.PrintHeight == 0 {
		missingAttributes = append(missingAttributes, "printHeigth")
	}
	if pmData.Data.Attributes.Latitude == 0.0 {
		missingAttributes = append(missingAttributes, "latitude")
	}
	if pmData.Data.Attributes.Longitude == 0.0 {
		missingAttributes = append(missingAttributes, "longitude")
	}
	if pmData.Data.Attributes.Projection == "" {
		missingAttributes = append(missingAttributes, "projection")
	}

	if len(missingAttributes) > 0 {
		detail := fmt.Sprintf("missing attribute(s): %s", strings.Join(missingAttributes, ", "))
		appendError(pmErrorList, "5001", detail, pmData.Data.ID)
	}
}

/*
appendError append an error entry to the error list
*/
func appendError(pmErrorList *pd.PrintmapsErrorList, code string, detail string, mapID string) {

	var jaError pd.PrintmapsError

	switch code {
	case "1001":
		jaError.Status = strconv.Itoa(http.StatusUnsupportedMediaType) + " " + http.StatusText(http.StatusUnsupportedMediaType)
		jaError.Source.Pointer = "Content-Type"
		jaError.Title = "missing or unexpected http header field Content-Type"
	case "1002":
		jaError.Status = strconv.Itoa(http.StatusUnsupportedMediaType) + " " + http.StatusText(http.StatusUnsupportedMediaType)
		jaError.Source.Pointer = "Accept"
		jaError.Title = "missing or unexpected http header field Accept"
	case "2001":
		jaError.Status = strconv.Itoa(http.StatusBadRequest) + " " + http.StatusText(http.StatusBadRequest)
		jaError.Source.Pointer = "body"
		jaError.Title = "missing or undecodable http body (json)"
	case "3001":
		jaError.Status = strconv.Itoa(http.StatusUnprocessableEntity) + " " + http.StatusText(http.StatusUnprocessableEntity)
		jaError.Source.Pointer = "data.type"
		jaError.Title = "invalid type"
	case "3002":
		jaError.Status = strconv.Itoa(http.StatusUnprocessableEntity) + " " + http.StatusText(http.StatusUnprocessableEntity)
		jaError.Source.Pointer = "data.attributes.fileformat"
		jaError.Title = "invalid attribute fileformat"
	case "3003":
		jaError.Status = strconv.Itoa(http.StatusUnprocessableEntity) + " " + http.StatusText(http.StatusUnprocessableEntity)
		jaError.Source.Pointer = "data.attributes.scale"
		jaError.Title = "invalid attribute scale"
	case "3004":
		jaError.Status = strconv.Itoa(http.StatusUnprocessableEntity) + " " + http.StatusText(http.StatusUnprocessableEntity)
		jaError.Source.Pointer = "data.attributes.printWidth"
		jaError.Title = "invalid attribute printWidth"
	case "3005":
		jaError.Status = strconv.Itoa(http.StatusUnprocessableEntity) + " " + http.StatusText(http.StatusUnprocessableEntity)
		jaError.Source.Pointer = "data.attributes.printHeight"
		jaError.Title = "invalid attribute printHeight"
	case "3006":
		jaError.Status = strconv.Itoa(http.StatusUnprocessableEntity) + " " + http.StatusText(http.StatusUnprocessableEntity)
		jaError.Source.Pointer = "data.attributes.latitude"
		jaError.Title = "invalid attribute latitude"
	case "3007":
		jaError.Status = strconv.Itoa(http.StatusUnprocessableEntity) + " " + http.StatusText(http.StatusUnprocessableEntity)
		jaError.Source.Pointer = "data.attributes.longitude"
		jaError.Title = "invalid attribute longitude"
	case "3008":
		jaError.Status = strconv.Itoa(http.StatusUnprocessableEntity) + " " + http.StatusText(http.StatusUnprocessableEntity)
		jaError.Source.Pointer = "data.attributes.style"
		jaError.Title = "invalid attribute style"
	case "3013":
		jaError.Status = strconv.Itoa(http.StatusUnprocessableEntity) + " " + http.StatusText(http.StatusUnprocessableEntity)
		jaError.Source.Pointer = "data.attributes.latitude and/or data.attributes.longitude"
		jaError.Title = "no map data available"
	case "3014":
		jaError.Status = strconv.Itoa(http.StatusUnprocessableEntity) + " " + http.StatusText(http.StatusUnprocessableEntity)
		jaError.Source.Pointer = "data.attributes.projection"
		jaError.Title = "invalid attribute projection"
	case "4002":
		jaError.Status = strconv.Itoa(http.StatusNotFound) + " " + http.StatusText(http.StatusNotFound)
		jaError.Source.Pointer = "id"
		jaError.Title = "id not found"
	case "5001":
		jaError.Status = strconv.Itoa(http.StatusPreconditionFailed) + " " + http.StatusText(http.StatusPreconditionFailed)
		jaError.Source.Pointer = "data.attributes"
		jaError.Title = "map build rejected, required attributes missing"
	case "6001":
		jaError.Status = strconv.Itoa(http.StatusPreconditionFailed) + " " + http.StatusText(http.StatusPreconditionFailed)
		jaError.Source.Pointer = "POST: api/beta2/maps/mapfile"
		jaError.Title = "map build order missing"
	case "6002":
		jaError.Status = strconv.Itoa(http.StatusPreconditionFailed) + " " + http.StatusText(http.StatusPreconditionFailed)
		jaError.Source.Pointer = "SERVER: asynchronous build process"
		jaError.Title = "map build not started yet"
	case "6003":
		jaError.Status = strconv.Itoa(http.StatusPreconditionFailed) + " " + http.StatusText(http.StatusPreconditionFailed)
		jaError.Source.Pointer = "SERVER: asynchronous build process"
		jaError.Title = "map build not completed yet"
	case "6004":
		jaError.Status = strconv.Itoa(http.StatusPreconditionFailed) + " " + http.StatusText(http.StatusPreconditionFailed)
		jaError.Source.Pointer = "SERVER: map build process"
		jaError.Title = "map build process not successful"
	case "7001":
		jaError.Status = strconv.Itoa(http.StatusRequestEntityTooLarge) + " " + http.StatusText(http.StatusRequestEntityTooLarge)
		jaError.Source.Pointer = "POST: api/beta2/maps/upload"
		jaError.Title = "size of uploaded file exceeds upload limit"
	case "7002":
		jaError.Status = strconv.Itoa(http.StatusUnsupportedMediaType) + " " + http.StatusText(http.StatusUnsupportedMediaType)
		jaError.Source.Pointer = "POST: api/beta2/maps/upload"
		jaError.Title = "insecure file rejected"
	default:
		jaError.Status = strconv.Itoa(http.StatusInternalServerError) + " " + http.StatusText(http.StatusInternalServerError)
		jaError.Source.Pointer = "unknown error code"
		jaError.Title = "unexpected program error"
	}

	jaError.ID = mapID
	jaError.Code = code
	jaError.Detail = detail
	pmErrorList.Errors = append(pmErrorList.Errors, jaError)
}
