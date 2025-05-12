package services

import (
	"encoding/json"
	"io"
	"math/rand/v2"
	"os"
)

type Categories struct {
	Categories []Category `json:"categories"`
}

type Category struct {
	Name      string     `json:"name"`
	Questions []Question `json:"questions"`
}

type Question struct {
	Regular string `json:"regular"`
	Sneaky  string `json:"sneaky"`
}

var categoriesList Categories

func InitializeQuestionService() {
	jsonFile, err := os.Open("questions.json")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &categoriesList)
	if err != nil {
		panic(err)
	}
}

func selectRandomQuestion() (string, string, error) {
	randomCategoryIndex := rand.IntN(len(categoriesList.Categories) - 1)
	randomQuestionIndex := rand.IntN(len(categoriesList.Categories[randomCategoryIndex].Questions) - 1)
	randomCategory := categoriesList.Categories[randomCategoryIndex]
	randomQuestion := randomCategory.Questions[randomQuestionIndex]
	return randomQuestion.Regular, randomQuestion.Sneaky, nil
}

func selectQuestionFromCategory(categoryName string) (string, string, error) {
	for _, category := range categoriesList.Categories {
		if category.Name == categoryName {
			randomQuestionIndex := rand.IntN(len(category.Questions) - 1)
			randomQuestion := category.Questions[randomQuestionIndex]
			return randomQuestion.Regular, randomQuestion.Sneaky, nil
		}
	}
	return "", "", nil
}

func GetAvailableCategories() ([]string, error) {
	categories := make([]string, len(categoriesList.Categories))
	for i, category := range categoriesList.Categories {
		categories[i] = category.Name
	}
	return categories, nil
}
