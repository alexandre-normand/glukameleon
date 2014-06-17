package main

import (
	"encoding/xml"
	"fmt"
	"github.com/alexandre-normand/glukameleon/filewriter"
	"github.com/alexandre-normand/glukit/app/apimodel"
	"github.com/alexandre-normand/glukit/app/dexcomimporter"
	"github.com/alexandre-normand/glukit/app/streaming"
	"github.com/alexandre-normand/glukit/app/util"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var logger = log.New(os.Stderr, "glukameleon", log.LstdFlags)

type ByModDate []os.FileInfo

func main() {
	var inputDirectory string
	var outputDirectory string
	var daysPerFile int

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
			convert(inputDirectory, outputDirectory, daysPerFile)
		},
	}

	convertCmd.Flags().StringVarP(&inputDirectory, "inputDirectory", "i", "", "Source directory to read from")
	convertCmd.Flags().StringVarP(&outputDirectory, "outputDirectory", "o", "", "Output directory to write to")
	convertCmd.Flags().IntVarP(&daysPerFile, "daysperfile", "d", 7, "Number of days to write per output file")

	log.SetOutput(os.Stderr)
	glukameleonCmd.AddCommand(convertCmd)
	glukameleonCmd.Execute()
}

func convert(inputDirectory, outputDirectory string, daysPerFile int) {
	logger.Printf("in Convert")
	_, err := ioutil.ReadDir(inputDirectory)
	if err != nil {
		log.Fatalf("Can't read contents of %s: %v", inputDirectory, err)
	}

	runConvert(inputDirectory, outputDirectory, daysPerFile)
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

func (d ByModDate) Len() int {
	return len(d)
}

func (d ByModDate) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d ByModDate) Less(i, j int) bool {
	return d[i].ModTime().Unix() < d[j].ModTime().Unix()
}

func runConvert(inputDirectory string, outputDirectory string, daysPerFile int) error {
	calibrationFileWriter := filewriter.NewCalibrationReadBatchFileWriter(outputDirectory)
	calibrationStreamer := streaming.NewCalibrationReadStreamerDuration(calibrationFileWriter, time.Hour*24*time.Duration(daysPerFile))

	glucoseFileWriter := filewriter.NewGlucoseReadBatchFileWriter(outputDirectory)
	glucoseStreamer := streaming.NewGlucoseStreamerDuration(glucoseFileWriter, time.Hour*24*time.Duration(daysPerFile))

	injectionFileWriter := filewriter.NewInjectionBatchFileWriter(outputDirectory)
	injectionStreamer := streaming.NewInjectionStreamerDuration(injectionFileWriter, time.Hour*24*time.Duration(daysPerFile))

	mealFileWriter := filewriter.NewMealBatchFileWriter(outputDirectory)
	mealStreamer := streaming.NewMealStreamerDuration(mealFileWriter, time.Hour*24*time.Duration(daysPerFile))

	exerciseFileWriter := filewriter.NewExerciseBatchFileWriter(outputDirectory)
	exerciseStreamer := streaming.NewExerciseStreamerDuration(exerciseFileWriter, time.Hour*24*time.Duration(daysPerFile))

	dirFiles, err := ioutil.ReadDir(inputDirectory)
	if err != nil {
		return err
	}

	allFiles := make([]os.FileInfo, 0)
	for _, f := range dirFiles {
		if !f.IsDir() {
			allFiles = append(allFiles, f)
		}
	}

	sort.Sort(ByModDate(allFiles))

	for _, f := range allFiles {
		logger.Printf("Converting file [%s]...", f.Name())
		dec := newInputDecoder(inputDirectory, f.Name())
		if dec == nil {
			continue
		}
		for {
			// Read tokens from the XML document in a stream.
			t, _ := dec.Token()
			if t == nil {
				break
			}

			// Inspect the type of the token just read.
			switch se := t.(type) {
			case xml.StartElement:
				switch se.Name.Local {
				case "Glucose":
					var read dexcomimporter.Glucose
					// decode a whole chunk of following XML into the
					dec.DecodeElement(&read, &se)
					glucoseRead, err := dexcomimporter.ConvertXmlGlucoseRead(read)
					if err != nil {
						return err
					}

					if glucoseRead != nil && glucoseRead.Value > 0 {
						glucoseStreamer, err = glucoseStreamer.WriteGlucoseRead(*glucoseRead)

						if err != nil {
							return err
						}
					}
					break
				case "Event":
					var event dexcomimporter.Event
					dec.DecodeElement(&event, &se)
					internalEventTime, err := util.GetTimeUTC(event.InternalTime)
					if err != nil {
						log.Printf("Skipping [%s] event [%v], bad internal time [%s]: %v", event.EventType, event, event.InternalTime, err)
						continue
					}

					location := util.GetLocaltimeOffset(event.EventTime, internalEventTime)

					eventTime, err := util.GetTimeWithImpliedLocation(event.EventTime, location)
					if err != nil {
						log.Printf("Skipping [%s] event [%v], bad event time [%s]: %v", event.EventType, event, event.EventTime, err)
						continue
					}

					if event.EventType == "Carbs" {
						var mealQuantityInGrams int
						fmt.Sscanf(event.Description, "Carbs %d grams", &mealQuantityInGrams)

						meal := apimodel.Meal{apimodel.Time{apimodel.GetTimeMillis(eventTime), location.String()}, float32(mealQuantityInGrams), 0., 0., 0.}

						mealStreamer, err = mealStreamer.WriteMeal(meal)
						if err != nil {
							return err
						}

					} else if event.EventType == "Insulin" {
						var insulinUnits float32
						_, err := fmt.Sscanf(event.Description, "Insulin %f units", &insulinUnits)
						if err != nil {
							log.Printf("Failed to parse event as injection [%s]: %v", event.Description, err)
						} else {
							injection := apimodel.Injection{apimodel.Time{apimodel.GetTimeMillis(eventTime), location.String()}, float32(insulinUnits), "", ""}

							injectionStreamer, err = injectionStreamer.WriteInjection(injection)

							if err != nil {
								return err
							}
						}
					} else if strings.HasPrefix(event.EventType, "Exercise") {
						var duration int
						var intensity string
						fmt.Sscanf(event.Description, "Exercise %s (%d minutes)", &intensity, &duration)

						exercise := apimodel.Exercise{apimodel.Time{apimodel.GetTimeMillis(eventTime), location.String()}, duration, intensity, ""}
						exerciseStreamer, err = exerciseStreamer.WriteExercise(exercise)
						if err != nil {
							return err
						}
					}

					break
				case "Meter":
					var c dexcomimporter.Calibration
					dec.DecodeElement(&c, &se)

					if calibrationRead, err := dexcomimporter.ConvertXmlCalibrationRead(c); err != nil {
						return err
					} else {
						calibrationStreamer, err = calibrationStreamer.WriteCalibration(*calibrationRead)

						if err != nil {
							return err
						}
					}
					break
				}
			}
		}
	}

	// Close the streams and flush anything pending
	glucoseStreamer, err = glucoseStreamer.Close()
	if err != nil {
		return err
	}
	calibrationStreamer, err = calibrationStreamer.Close()
	if err != nil {
		return err
	}

	injectionStreamer, err = injectionStreamer.Close()
	if err != nil {
		return err
	}

	mealStreamer, err = mealStreamer.Close()
	if err != nil {
		return err
	}

	exerciseStreamer, err = exerciseStreamer.Close()
	if err != nil {
		return err
	}

	return nil
}
