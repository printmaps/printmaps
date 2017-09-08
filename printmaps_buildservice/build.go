// build map

/*
  DIN  |  size in mm   | pixel at 300ppi | pixel at 72ppi
-------|---------------|-----------------|---------------
  4A0  |  1682 × 2378  |  19866 x 28087  |  4768 x 6741
  2A0  |  1189 × 1682  |  14043 x 19866  |  3370 x 4768
  A0   |   841 x 1189  |   9933 x 14043  |  2384 x 3370
  A1   |   594 x  841  |   7016 x  9933  |  1684 x 2384
  A2   |   420 x  594  |   4961 x  7016  |  1191 x 1684
  A3   |   297 x  420  |   3508 x  4961  |   842 x 1191
  A4   |   210 x  297  |   2480 x  3508  |   595 x  842
*/

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// MapnikData describes the data returned by the mapnik driver in "info mode"
type MapnikData struct {
	scale       float64
	scaleFactor float64
	BoxPixel    BoxPixel
	BoxEPSG3857 BoxEPSG3857
	BoxEPSG4326 BoxEPSG4326
	layers      string
}

/*
buildMapnikMap builds the map
*/
func buildMapnikMap(tempdir string, pmData PrintmapsData, pmState *PrintmapsState) error {

	var err error

	// find mapnik xml file
	mapnikXMLPath := ""
	mapnikXMLFile := ""
	for _, style := range config.Styles {
		if pmData.Data.Attributes.Style == style.Name {
			mapnikXMLPath = style.XMLPath
			mapnikXMLFile = style.XMLFile
			break
		}
	}
	if mapnikXMLFile == "" {
		err = errors.New("map style not found")
		log.Printf("unexpected error <%s> in buildMapnikMap(), style = <%s>", err, pmData.Data.Attributes.Style)
		return err
	}
	mapnikXML := filepath.Join(mapnikXMLPath, mapnikXMLFile)

	hideLayersFeature := ""
	if pmData.Data.Attributes.HideLayers != "" {
		hideLayersFeature = fmt.Sprintf("--hide-layers %s", pmData.Data.Attributes.HideLayers)
	}

	mapnikMapname := filepath.Join(tempdir, mapBasename+"."+pmData.Data.Attributes.Fileformat)

	pixelPerInch := 72 // pdf, svg
	if pmData.Data.Attributes.Fileformat == "png" {
		pixelPerInch = 300
	}
	// call mapnik driver in "info mode" (get the build parameters)

	command := fmt.Sprintf("%s --debug --info %s --scale %d --size %f %f --ppi %d --center %f %f %s %s",
		config.Mapnikdriver,
		hideLayersFeature, pmData.Data.Attributes.Scale,
		pmData.Data.Attributes.PrintWidth, pmData.Data.Attributes.PrintHeight,
		pixelPerInch,
		pmData.Data.Attributes.Longitude, pmData.Data.Attributes.Latitude,
		mapnikXML, mapnikMapname)

	_, commandOutput, err := runCommand(command)
	if err != nil {
		message := fmt.Sprintf("%v: %s", err, commandOutput)
		log.Printf("error <%v> at runCommand()", message)
		// the mapnik error message starts with the leading identifier "RuntimeError:"
		searchToken := "RuntimeError:"
		searchIndex := strings.Index(string(commandOutput), searchToken)
		if searchIndex != -1 {
			entries := strings.SplitAfterN(string(commandOutput), "RuntimeError:", 2)
			message = strings.TrimSpace(entries[1])
		}
		return errors.New(message)
	}

	mapnikData := MapnikData{}
	err = parseMapnikData(commandOutput, &mapnikData)
	if err != nil {
		message := fmt.Sprintf("error <%v> at parseMapnikData()", err)
		log.Printf("%s", message)
	}
	// log.Printf("mapnikData = %#v\n", mapnikData)

	// create user mapnik xml file
	mapnikXML, err = createUserMapnikXML(pmData, mapnikData)
	if err != nil {
		log.Printf("unexpected error <%s> in buildMapnikMap()", err)
		return err
	}
	if config.Testmode == false {
		defer func() {
			if err = os.Remove(mapnikXML); err != nil {
				log.Printf("unexpected error <%s> at os.Remove(), file = <%s>", err, mapnikXML)
			}
		}()
	}

	// call mapnik driver in "build mode"

	command = fmt.Sprintf("%s --debug %s --scale %d --size %f %f --ppi %d --center %f %f %s %s",
		config.Mapnikdriver,
		hideLayersFeature, pmData.Data.Attributes.Scale,
		pmData.Data.Attributes.PrintWidth, pmData.Data.Attributes.PrintHeight,
		pixelPerInch,
		pmData.Data.Attributes.Longitude, pmData.Data.Attributes.Latitude,
		mapnikXML, mapnikMapname)

	_, commandOutput, err = runCommand(command)
	if err != nil {
		message := ""
		if config.Testmode {
			message = fmt.Sprintf("%v: %s", err, commandOutput)
			log.Printf("error <%v> at runCommand()", message)
		}
		// the mapnik error message starts with the leading identifier "RuntimeError:"
		searchToken := "RuntimeError:"
		searchIndex := strings.Index(string(commandOutput), searchToken)
		if searchIndex != -1 {
			entries := strings.SplitAfterN(string(commandOutput), "RuntimeError:", 2)
			message = strings.TrimSpace(entries[1])
			// cut everything out between first and last slash (for security reasons)
			indexFirstSlash := strings.Index(message, "/")
			if indexFirstSlash != -1 {
				indexLastSlash := strings.LastIndex(message, "/")
				tempMessage := message[0:indexFirstSlash]
				message = tempMessage + message[(indexLastSlash+1):]
			}
		}
		return errors.New(message)
	}

	pmState.Data.Attributes.MapBuildBoxMillimeter.Width = pmData.Data.Attributes.PrintWidth
	pmState.Data.Attributes.MapBuildBoxMillimeter.Height = pmData.Data.Attributes.PrintHeight
	pmState.Data.Attributes.MapBuildBoxPixel = mapnikData.BoxPixel
	pmState.Data.Attributes.MapBuildBoxEPSG3857 = mapnikData.BoxEPSG3857
	pmState.Data.Attributes.MapBuildBoxEPSG4326 = mapnikData.BoxEPSG4326

	return nil
}

/*
createUserMapnikXML creates an individual user mapnik xml file
*/
func createUserMapnikXML(pmData PrintmapsData, mapnikData MapnikData) (string, error) {

	var err error

	// find mapnik xml file
	mapnikXMLPath := ""
	mapnikXMLFile := ""
	for _, style := range config.Styles {
		if pmData.Data.Attributes.Style == style.Name {
			mapnikXMLPath = style.XMLPath
			mapnikXMLFile = style.XMLFile
			break
		}
	}
	if mapnikXMLFile == "" {
		message := fmt.Sprintf("map style <%s> not found", pmData.Data.Attributes.Style)
		log.Printf("unexpected error <%s> in createUserMapnikXML()", message)
		return "", errors.New(message)
	}

	// read mapnik xml file
	filename := filepath.Join(mapnikXMLPath, mapnikXMLFile)
	mapnikLines, err := slurpFile(filename)
	if err != nil {
		message := fmt.Sprintf("error <%v> at slurpFile(); file = <%v>", err, filename)
		log.Printf("unexpected error <%s> in createUserMapnikXML()", err)
		return "", errors.New(message)
	}

	// create include section
	var includeLines []string

	// special style 'raster map'
	if pmData.Data.Attributes.Style == "raster10" {
		includeLines = createRasterMap(includeLines, mapnikData)
	}

	// user data
	includeLines = createUserData(includeLines, pmData)

	// user items
	includeLines = createUserItems(includeLines, pmData, mapnikData, pmData.Data.Attributes.PrintWidth, pmData.Data.Attributes.PrintHeight)

	// user scalebar
	includeLines = createUserScalebar(includeLines, pmData, mapnikData, pmData.Data.Attributes.PrintWidth, pmData.Data.Attributes.PrintHeight)

	// modify file references
	includeLines = modifyFileReferences(includeLines, pmData)

	// create result file
	filename = filepath.Join(mapnikXMLPath, pmData.Data.ID+"-"+mapnikXMLFile)
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		message := fmt.Sprintf("error <%v> at os.OpenFile(); file = <%s>", err, filename)
		log.Printf("unexpected error <%s> in createUserMapnikXML()", err)
		return "", errors.New(message)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Printf("unexpected error <%s> at file.Close()", err)
		}
	}()

	writer := bufio.NewWriter(file)

	for _, mapnikLine := range mapnikLines {
		if strings.Index(mapnikLine, "</Map>") != -1 {
			// insert xml rendering instructions for user defined data elements
			for _, includeLine := range includeLines {
				_, err = fmt.Fprintf(writer, "%s", includeLine)
				if err != nil {
					message := fmt.Sprintf("error <%v> at fmt.Fprintf(); file = <%v>", err, filename)
					log.Printf("unexpected error <%s> in createUserMapnikXML()", err)
					return "", errors.New(message)
				}
			}
		}
		// write line
		_, err = fmt.Fprintf(writer, "%s\n", mapnikLine)
		if err != nil {
			message := fmt.Sprintf("error <%v> at fmt.Fprintf(); file = <%v>", err, filename)
			log.Printf("unexpected error <%s> in createUserMapnikXML()", err)
			return "", errors.New(message)
		}
	}

	if err = writer.Flush(); err != nil {
		message := fmt.Sprintf("error <%v> at writer.Flush(); file = <%s>", err, filename)
		log.Printf("unexpected error <%s> in createUserMapnikXML()", err)
		return "", errors.New(message)
	}

	return filename, nil
}

/*
slurpFile slurps all lines of a text file into a slice of strings
*/
func slurpFile(filename string) ([]string, error) {

	var lines []string

	file, err := os.Open(filename)
	if err != nil {
		return lines, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	return lines, nil
}

/*
createRasterMap creates a technical map with a 10 x 10 raster
*/
func createRasterMap(lineBuffer []string, mapnikData MapnikData) []string {

	rasterName := "raster10"
	BoxEPSG3857 := mapnikData.BoxEPSG3857

	lineBuffer = append(lineBuffer, fmt.Sprintf("\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("<Style name='%s'>\n", rasterName))
	lineBuffer = append(lineBuffer, fmt.Sprintf("  <Rule>\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("    <LineSymbolizer stroke='grey' stroke-width='1' />\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("  </Rule>\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("</Style>\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("<Layer name='%s' srs='+init=epsg:3857'>\n", rasterName))
	lineBuffer = append(lineBuffer, fmt.Sprintf("  <StyleName>%s</StyleName>\n", rasterName))
	lineBuffer = append(lineBuffer, fmt.Sprintf("  <Datasource>\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("    <Parameter name='type'>csv</Parameter>\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("    <Parameter name='inline'>\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("id|name|wkt\n"))

	xInterval := (BoxEPSG3857.XMax - BoxEPSG3857.XMin) / 10.0
	yInterval := (BoxEPSG3857.YMax - BoxEPSG3857.YMin) / 10.0
	for index := 1; index < 10; index++ {
		// horizontal raster line
		horizontalLeftX := BoxEPSG3857.XMin
		horizontalLeftY := BoxEPSG3857.YMin + (float64(index) * yInterval)
		horizontalRightX := BoxEPSG3857.XMax
		horizontalRightY := BoxEPSG3857.YMin + (float64(index) * yInterval)
		// vertical raster line
		verticalLowerX := BoxEPSG3857.XMin + (float64(index) * xInterval)
		verticalLowerY := BoxEPSG3857.YMin
		verticalUpperX := BoxEPSG3857.XMin + (float64(index) * xInterval)
		verticalUpperY := BoxEPSG3857.YMax
		// create data entries
		lineBuffer = append(lineBuffer, fmt.Sprintf("%d|horizontal|LINESTRING(%f %f, %f %f)\n", index, horizontalLeftX, horizontalLeftY, horizontalRightX, horizontalRightY))
		lineBuffer = append(lineBuffer, fmt.Sprintf("%d|vertical|LINESTRING(%f %f, %f %f)\n", index, verticalLowerX, verticalLowerY, verticalUpperX, verticalUpperY))
	}

	lineBuffer = append(lineBuffer, fmt.Sprintf("    </Parameter>\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("  </Datasource>\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("</Layer>\n"))

	return lineBuffer
}

/*
createUserData creates the user defined data objects
OGR:
  <Parameter name="type">ogr</Parameter>
  <Parameter name="file">test_point_line.gpx</Parameter>
  <Parameter name="layer">waypoints</Parameter>
ShapeFile:
  <Parameter name="type">shape</Parameter>
  <Parameter name="file">/path/to/your/shapefile.shp</Parameter>
GDAL:
  <Parameter name="type">gdal</Parameter>
  <Parameter name="file">/path/to/your/data/raster.tiff</Parameter>
*/
func createUserData(lineBuffer []string, pmData PrintmapsData) []string {

	for index, userData := range pmData.Data.Attributes.UserData {
		userDataName := fmt.Sprintf("userdata-%d", index)
		lineBuffer = append(lineBuffer, fmt.Sprintf("\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("<Style name='%s'>\n", userDataName))
		lineBuffer = append(lineBuffer, fmt.Sprintf("  <Rule>\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("    %s\n", userData.Style))
		lineBuffer = append(lineBuffer, fmt.Sprintf("  </Rule>\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("</Style>\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("<Layer name='%s' srs='%s'>\n", userDataName, userData.SRS))
		lineBuffer = append(lineBuffer, fmt.Sprintf("  <StyleName>%s</StyleName>\n", userDataName))
		lineBuffer = append(lineBuffer, fmt.Sprintf("  <Datasource>\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("    <Parameter name='type'>%s</Parameter>\n", userData.Type))
		lineBuffer = append(lineBuffer, fmt.Sprintf("    <Parameter name='file'>%s</Parameter>\n", userData.File))
		if userData.Layer != "" {
			lineBuffer = append(lineBuffer, fmt.Sprintf("    <Parameter name='layer'>%s</Parameter>\n", userData.Layer))
		}
		lineBuffer = append(lineBuffer, fmt.Sprintf("  </Datasource>\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("</Layer>\n"))
	}

	return lineBuffer
}

/*
createUserItems creates the user defined annotation elementes
*/
func createUserItems(lineBuffer []string, pmData PrintmapsData, mapnikData MapnikData, width float64, height float64) []string {

	for index, userItem := range pmData.Data.Attributes.UserItems {
		userItemName := fmt.Sprintf("useritem-%d", index)
		lineBuffer = append(lineBuffer, fmt.Sprintf("\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("<Style name='%s'>\n", userItemName))
		lineBuffer = append(lineBuffer, fmt.Sprintf("  <Rule>\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("    %s\n", userItem.Style))
		lineBuffer = append(lineBuffer, fmt.Sprintf("  </Rule>\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("</Style>\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("<Layer name='%s' srs='+init=epsg:3857'>\n", userItemName))
		lineBuffer = append(lineBuffer, fmt.Sprintf("  <StyleName>%s</StyleName>\n", userItemName))
		lineBuffer = append(lineBuffer, fmt.Sprintf("  <Datasource>\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("    <Parameter name='type'>csv</Parameter>\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("    <Parameter name='inline'>\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("id|name|wkt\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("1|%s|%s\n", userItemName, transformWellKnownText(userItem.WellKnownText, mapnikData, width, height)))
		lineBuffer = append(lineBuffer, fmt.Sprintf("    </Parameter>\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("  </Datasource>\n"))
		lineBuffer = append(lineBuffer, fmt.Sprintf("</Layer>\n"))
	}

	return lineBuffer
}

/*
createUserScalebar creates the user defined scalebar
*/
func createUserScalebar(lineBuffer []string, pmData PrintmapsData, mapnikData MapnikData, width float64, height float64) []string {

	if pmData.Data.Attributes.UserScalebar.Style == "" || pmData.Data.Attributes.UserScalebar.NatureLength <= 0.0 {
		// nothing to do
		return lineBuffer
	}

	// calculate coords per millimeter
	BoxEPSG3857 := mapnikData.BoxEPSG3857
	coordsXMM := (BoxEPSG3857.XMax - BoxEPSG3857.XMin) / width
	coordsYMM := (BoxEPSG3857.YMax - BoxEPSG3857.YMin) / height

	scalebarLength := pmData.Data.Attributes.UserScalebar.NatureLength * 1000.0 / float64(pmData.Data.Attributes.Scale)
	scalebarLeftX := BoxEPSG3857.XMin + pmData.Data.Attributes.UserScalebar.XPos*coordsXMM
	scalebarLeftY := BoxEPSG3857.YMin + pmData.Data.Attributes.UserScalebar.YPos*coordsYMM
	scalebarRightX := scalebarLeftX + scalebarLength*coordsXMM
	scaleBarRightY := scalebarLeftY
	scalebarDefinition := fmt.Sprintf("LINESTRING(%f %f, %f %f)", scalebarLeftX, scalebarLeftY, scalebarRightX, scaleBarRightY)

	scalebarName := "userscalebar"
	lineBuffer = append(lineBuffer, fmt.Sprintf("\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("<Style name='%s'>\n", scalebarName))
	lineBuffer = append(lineBuffer, fmt.Sprintf("  <Rule>\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("    %s\n", pmData.Data.Attributes.UserScalebar.Style))
	lineBuffer = append(lineBuffer, fmt.Sprintf("  </Rule>\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("</Style>\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("<Layer name='%s' srs='+init=epsg:3857'>\n", scalebarName))
	lineBuffer = append(lineBuffer, fmt.Sprintf("  <StyleName>%s</StyleName>\n", scalebarName))
	lineBuffer = append(lineBuffer, fmt.Sprintf("  <Datasource>\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("    <Parameter name='type'>csv</Parameter>\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("    <Parameter name='inline'>\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("id|name|wkt\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("1|%s|%s\n", scalebarName, scalebarDefinition))
	lineBuffer = append(lineBuffer, fmt.Sprintf("    </Parameter>\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("  </Datasource>\n"))
	lineBuffer = append(lineBuffer, fmt.Sprintf("</Layer>\n"))

	return lineBuffer
}

/*
modifyFileReferences modifies all file references
*/
func modifyFileReferences(lineBuffer []string, pmData PrintmapsData) []string {

	// special handling for file path
	layerPath := filepath.Join(PathWorkdir, PathMaps, pmData.Data.ID)
	filePathDefaultMarkers := config.Markersdir
	filePathUserMarkers := filepath.Join(PathWorkdir, PathMaps, pmData.Data.ID)

	// search tokens
	layerToken := "name='file'>"
	fileToken := "file='"

	// add path
	for lineIndex, includeLine := range lineBuffer {
		searchIndex := strings.Index(includeLine, layerToken)
		if searchIndex != -1 {
			// source : <Parameter name='file'>userfile</Parameter>
			// dest   : <Parameter name='file'>/home/kto/printmaps/maps/ee493c7e-b37f-4823-936b-9b29ac7348d4/userfile</Parameter>
			entries := strings.SplitAfterN(includeLine, layerToken, 2)
			lineBuffer[lineIndex] = entries[0] + filepath.Join(layerPath, entries[1])
		}
		searchIndex = strings.Index(includeLine, fileToken)
		if searchIndex != -1 {
			entries := strings.SplitAfterN(includeLine, fileToken, 2)
			if strings.Index(entries[1], "Printmaps") == 0 {
				// source : ... file='Printmaps_Ball_Right_Red2.svg' ...
				// dest   : ... file='/home/kto/printmaps/markers/Printmaps_Ball_Right_Red2.svg' ...
				lineBuffer[lineIndex] = entries[0] + filepath.Join(filePathDefaultMarkers, entries[1])
			} else {
				// source : ... file='MyPin.svg' ...
				// dest   : ... file='/home/kto/printmaps/maps/ee493c7e-b37f-4823-936b-9b29ac7348d4/MyPin.svg' ...
				lineBuffer[lineIndex] = entries[0] + filepath.Join(filePathUserMarkers, entries[1])
			}
		}
	}

	// also possible
	layerToken = "name=\"file\">"
	fileToken = "file=\""

	// add path
	for lineIndex, includeLine := range lineBuffer {
		searchIndex := strings.Index(includeLine, layerToken)
		if searchIndex != -1 {
			// source : <Parameter name='file'>userfile</Parameter>
			// dest   : <Parameter name='file'>/home/kto/printmaps/maps/ee493c7e-b37f-4823-936b-9b29ac7348d4/userfile</Parameter>
			entries := strings.SplitAfterN(includeLine, layerToken, 2)
			lineBuffer[lineIndex] = entries[0] + filepath.Join(layerPath, entries[1])
		}
		searchIndex = strings.Index(includeLine, fileToken)
		if searchIndex != -1 {
			entries := strings.SplitAfterN(includeLine, fileToken, 2)
			if strings.Index(entries[1], "Printmaps") == 0 {
				// source : ... file='Printmaps_Ball_Right_Red2.svg' ...
				// dest   : ... file='/home/kto/printmaps/markers/Printmaps_Ball_Right_Red2.svg' ...
				lineBuffer[lineIndex] = entries[0] + filepath.Join(filePathDefaultMarkers, entries[1])
			} else {
				// source : ... file='MyPin.svg' ...
				// dest   : ... file='/home/kto/printmaps/maps/ee493c7e-b37f-4823-936b-9b29ac7348d4/MyPin.svg' ...
				lineBuffer[lineIndex] = entries[0] + filepath.Join(filePathUserMarkers, entries[1])
			}
		}
	}

	return lineBuffer
}

/*
parseMapnikData parses the output from the mapnik driver, example:
nik4-printmaps.py --debug --scale 10000 --size 297 420 --ppi 300 --center 7.6279 51.9506 mapnik.xml muenster.png
scale=1.37348285369
scale_factor=3.30760749724
size=3508,4961
bbox=Box2d(846724.854897,6787791.30045,851543.032747,6794605.14888)
bbox_wgs84=Box2d(7.60625878598,51.9317329752,7.64954121402,51.9694590903)
layers=coast-poly,waterarea,buildings,highways
*/
func parseMapnikData(commandOutput []byte, mapnikData *MapnikData) error {

	lines := strings.Split(string(commandOutput), "\n")

	i := -1
	for index, line := range lines {
		// find start line
		if strings.Index(line, "scale=") != -1 {
			i = index
			break
		}
	}

	if i == -1 {
		message := fmt.Sprintf("expected output not found")
		return errors.New(message)
	}

	// scale (inputscale * 0.00028 / scale_factor / cos(lat_center))
	_, err := fmt.Sscanf(lines[i+0], "scale=%f", &mapnikData.scale)
	if err != nil {
		message := fmt.Sprintf("error <%s> at fmt.Sscanf() (scale)", err)
		return errors.New(message)
	}

	// scale_factor (scale_factor * 90.7 = ppi)
	_, err = fmt.Sscanf(lines[i+1], "scale_factor=%f", &mapnikData.scaleFactor)
	if err != nil {
		message := fmt.Sprintf("error <%s> at fmt.Sscanf() (scale_factor)", err)
		return errors.New(message)
	}

	// size (pixel)
	_, err = fmt.Sscanf(lines[i+2], "size=%d,%d", &mapnikData.BoxPixel.Width, &mapnikData.BoxPixel.Height)
	if err != nil {
		message := fmt.Sprintf("error <%s> at fmt.Sscanf() (size)", err)
		return errors.New(message)
	}

	// bbox (google web mercator, epsg:3857)
	_, err = fmt.Sscanf(lines[i+3], "bbox=Box2d(%f,%f,%f,%f)",
		&mapnikData.BoxEPSG3857.XMin, &mapnikData.BoxEPSG3857.YMin,
		&mapnikData.BoxEPSG3857.XMax, &mapnikData.BoxEPSG3857.YMax)
	if err != nil {
		message := fmt.Sprintf("error <%s> at fmt.Sscanf() (bbox=Box2d)", err)
		return errors.New(message)
	}

	// bbox_wgs84 (wgs 84, epsg:4326)
	_, err = fmt.Sscanf(lines[i+4], "bbox_wgs84=Box2d(%f,%f,%f,%f)",
		&mapnikData.BoxEPSG4326.LonMin, &mapnikData.BoxEPSG4326.LatMin,
		&mapnikData.BoxEPSG4326.LonMax, &mapnikData.BoxEPSG4326.LatMax)
	if err != nil {
		message := fmt.Sprintf("error <%s> at fmt.Sscanf() (bbox_wgs84=Box2d)", err)
		return errors.New(message)
	}

	// layers (special case: it's possible that a style has no layer at all)
	_, err = fmt.Sscanf(lines[i+5], "layers=%s", &mapnikData.layers)
	if err != nil {
		if err != io.EOF {
			message := fmt.Sprintf("error <%s> at fmt.Sscanf() (layers)", err)
			return errors.New(message)
		}
	}

	return nil
}

/*
transformWellKnownText transforms the coordinates of an arbitrary well-know-text object
*/
func transformWellKnownText(input string, mapnikData MapnikData, width float64, height float64) string {

	output := ""
	coordinateValue := ""
	coordinateType := "x"
	coordinate := false

	// calculate coords per millimeter
	BoxEPSG3857 := mapnikData.BoxEPSG3857
	coordsXMM := (BoxEPSG3857.XMax - BoxEPSG3857.XMin) / width
	coordsYMM := (BoxEPSG3857.YMax - BoxEPSG3857.YMin) / height

	for _, runeValue := range input {
		if unicode.IsDigit(runeValue) || runeValue == '.' {
			coordinateValue += string(runeValue)
			coordinate = true
		} else {
			if coordinate {
				coordinateFloat, err := strconv.ParseFloat(coordinateValue, 64)
				if err != nil {
					log.Printf("error <%v> at strconv.ParseFloat(), value = <%s>", err, coordinateValue)
				}
				if coordinateType == "x" {
					output += fmt.Sprintf("%.1f", (BoxEPSG3857.XMin + (coordinateFloat * coordsXMM)))
					coordinateType = "y"
				} else {
					output += fmt.Sprintf("%.1f", (BoxEPSG3857.YMin + (coordinateFloat * coordsYMM)))
					coordinateType = "x"
				}
				coordinateValue = ""
				coordinate = false
			}
			output += string(runeValue)
		}
	}

	return output
}
