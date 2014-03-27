package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/alexandre-normand/glukit/app/apimodel"
	"github.com/alexandre-normand/glukit/app/util"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var logger = log.New(os.Stderr, "glukameleon", log.LstdFlags)

func main() {
	var inputDirectory string
	var outputFormat string
	var recordType string

	var glukameleonCmd = &cobra.Command{
		Use:   "glukameleon",
		Short: "glukameleon allows to convert dexcom G4 data between different formats",
		Long: `A Fast and Functional converter of XML to JSON Dexcom G4 data. 
	       Code can be found at http://github.com/alexandre-normand/glukameleon`,
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}

	var convertCmd = &cobra.Command{
		Use:   "convert",
		Short: "convert takes a directory of files and converts them to the desired format in a unified file",
		Long:  `Only XML to JSON is support at the moment`,
		Run: func(cmd *cobra.Command, args []string) {
			convert(inputDirectory, recordType, outputFormat)
		},
	}

	convertCmd.Flags().StringVarP(&inputDirectory, "inputDirectory", "i", "", "Source directory to read from")
	convertCmd.Flags().StringVarP(&outputFormat, "format", "f", "", "Output format to write to")
	convertCmd.Flags().StringVarP(&recordType, "recordType", "r", "", "Record type to convert (calibration, glucose, injection, carb)")

	glukameleonCmd.AddCommand(convertCmd)
	glukameleonCmd.Execute()
}

func convert(inputDirectory, recordType, format string) {
	logger.Printf("in Convert")
	_, err := ioutil.ReadDir(inputDirectory)
	if err != nil {
		log.Fatalf("Can't read contents of %s: %v", inputDirectory, err)
	}

	var enc *json.Encoder
	switch format {
	case "json":
		enc = json.NewEncoder(os.Stdout)
		break
	}

	convertFiles(inputDirectory, recordType, enc)
}

func convertFiles(inputDirectory string, recordType string, encoder *json.Encoder) {
	logger.Printf("in ConvertFiles")
	files, err := ioutil.ReadDir(inputDirectory)
	if err != nil {
		log.Fatalf("Can't read contents of %s: %v", inputDirectory, err)
	}

	decoders := make([]*xml.Decoder, 0)
	for _, f := range files {
		if !f.IsDir() {
			dec := newInputDecoder(inputDirectory, f.Name(), recordType, encoder)
			if dec != nil {
				decoders = append(decoders, dec)
			}
		}
	}

	if len(decoders) == 0 {
		logger.Printf("Nothing to do.")
		return
	}
	runConvert(recordType, decoders, encoder)
}

func newInputDecoder(directory string, file string, recordType string, enc *json.Encoder) *xml.Decoder {
	switch ext := filepath.Ext(file); {
	case ext == ".xml":
		fileToConvert, err := os.Open(filepath.Join(directory, file))
		if err != nil {
			log.Fatalf("Can't convert file [%s]: %v", filepath.Join(directory, file), err)
		}

		return xml.NewDecoder(fileToConvert)
	default:
		logger.Printf("Skipping unknown filetype: [%s]", ext)
		return nil
	}
}

func runConvert(recordType string, decoders []*xml.Decoder, enc *json.Encoder) {
	calibrations := make([]apimodel.Calibration, 0)
	glucoseReads := make([]apimodel.Glucose, 0)
	for _, dec := range decoders {
		for {
			// Read tokens from the XML document in a stream.
			t, _ := dec.Token()
			if t == nil {
				logger.Printf("finished reading file")
				break
			}

			// Inspect the type of the token just read.
			switch se := t.(type) {
			case xml.StartElement:
				// If we just read a StartElement token
				// ...and its name is "Glucose"
				switch se.Name.Local {
				case "Glucose":
					var read apimodel.Glucose
					// decode a whole chunk of following XML into the
					dec.DecodeElement(&read, &se)

					if recordType == "glucose" {
						glucoseReads = append(glucoseReads, read)
					}
					break
				case "Event":
					var event apimodel.Event
					dec.DecodeElement(&event, &se)
					//internalEventTime := util.GetTimeInSeconds(event.InternalTime)

					// Skip everything that's before the last import's read time

					if event.EventType == "Carbs" {
						var carbQuantityInGrams int
						fmt.Sscanf(event.Description, "Carbs %d grams", &carbQuantityInGrams)

					} else if event.EventType == "Insulin" {
						var insulinUnits float32
						_, err := fmt.Sscanf(event.Description, "Insulin %f units", &insulinUnits)
						if err != nil {
							util.Propagate(err)
						}

					} else if strings.HasPrefix(event.EventType, "Exercise") {
						var duration int
						var intensity string
						fmt.Sscanf(event.Description, "Exercise %s (%d minutes)", &intensity, &duration)
					}

				case "Meter":
					var c apimodel.Calibration
					dec.DecodeElement(&c, &se)

					if recordType == "calibration" {
						calibrations = append(calibrations, c)
					}
					break
				}
			}
		}
	}

	if len(calibrations) > 0 {
		enc.Encode(&calibrations)
	}

	if len(glucoseReads) > 0 {
		enc.Encode(&glucoseReads)
	}
}
