package main

import (
	"encoding/json"
	"os"
)

type GameInfo struct {
	Name          string `json:"name"`
	Link          string `json:"link"`
	ImgURL        string `json:"imgURL"`
	OfferPrice    string `json:"offerPrice"`
	OriginalPrice string `json:"originalPrice"`
	ProductType   string `json:"productType"`
}

// saveToJSON сохраняет результат в JSON-файл
func saveToJSON(games []GameInfo, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // форматирование с отступами
	return encoder.Encode(games)
}

// readFromJSON читает данные из JSON-файла и возвращает массив с данными об играх
func readFromJSON(filename string) ([]GameInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var games []GameInfo
	err = json.NewDecoder(file).Decode(&games)
	if err != nil {
		return nil, err
	}

	return games, nil
}
