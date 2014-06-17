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
	lastBatchExercises := p[len(p)-1].Exercises
	lastElement := lastBatchExercises[len(lastBatchExercises)-1]
	outputPath := filepath.Join(w.destinationDirectory, fmt.Sprintf("exercises-%s_%s.json", firstElement.GetTime().Format(TIME_FILENAME_FORMAT), lastElement.GetTime().Format(TIME_FILENAME_FORMAT)))
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

func (w *ExerciseBatchFileWriter) WriteExerciseBatch(p []apimodel.Exercise) (glukitio.ExerciseBatchWriter, error) {
	dayOfExercises := make([]apimodel.DayOfExercises, 1)
	dayOfExercises[0] = apimodel.DayOfExercises{p}
	return w.WriteExerciseBatches(dayOfExercises)
}

func (w *ExerciseBatchFileWriter) Flush() (glukitio.ExerciseBatchWriter, error) {
	return w, nil
}
