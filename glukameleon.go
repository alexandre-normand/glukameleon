package main

import (
	"github.com/alexandre-normand/glukit/app/apimodel"
	"github.com/alexandre-normand/glukit/app/bufio"
	"github.com/voxelbrain/goptions"
)

func main() {
	options := struct {
		InputDirectory string        `goptions:"-d, --directory, required, description='Directory containing input files'"`
		Help           goptions.Help `goptions:"-h, --help, description='Show this help'"`

		goptions.Verbs
		Convert struct {
			RecordType string `goptions:"-r, --record-type, required, description='The type of record to process (calibration, glucose, injection, carb)'"`
			Format     string `goptions:"-t, --t, required, description='The type of output (json)'"`
		} `goptions:"convert"`
	}{ // Defaults
		Format: "json",
	}
	goptions.ParseAndFail(&options)

	files, err := ioutil.ReadDir(options.InputDirectory)
	if err != nil {
		log.Fatalf("Can't read contents of %s: %v", options.InputDirectory, err)
	}

	convert(options.InputDirectory, options.Convert.RecordType, options.Convert.Format)
}

func convert(intputDirectory, recordType, format string) {
	var enc Encoder
	switch format {
	case "json":
		enc := json.Encoder(os.Stdout)
		break
	}

	convertFilesToJson(inputDirectory, recordType, enc)
}

func convertFiles(inputDirectory string, recordType string, encoder Encoder) int {
	files, err := ioutil.ReadDir(inputDirectory)
	if err != nil {
		log.Fatalf("Can't read contents of %s: %v", inputDirectory, err)
	}

	for _, f := range files {
		if f.IsDir() {
			convertFilesToJson(f.Name(), w)
		}

		convertFileToJson(f.Name(), w)
	}
}

func convertFile(file string, enc Encoder) {
	files, err := ioutil.ReadDir(inputDirectory)
	var dec Decoder
	switch filepath.Ext(file.Name()) {
	case "xml":
		dec := xml.Decoder(os.File(file))
		break
	default:
		log.Printf("Skipping unknown filetype: [%s]", name)
		break
	}

	runConvert(dec, enc)
}

func runConvert(recordType string, dec xml.Decoder, enc json.Encoder) {
	for {
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()
		if t == nil {
			context.Debugf("finished reading file")
			break
		}

		// Inspect the type of the token just read.
		switch se := t.(type) {
		case xml.StartElement:
			// If we just read a StartElement token
			// ...and its name is "Glucose"
			switch se.Name.Local {
			case "Glucose":
				var read apimodel.Read
				// decode a whole chunk of following XML into the
				dec.DecodeElement(&read, &se)

				if recordType == "glucose" {
					enc.Encode(&read)
				}
				break
			case "Event":
				var event Event
				decoder.DecodeElement(&event, &se)
				internalEventTime := util.GetTimeInSeconds(event.InternalTime)

				// Skip everything that's before the last import's read time
				if internalEventTime > startTime.Unix() {
					if event.EventType == "Carbs" {
						var carbQuantityInGrams int
						fmt.Sscanf(event.Description, "Carbs %d grams", &carbQuantityInGrams)
						carb := model.Carb{model.Timestamp{event.EventTime, internalEventTime}, float32(carbQuantityInGrams), model.UNDEFINED_READ}

					} else if event.EventType == "Insulin" {
						var insulinUnits float32
						_, err := fmt.Sscanf(event.Description, "Insulin %f units", &insulinUnits)
						if err != nil {
							util.Propagate(err)
						}
						injection := model.Injection{model.Timestamp{event.EventTime, internalEventTime}, float32(insulinUnits), model.UNDEFINED_READ}

					} else if strings.HasPrefix(event.EventType, "Exercise") {
						var duration int
						var intensity string
						fmt.Sscanf(event.Description, "Exercise %s (%d minutes)", &intensity, &duration)
						exercise := model.Exercise{model.Timestamp{event.EventTime, internalEventTime}, duration, intensity}

						lastExercise = exercise
					}
				}

			case "Meter":
				var c apimodel.Calibration
				decoder.DecodeElement(&c, &se)

				if recordType == "calibration" {
					enc.Encode(&read)
				}
				break
			}
		}
	}
}
