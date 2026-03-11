package main

import (
	"encoding/json"
	"os"
)

type Game struct {
	Name          string `json:"name"`
	Link          string `json:"link"`
	ImgURL        string `json:"imgURL"`
	OfferPrice    string `json:"offerPrice"`
	OriginalPrice string `json:"originalPrice"`
	ProductType   string `json:"productType"`
}

// saveToJSON сохраняет результат в JSON-файл
func saveToJSON(resultMap map[string]map[string]string, filename string) error {

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

// readFromJSON читает данные из JSON-файла и возвращает карту с данными об играх
func readFromJSON(filename string) (map[string]map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var games []Game
	err = json.NewDecoder(file).Decode(&games)
	if err != nil {
		return nil, err
	}

	resultMap := make(map[string]map[string]string)
	for _, game := range games {
		resultMap[game.Name] = map[string]string{
			"link":          game.Link,
			"imgURL":        game.ImgURL,
			"offerPrice":    game.OfferPrice,
			"originalPrice": game.OriginalPrice,
			"productType":   game.ProductType,
		}
	}

	return resultMap, nil
}
