package filewriter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/alexandre-normand/glukit/app/apimodel"
	"github.com/alexandre-normand/glukit/app/glukitio"
	"os"
	"path/filepath"
)

type CalibrationReadBatchFileWriter struct {
	destinationDirectory string
}

// NewDataStoreCalibrationReadBatchWriter creates a new CalibrationReadBatchWriter that persists to the datastore
func NewCalibrationReadBatchFileWriter(destinationDirectory string) *CalibrationReadBatchFileWriter {
	w := new(CalibrationReadBatchFileWriter)
	w.destinationDirectory = destinationDirectory
	return w
}

func (w *CalibrationReadBatchFileWriter) WriteCalibrationBatches(p []apimodel.DayOfCalibrationReads) (glukitio.CalibrationBatchWriter, error) {
	firstElement := p[0].Reads[0]
	outputPath := filepath.Join(w.destinationDirectory, fmt.Sprintf("calibrationReads-%s.json", firstElement.GetTime().Format(TIME_FILENAME_FORMAT)))
	f, err := os.Create(outputPath)
	if err != nil {
		return w, err
	}

	fileWriter := bufio.NewWriter(f)
	encoder := json.NewEncoder(fileWriter)
	err = encoder.Encode(p[0].Reads)
	if err != nil {
		return w, err
	} else {
		return w, nil
	}
}

func (w *CalibrationReadBatchFileWriter) WriteCalibrationBatch(p []apimodel.CalibrationRead) (glukitio.CalibrationBatchWriter, error) {
	dayOfCalibrationReads := make([]apimodel.DayOfCalibrationReads, 1)
	dayOfCalibrationReads[0] = apimodel.NewDayOfCalibrationReads(p)
	return w.WriteCalibrationBatches(dayOfCalibrationReads)
}

func (w *CalibrationReadBatchFileWriter) Flush() (glukitio.CalibrationBatchWriter, error) {
	return w, nil
}
