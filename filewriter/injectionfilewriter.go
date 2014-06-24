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

	fileWriter := bufio.NewWriter(f)
	encoder := json.NewEncoder(fileWriter)
	err = encoder.Encode(p[0].Injections)
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
