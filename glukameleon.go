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

const (
	GLUCOSE     = "glucose"
	INJECTION   = "injection"
	MEAL        = "meal"
	EXERCISE    = "exercise"
	CALIBRATION = "calibration"
)

func main() {
	var inputDirectory string
	var outputFormat string
	var glucosePath string
	var calibrationPath string
	var injectionPath string
	var mealPath string
	var exercisePath string

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
			convert(inputDirectory, outputFormat, glucosePath, calibrationPath, injectionPath, mealPath, exercisePath)
		},
	}

	convertCmd.Flags().StringVarP(&inputDirectory, "inputDirectory", "d", "", "Source directory to read from")
	convertCmd.Flags().StringVarP(&outputFormat, "format", "f", "", "Output format to write to")
	convertCmd.Flags().StringVarP(&glucosePath, "glucose", "g", "", "Output path for the glucose (not written is omitted)")
	convertCmd.Flags().StringVarP(&calibrationPath, "calibration", "c", "", "Output path for the calibration reads (not written is omitted)")
	convertCmd.Flags().StringVarP(&injectionPath, "injection", "i", "", "Output path for the injections (not written is omitted)")
	convertCmd.Flags().StringVarP(&mealPath, "meal", "m", "", "Output path for the meal data/carbs (not written is omitted)")
	convertCmd.Flags().StringVarP(&exercisePath, "exercise", "e", "", "Output path for the exercise data (not written is omitted)")

	glukameleonCmd.AddCommand(convertCmd)
	glukameleonCmd.Execute()
}

func convert(inputDirectory, format, glucosePath, calibrationPath, injectionPath, mealPath, exercisePath string) {
	_, err := ioutil.ReadDir(inputDirectory)
	if err != nil {
		log.Fatalf("Can't read contents of %s: %v", inputDirectory, err)
	}

	encoders := encodersForRecordTypes(glucosePath, calibrationPath, injectionPath, mealPath, exercisePath)

	convertFiles(inputDirectory, encoders)
}

func convertFiles(inputDirectory string, encoders map[string]*json.Encoder) {
	decoders := decodersForFiles(inputDirectory)

	if len(decoders) == 0 {
		logger.Printf("Nothing to do.")
		return
	}
	runConvert(decoders, encoders)
}

func encodersForRecordTypes(glucosePath, calibrationPath, injectionPath, mealPath, exercisePath string) map[string]*json.Encoder {
	var encoders = make(map[string]*json.Encoder)
	if len(glucosePath) > 0 {
		writer, err := os.Create(glucosePath)
		if err != nil {
			log.Fatalf("Can't open %s for writing: %v", glucosePath, err)
		}

		encoders[GLUCOSE] = json.NewEncoder(writer)
	}

	if len(calibrationPath) > 0 {
		writer, err := os.Create(calibrationPath)
		if err != nil {
			log.Fatalf("Can't open %s for writing: %v", calibrationPath, err)
		}
		encoders[CALIBRATION] = json.NewEncoder(writer)
	}

	if len(injectionPath) > 0 {
		writer, err := os.Create(injectionPath)
		if err != nil {
			log.Fatalf("Can't open %s for writing: %v", injectionPath, err)
		}
		encoders[INJECTION] = json.NewEncoder(writer)
	}

	if len(mealPath) > 0 {
		writer, err := os.Create(mealPath)
		if err != nil {
			log.Fatalf("Can't open %s for writing: %v", mealPath, err)
		}
		encoders[MEAL] = json.NewEncoder(writer)
	}

	if len(exercisePath) > 0 {
		writer, err := os.Create(exercisePath)
		if err != nil {
			log.Fatalf("Can't open %s for writing: %v", exercisePath, err)
		}
		encoders[EXERCISE] = json.NewEncoder(writer)
	}

	return encoders
}

func decodersForFiles(inputDirectory string) []*xml.Decoder {
	files, err := ioutil.ReadDir(inputDirectory)
	if err != nil {
		log.Fatalf("Can't read contents of %s: %v", inputDirectory, err)
	}

	decoders := make([]*xml.Decoder, 0)
	for _, f := range files {
		if !f.IsDir() {
			dec := newInputDecoder(inputDirectory, f.Name())
			if dec != nil {
				decoders = append(decoders, dec)
			}
		}
	}

	return decoders
}

func newInputDecoder(directory string, file string) *xml.Decoder {
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

func runConvert(decoders []*xml.Decoder, encoders map[string]*json.Encoder) {
	calibrations := make([]apimodel.Calibration, 0)
	glucoseReads := make([]apimodel.Glucose, 0)
	injections := make([]apimodel.Injection, 0)
	meals := make([]apimodel.Meal, 0)
	exercises := make([]apimodel.Exercise, 0)
	for _, dec := range decoders {
		for {
			// Read tokens from the XML document in a stream.
			t, _ := dec.Token()
			if t == nil {
				break
			}

			switch se := t.(type) {
			case xml.StartElement:
				switch se.Name.Local {
				case "Glucose":
					var read apimodel.Glucose
					dec.DecodeElement(&read, &se)
					glucoseReads = append(glucoseReads, read)

					break
				case "Event":
					var event apimodel.Event
					dec.DecodeElement(&event, &se)
					if event.EventType == "Carbs" {
						var carbQuantityInGrams int
						fmt.Sscanf(event.Description, "Carbs %d grams", &carbQuantityInGrams)

						carb := apimodel.Meal{apimodel.EventTimestamp{event.DisplayTime, event.InternalTime, event.EventTime}, float32(carbQuantityInGrams), 0., 0., 0.}
						meals = append(meals, carb)
					} else if event.EventType == "Insulin" {
						var insulinUnits float32
						_, err := fmt.Sscanf(event.Description, "Insulin %f units", &insulinUnits)
						if err != nil {
							util.Propagate(err)
						}

						injection := apimodel.Injection{apimodel.EventTimestamp{event.DisplayTime, event.InternalTime, event.EventTime}, float32(insulinUnits), "", ""}
						injections = append(injections, injection)

					} else if strings.HasPrefix(event.EventType, "Exercise") {
						var duration int
						var intensity string
						fmt.Sscanf(event.Description, "Exercise %s (%d minutes)", &intensity, &duration)

						exercise := apimodel.Exercise{apimodel.EventTimestamp{event.DisplayTime, event.InternalTime, event.EventTime}, duration, intensity, ""}
						exercises = append(exercises, exercise)
					}
				case "Meter":
					var c apimodel.Calibration
					dec.DecodeElement(&c, &se)
					calibrations = append(calibrations, c)

					break
				}
			}
		}
	}

	if encoder, enabled := encoders[CALIBRATION]; len(calibrations) > 0 && enabled {
		encoder.Encode(&calibrations)
	}

	if encoder, enabled := encoders[GLUCOSE]; len(glucoseReads) > 0 && enabled {
		encoder.Encode(&glucoseReads)
	}

	if encoder, enabled := encoders[INJECTION]; len(injections) > 0 && enabled {
		encoder.Encode(&injections)
	}

	if encoder, enabled := encoders[EXERCISE]; len(exercises) > 0 && enabled {
		encoder.Encode(&exercises)
	}

	if encoder, enabled := encoders[MEAL]; len(meals) > 0 && enabled {
		encoder.Encode(&meals)
	}
}
