package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// printResultMap выводит результаты в консоль
func printResultMap(resultMap map[string]map[string]string) {
	fieldsOrder := []string{"link", "offerPrice", "originalPrice"} // Порядок полей

	for name, data := range resultMap {
		fmt.Printf("Имя: %s\n", name)
		for _, field := range fieldsOrder {
			if value, ok := data[field]; ok {
				fmt.Printf("  %s: %s\n", field, value)
			}
		}
		fmt.Println()
	}
}

// parseAllGamesFromURL парсит все страницы от start до end (номера страниц подставляются в URL)
// start и end — строки, пустые значения заменяются на "1" и "200" соответственно.
func parseAllGamesFromURL(url string, start string, end string, filename string) map[string]map[string]string {
	// Устанавливаем значения по умолчанию
	if start == "" {
		start = "1"
	}
	if end == "" {
		end = "200"
	}

	startPage, err := strconv.Atoi(start)
	if err != nil {
		panic(fmt.Sprintf("некорректное значение start: %v", err))
	}
	endPage, err := strconv.Atoi(end)
	if err != nil {
		panic(fmt.Sprintf("некорректное значение end: %v", err))
	}

	// Определяем базовую часть URL (без номера страницы)
	lastSlash := strings.LastIndex(url, "/")
	if lastSlash == -1 {
		panic("неверный формат URL: отсутствует слеш")
	}
	baseURL := url[:lastSlash] // например: "https://.../category/3f772501-f6f8-49b7-abac-874a88ca4897"

	resultMap := make(map[string]map[string]string)

	for i := startPage; i <= endPage; i++ {
		pageURL := fmt.Sprintf("%s/%d", baseURL, i)
		//fmt.Printf("Парсинг страницы %d: %s\n", i, pageURL)

		tmpMap := parseGamesFromURL(pageURL)

		// Если на странице нет игр, считаем, что достигли конца каталога
		if len(tmpMap) == 0 {
			fmt.Printf("На странице %d не найдено игр. Останавливаемся.\n", i)
			// Сохраняем перед выходом
			if err := saveToJSON(resultMap, filename); err != nil {
				fmt.Printf("Ошибка сохранения перед остановкой: %v\n", err)
			}
			break
		}

		//fmt.Println("Найдено игр:", len(tmpMap))

		// Объединяем карты (если имя повторяется, данные перезапишутся)
		for name, data := range tmpMap {
			resultMap[name] = data
		}

		// Сохраняем после каждой страницы
		if err := saveToJSON(resultMap, filename); err != nil {
			fmt.Printf("Ошибка сохранения после страницы %d: %v\n", i, err)
		}

		delay := rand.Intn(5) + 1 // от 1 до 5 секунд
		fmt.Printf("Задержка перед следующей страницей: %d\n\n", delay)
		time.Sleep(time.Duration(delay) * time.Second)

	}

	return resultMap
}

// parseGamesFromURL парсит одну страницу и возвращает карту с данными об играх
func parseGamesFromURL(url string) map[string]map[string]string {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBody, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	productNameSelection := respBody.Find("#product-name")

	fmt.Println("На странице:", url)
	fmt.Println("Найдено элементов:", productNameSelection.Length())

	resultMap := make(map[string]map[string]string)

	productNameSelection.Each(func(i int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Text())

		// Родительская ссылка
		link := s.Closest("a")
		href, _ := link.Attr("href")

		// Получение цены, поднимаемся на уровень выше
		divClass := s.Closest("[class^='psw-product-tile']")

		//Получене ссылки на изображение, поднимаемся выше
		imgSel := divClass.Find("img")
		imgURL, _ := imgSel.Attr("src")
		fixImgURL := strings.Replace(imgURL, "?w=54&thumb=true", "?w=440", 1)

		// Получение остальных данных
		offerPrice := divClass.Find("[data-qa$='#price#display-price']").Text()
		originalPrice := divClass.Find("[data-qa$='#price#price-strikethrough']").Text()
		productType := divClass.Find("span[data-qa$='#product-type']").Text()

		resultMap[name] = make(map[string]string)
		resultMap[name]["link"] = "https://store.playstation.com" + href
		resultMap[name]["offerPrice"] = offerPrice
		resultMap[name]["originalPrice"] = originalPrice
		resultMap[name]["productType"] = productType
		resultMap[name]["imgURL"] = fixImgURL
	})

	return resultMap
}

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

// getUserInput запрашивает у пользователя URL, начальную и конечную страницы
func getUserInput() (string, string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== Парсер игр PlayStation Store ===")
	fmt.Println("Эта программа собирает данные об играх со страниц категории.")
	fmt.Println("Укажите параметры (или нажмите Enter для значений по умолчанию).")

	// Запрос URL
	fmt.Print("Введите URL (должен заканчиваться на '/1', например:\nhttps://store.playstation.com/en-in/category/3f772501-f6f8-49b7-abac-874a88ca4897/1):\n> ")
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)
	if url == "" {
		url = "https://store.playstation.com/en-in/category/3f772501-f6f8-49b7-abac-874a88ca4897/1"
		fmt.Println("Используется URL по умолчанию:", url)
	}

	// Запрос start
	fmt.Print("\nВведите номер начальной страницы (по умолчанию 1):\n> ")
	start, _ := reader.ReadString('\n')
	start = strings.TrimSpace(start)
	if start == "" {
		start = "1"
		fmt.Println("Начальная страница: 1")
	}

	// Запрос end
	fmt.Print("\nВведите номер конечной страницы (по умолчанию 200):\n> ")
	end, _ := reader.ReadString('\n')
	end = strings.TrimSpace(end)
	if end == "" {
		end = "200"
		fmt.Println("Конечная страница: 200")
	}

	return url, start, end
}

func main() {
	fmt.Println("\nЗапуск парсинга...")
	url, start, end := getUserInput()

	resultMap := parseAllGamesFromURL(url, start, end, "out.json")

	fmt.Println("\n=== Результаты парсинга ===")
	//printResultMap(resultMap)
	fmt.Printf("Всего собрано игр: %d\n", len(resultMap))
}
