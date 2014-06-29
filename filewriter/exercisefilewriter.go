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

type ExerciseBatchFileWriter struct {
	destinationDirectory string
}

// NewDataStoreExerciseBatchWriter creates a new ExerciseBatchWriter that persists to the datastore
func NewExerciseBatchFileWriter(destinationDirectory string) *ExerciseBatchFileWriter {
	w := new(ExerciseBatchFileWriter)
	w.destinationDirectory = destinationDirectory
	return w
}

func (w *ExerciseBatchFileWriter) WriteExerciseBatches(p []apimodel.DayOfExercises) (glukitio.ExerciseBatchWriter, error) {
	firstElement := p[0].Exercises[0]
	outputPath := filepath.Join(w.destinationDirectory, fmt.Sprintf("exercises-%s.json", firstElement.GetTime().Format(TIME_FILENAME_FORMAT)))
	f, err := os.Create(outputPath)
	if err != nil {
		return w, err
	}
	defer f.Close()

	fileWriter := bufio.NewWriter(f)
	defer fileWriter.Flush()

	encoder := json.NewEncoder(fileWriter)

	allExercises := p[0].Exercises

	if len(p) > 1 {
		for i := range p[1:] {
			allExercises = mergeExerciseBatches(allExercises, p[i].Exercises)
		}
	}

	err = encoder.Encode(allExercises)
	if err != nil {
		return w, err
	} else {
		return w, nil
	}
}

func (w *ExerciseBatchFileWriter) WriteExerciseBatch(p []apimodel.Exercise) (glukitio.ExerciseBatchWriter, error) {
	dayOfExercises := make([]apimodel.DayOfExercises, 1)
	dayOfExercises[0] = apimodel.NewDayOfExercises(p)
	return w.WriteExerciseBatches(dayOfExercises)
}

func (w *ExerciseBatchFileWriter) Flush() (glukitio.ExerciseBatchWriter, error) {
	return w, nil
}

func mergeExerciseBatches(first, second []apimodel.Exercise) []apimodel.Exercise {
	newslice := make([]apimodel.Exercise, len(first)+len(second))
	copy(newslice, first)
	copy(newslice[len(first):], second)
	return newslice
}
