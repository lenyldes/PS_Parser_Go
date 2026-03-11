package main

import (
	"encoding/json"
	"os"
)

// saveToJSON сохраняет результат в JSON-файл
func saveToJSON(resultMap map[string]map[string]string, filename string) error {
	type Game struct {
		Name          string `json:"name"`
		Link          string `json:"link"`
		ImgURL        string `json:"imgURL"`
		OfferPrice    string `json:"offerPrice"`
		OriginalPrice string `json:"originalPrice"`
		ProductType   string `json:"productType"`
	}

	var games []Game
	for name, data := range resultMap {
		games = append(games, Game{
			Name:          name,
			Link:          data["link"],
			ImgURL:        data["imgURL"],
			OfferPrice:    data["offerPrice"],
			OriginalPrice: data["originalPrice"],
			ProductType:   data["productType"],
		})
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // форматирование с отступами
	return encoder.Encode(games)
}
