/*
Purpose:
- generates svg patterns (hatches)

Description:
- Printmaps utility program.

Releases:
- 0.1.0 - 2017/06/02 : beta 1

Author:
- Klaus Tockloth

Copyright and license:
- Copyright (c) 2017 Klaus Tockloth
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

Contact (eMail):
- printmaps.service@gmail.com

Remarks:
- NN

Links:
- NN
*/

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

// general program info
var (
	progName    = os.Args[0]
	progVersion = "0.1.0"
	progDate    = "2017/05/30"
	progPurpose = "generates svg patterns (hatches)"
	progInfo    = "Printmaps utility program."
)

// colors
var colors = map[string]string{
	"Azure":      "#81BDE3",
	"Black":      "#313538",
	"Blue":       "#4888C2",
	"Chartreuse": "#C2D417",
	"Green":      "#51A13D",
	"Grey":       "#A3A4A6",
	"Orange":     "#F2A500",
	"Pink":       "#C763AE",
	"Red":        "#C93F4A",
	"Violet":     "#776EB8",
	"White":      "#E0E0E0",
	"Yellow":     "#FAD75F",
}

// angles
var angles = []int{0, 45, 90, 135}

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

	printUsage()
	generateLines()
	generateGrids()
	generateDots()
}

/*
printUsage prints the usage of this program
*/
func printUsage() {

	fmt.Printf("\nProgram:\n")
	fmt.Printf("  Name    : %s\n", progName)
	fmt.Printf("  Release : %s - %s\n", progVersion, progDate)
	fmt.Printf("  Purpose : %s\n", progPurpose)
	fmt.Printf("  Info    : %s\n", progInfo)
	fmt.Printf("\n")
}

/*
generateLines generates some simple svg line patterns
*/
func generateLines() {

	for name, value := range colors {
		for width := 1; width < 11; width++ {
			for _, angle := range angles {
				filename := fmt.Sprintf("./patterns/Printmaps_Hatch_%d_%s_%d.svg", angle, name, width)
				fmt.Printf("generate %s ...\n", filename)
				file, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
				if err != nil {
					log.Fatalf("error <%v> at os.OpenFile(), file = <%v>", err, filename)
				}
				writer := bufio.NewWriter(file)
				fmt.Fprintf(writer, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
				fmt.Fprintf(writer, "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" style=\"isolation:isolate\" viewBox=\"0 0 200 200\" width=\"200\" height=\"200\">\n")
				fmt.Fprintf(writer, "   <pattern id=\"hatch\" width=\"20\" height=\"20\" patternTransform=\"rotate(%d 0 0)\" patternUnits=\"userSpaceOnUse\">\n", angle)
				fmt.Fprintf(writer, "     <line x1=\"0\" y1=\"0\" x2=\"0\" y2=\"20\" style=\"stroke:%s; stroke-width:%d\"/>\n", value, width)
				fmt.Fprintf(writer, "   </pattern>\n")
				fmt.Fprintf(writer, "   <rect x=\"0\" y=\"0\" width=\"100%%\" height=\"100%%\" fill=\"url(#hatch)\"/>\n")
				fmt.Fprintf(writer, "</svg>\n")
				if err := writer.Flush(); err != nil {
					log.Fatalf("error <%v> at writer.Flush(), file = <%v>", err, filename)
				}
				if err := file.Close(); err != nil {
					log.Fatalf("error <%v> at file.Close(), file = <%v>", err, filename)
				}
			}
		}
	}
}

/*
generateGrids generates some simple svg grid patterns
*/
func generateGrids() {

	// raster
	for name, value := range colors {
		for width := 1; width < 11; width++ {
			filename := fmt.Sprintf("./patterns/Printmaps_Hatch_Raster_%s_%d.svg", name, width)
			fmt.Printf("generate %s ...\n", filename)
			file, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
			if err != nil {
				log.Fatalf("error <%v> at os.OpenFile(), file = <%v>", err, filename)
			}
			writer := bufio.NewWriter(file)
			fmt.Fprintf(writer, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
			fmt.Fprintf(writer, "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" style=\"isolation:isolate\" viewBox=\"0 0 200 200\" width=\"200\" height=\"200\">\n")
			fmt.Fprintf(writer, "   <pattern id=\"hatch1\" width=\"20\" height=\"20\" patternTransform=\"rotate(0 0 0)\" patternUnits=\"userSpaceOnUse\">\n")
			fmt.Fprintf(writer, "     <line x1=\"0\" y1=\"0\" x2=\"0\" y2=\"20\" style=\"stroke:%s; stroke-width:%d\"/>\n", value, width)
			fmt.Fprintf(writer, "   </pattern>\n")
			fmt.Fprintf(writer, "   <pattern id=\"hatch2\" width=\"20\" height=\"20\" patternTransform=\"rotate(90 0 0)\" patternUnits=\"userSpaceOnUse\">\n")
			fmt.Fprintf(writer, "     <line x1=\"0\" y1=\"0\" x2=\"0\" y2=\"20\" style=\"stroke:%s; stroke-width:%d\"/>\n", value, width)
			fmt.Fprintf(writer, "   </pattern>\n")
			fmt.Fprintf(writer, "   <rect x=\"0\" y=\"0\" width=\"100%%\" height=\"100%%\" fill=\"url(#hatch1)\"/>\n")
			fmt.Fprintf(writer, "   <rect x=\"0\" y=\"0\" width=\"100%%\" height=\"100%%\" fill=\"url(#hatch2)\"/>\n")
			fmt.Fprintf(writer, "</svg>\n")
			if err := writer.Flush(); err != nil {
				log.Fatalf("error <%v> at writer.Flush(), file = <%v>", err, filename)
			}
			if err := file.Close(); err != nil {
				log.Fatalf("error <%v> at file.Close(), file = <%v>", err, filename)
			}
		}
	}

	// cross
	for name, value := range colors {
		for width := 1; width < 11; width++ {
			filename := fmt.Sprintf("./patterns/Printmaps_Hatch_Cross_%s_%d.svg", name, width)
			fmt.Printf("generate %s ...\n", filename)
			file, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
			if err != nil {
				log.Fatalf("error <%v> at os.OpenFile(), file = <%v>", err, filename)
			}
			writer := bufio.NewWriter(file)
			fmt.Fprintf(writer, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
			fmt.Fprintf(writer, "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" style=\"isolation:isolate\" viewBox=\"0 0 200 200\" width=\"200\" height=\"200\">\n")
			fmt.Fprintf(writer, "   <pattern id=\"hatch1\" width=\"20\" height=\"20\" patternTransform=\"rotate(45 0 0)\" patternUnits=\"userSpaceOnUse\">\n")
			fmt.Fprintf(writer, "     <line x1=\"0\" y1=\"0\" x2=\"0\" y2=\"20\" style=\"stroke:%s; stroke-width:%d\"/>\n", value, width)
			fmt.Fprintf(writer, "   </pattern>\n")
			fmt.Fprintf(writer, "   <pattern id=\"hatch2\" width=\"20\" height=\"20\" patternTransform=\"rotate(135 0 0)\" patternUnits=\"userSpaceOnUse\">\n")
			fmt.Fprintf(writer, "     <line x1=\"0\" y1=\"0\" x2=\"0\" y2=\"20\" style=\"stroke:%s; stroke-width:%d\"/>\n", value, width)
			fmt.Fprintf(writer, "   </pattern>\n")
			fmt.Fprintf(writer, "   <rect x=\"0\" y=\"0\" width=\"100%%\" height=\"100%%\" fill=\"url(#hatch1)\"/>\n")
			fmt.Fprintf(writer, "   <rect x=\"0\" y=\"0\" width=\"100%%\" height=\"100%%\" fill=\"url(#hatch2)\"/>\n")
			fmt.Fprintf(writer, "</svg>\n")
			if err := writer.Flush(); err != nil {
				log.Fatalf("error <%v> at writer.Flush(), file = <%v>", err, filename)
			}
			if err := file.Close(); err != nil {
				log.Fatalf("error <%v> at file.Close(), file = <%v>", err, filename)
			}
		}
	}
}

/*
generateDots generates some simple svg dot patterns
*/
func generateDots() {

	for name, value := range colors {
		for diameter := 1; diameter < 6; diameter++ {
			filename := fmt.Sprintf("./patterns/Printmaps_Dot_%s_%d.svg", name, diameter)
			fmt.Printf("generate %s ...\n", filename)
			file, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
			if err != nil {
				log.Fatalf("error <%v> at os.OpenFile(), file = <%v>", err, filename)
			}
			writer := bufio.NewWriter(file)
			fmt.Fprintf(writer, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
			fmt.Fprintf(writer, "<svg xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" style=\"isolation:isolate\" viewBox=\"0 0 200 200\" width=\"200\" height=\"200\">\n")
			fmt.Fprintf(writer, "   <pattern id=\"dot\" width=\"20\" height=\"20\" patternUnits=\"userSpaceOnUse\">\n")
			fmt.Fprintf(writer, "     <circle cx=\"%d\" cy=\"%d\" r=\"%d\" fill=\"%s\"/>\n", diameter, diameter, diameter, value)
			fmt.Fprintf(writer, "   </pattern>\n")
			fmt.Fprintf(writer, "   <rect x=\"0\" y=\"0\" width=\"100%%\" height=\"100%%\" fill=\"url(#dot)\"/>\n")
			fmt.Fprintf(writer, "</svg>\n")
			if err := writer.Flush(); err != nil {
				log.Fatalf("error <%v> at writer.Flush(), file = <%v>", err, filename)
			}
			if err := file.Close(); err != nil {
				log.Fatalf("error <%v> at file.Close(), file = <%v>", err, filename)
			}
		}
	}
}
