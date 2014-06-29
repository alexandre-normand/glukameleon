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
	outputPath := filepath.Join(w.destinationDirectory, fmt.Sprintf("glucoseReads-%s.json", firstElement.GetTime().Format(TIME_FILENAME_FORMAT)))
	f, err := os.Create(outputPath)
	if err != nil {
		return w, err
	}
	defer f.Close()

	fileWriter := bufio.NewWriter(f)
	defer fileWriter.Flush()

	encoder := json.NewEncoder(fileWriter)

	allReads := p[0].Reads

	if len(p) > 1 {
		for i := range p[1:] {
			allReads = mergeGlucoseReadBatches(allReads, p[i].Reads)
		}
	}

	err = encoder.Encode(allReads)
	if err != nil {
		return w, err
	} else {
		return w, nil
	}
}

func (w *GlucoseReadBatchFileWriter) WriteGlucoseReadBatch(p []apimodel.GlucoseRead) (glukitio.GlucoseReadBatchWriter, error) {
	dayOfGlucoseReads := make([]apimodel.DayOfGlucoseReads, 1)
	dayOfGlucoseReads[0] = apimodel.NewDayOfGlucoseReads(p)
	return w.WriteGlucoseReadBatches(dayOfGlucoseReads)
}

func (w *GlucoseReadBatchFileWriter) Flush() (glukitio.GlucoseReadBatchWriter, error) {
	return w, nil
}

func mergeGlucoseReadBatches(first, second []apimodel.GlucoseRead) []apimodel.GlucoseRead {
	newslice := make([]apimodel.GlucoseRead, len(first)+len(second))
	copy(newslice, first)
	copy(newslice[len(first):], second)
	return newslice
}
