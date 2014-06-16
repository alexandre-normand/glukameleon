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

type MealBatchFileWriter struct {
	destinationDirectory string
}

// NewDataStoreMealBatchWriter creates a new MealBatchWriter that persists to the datastore
func NewMealBatchFileWriter(destinationDirectory string) *MealBatchFileWriter {
	w := new(MealBatchFileWriter)
	w.destinationDirectory = destinationDirectory
	return w
}

func (w *MealBatchFileWriter) WriteMealBatches(p []apimodel.DayOfMeals) (glukitio.MealBatchWriter, error) {
	firstElement := p[0].Meals[0]
	lastBatchMeals := p[len(p)-1].Meals
	lastElement := lastBatchMeals[len(lastBatchMeals)-1]
	outputPath := filepath.Join(w.destinationDirectory, fmt.Sprintf("meals-%s_%s.json", firstElement.GetTime().Format(TIME_FILENAME_FORMAT), lastElement.GetTime().Format(TIME_FILENAME_FORMAT)))
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

func (w *MealBatchFileWriter) WriteMealBatch(p []apimodel.Meal) (glukitio.MealBatchWriter, error) {
	dayOfMeals := make([]apimodel.DayOfMeals, 1)
	dayOfMeals[0] = apimodel.DayOfMeals{p}
	return w.WriteMealBatches(dayOfMeals)
}

func (w *MealBatchFileWriter) Flush() (glukitio.MealBatchWriter, error) {
	return w, nil
}
