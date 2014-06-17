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

type GlucoseReadBatchFileWriter struct {
	destinationDirectory string
}

// NewDataStoreGlucoseReadBatchWriter creates a new GlucoseReadBatchWriter that persists to the datastore
func NewGlucoseReadBatchFileWriter(destinationDirectory string) *GlucoseReadBatchFileWriter {
	w := new(GlucoseReadBatchFileWriter)
	w.destinationDirectory = destinationDirectory
	return w
}

func (w *GlucoseReadBatchFileWriter) WriteGlucoseReadBatches(p []apimodel.DayOfGlucoseReads) (glukitio.GlucoseReadBatchWriter, error) {
	firstElement := p[0].Reads[0]
	lastBatchReads := p[len(p)-1].Reads
	lastElement := lastBatchReads[len(lastBatchReads)-1]
	outputPath := filepath.Join(w.destinationDirectory, fmt.Sprintf("glucoseReads-%s_%s.json", firstElement.GetTime().Format(TIME_FILENAME_FORMAT), lastElement.GetTime().Format(TIME_FILENAME_FORMAT)))
	f, err := os.Create(outputPath)
	if err != nil {
		return w, err
	}

	fileWriter := bufio.NewWriter(f)
	encoder := json.NewEncoder(fileWriter)
	err = encoder.Encode(p)
	if err != nil {
		return w, err
	} else {
		return w, nil
	}
}

func (w *GlucoseReadBatchFileWriter) WriteGlucoseReadBatch(p []apimodel.GlucoseRead) (glukitio.GlucoseReadBatchWriter, error) {
	dayOfGlucoseReads := make([]apimodel.DayOfGlucoseReads, 1)
	dayOfGlucoseReads[0] = apimodel.DayOfGlucoseReads{p}
	return w.WriteGlucoseReadBatches(dayOfGlucoseReads)
}

func (w *GlucoseReadBatchFileWriter) Flush() (glukitio.GlucoseReadBatchWriter, error) {
	return w, nil
}
