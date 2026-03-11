package main

import (
	"fmt"
)

func main() {
	fmt.Println("\nЗапуск парсинга...")
	url, start, end := getUserInput()

	games := parseAllGamesFromURL(url, start, end, "out.json")

	fmt.Println("\n=== Результаты парсинга ===")
	//printResultMap(games)
	fmt.Printf("Всего собрано игр: %d\n", len(games))
}
