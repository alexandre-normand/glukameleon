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
	outputPath := filepath.Join(w.destinationDirectory, fmt.Sprintf("meals-%s.json", firstElement.GetTime().Format(TIME_FILENAME_FORMAT)))
	f, err := os.Create(outputPath)
	if err != nil {
		return w, err
	}
	defer f.Close()

	fileWriter := bufio.NewWriter(f)
	defer fileWriter.Flush()

	encoder := json.NewEncoder(fileWriter)

	allMeals := p[0].Meals
	if len(p) > 1 {
		for i := range p[1:] {
			allMeals = mergeMealBatches(allMeals, p[i].Meals)
		}
	}

	err = encoder.Encode(allMeals)
	if err != nil {
		return w, err
	} else {
		return w, nil
	}
}

func (w *MealBatchFileWriter) WriteMealBatch(p []apimodel.Meal) (glukitio.MealBatchWriter, error) {
	dayOfMeals := make([]apimodel.DayOfMeals, 1)
	dayOfMeals[0] = apimodel.NewDayOfMeals(p)
	return w.WriteMealBatches(dayOfMeals)
}

func (w *MealBatchFileWriter) Flush() (glukitio.MealBatchWriter, error) {
	return w, nil
}

func mergeMealBatches(first, second []apimodel.Meal) []apimodel.Meal {
	newslice := make([]apimodel.Meal, len(first)+len(second))
	copy(newslice, first)
	copy(newslice[len(first):], second)
	return newslice
}
