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

type InjectionBatchFileWriter struct {
	destinationDirectory string
}

// NewDataStoreInjectionBatchWriter creates a new InjectionBatchWriter that persists to the datastore
func NewInjectionBatchFileWriter(destinationDirectory string) *InjectionBatchFileWriter {
	w := new(InjectionBatchFileWriter)
	w.destinationDirectory = destinationDirectory
	return w
}

func (w *InjectionBatchFileWriter) WriteInjectionBatches(p []apimodel.DayOfInjections) (glukitio.InjectionBatchWriter, error) {
	firstElement := p[0].Injections[0]
	outputPath := filepath.Join(w.destinationDirectory, fmt.Sprintf("injections-%s.json", firstElement.GetTime().Format(TIME_FILENAME_FORMAT)))
	f, err := os.Create(outputPath)
	if err != nil {
		return w, err
	}
	defer f.Close()

	fileWriter := bufio.NewWriter(f)
	defer fileWriter.Flush()

	encoder := json.NewEncoder(fileWriter)

	allInjections := p[0].Injections
	if len(p) > 1 {
		for i := range p[1:] {
			allInjections = mergeInjectionBatches(allInjections, p[i].Injections)
		}
	}

	err = encoder.Encode(allInjections)
	if err != nil {
		return w, err
	} else {
		return w, nil
	}
}

func (w *InjectionBatchFileWriter) WriteInjectionBatch(p []apimodel.Injection) (glukitio.InjectionBatchWriter, error) {
	dayOfInjections := make([]apimodel.DayOfInjections, 1)
	dayOfInjections[0] = apimodel.NewDayOfInjections(p)
	return w.WriteInjectionBatches(dayOfInjections)
}

func (w *InjectionBatchFileWriter) Flush() (glukitio.InjectionBatchWriter, error) {
	return w, nil
}

func mergeInjectionBatches(first, second []apimodel.Injection) []apimodel.Injection {
	newslice := make([]apimodel.Injection, len(first)+len(second))
	copy(newslice, first)
	copy(newslice[len(first):], second)
	return newslice
}
