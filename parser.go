package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// parseAllGamesFromURL парсит все страницы от start до end (номера страниц подставляются в URL)
// start и end — строки, пустые значения заменяются на "1" и "200" соответственно.
func parseAllGamesFromURL(url string, start string, end string, filename string) []GameInfo {
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

	var allGames []GameInfo

	for i := startPage; i <= endPage; i++ {
		pageURL := fmt.Sprintf("%s/%d", baseURL, i)
		//fmt.Printf("Парсинг страницы %d: %s\n", i, pageURL)

		games := parseGamesFromURL(pageURL)

		// Если на странице нет игр, считаем, что достигли конца каталога
		if len(games) == 0 {
			fmt.Printf("На странице %d не найдено игр. Останавливаемся.\n", i)
			// Сохраняем перед выходом
			if err := saveToJSON(allGames, filename); err != nil {
				fmt.Printf("Ошибка сохранения перед остановкой: %v\n", err)
			}
			break
		}

		//fmt.Println("Найдено игр:", len(games))

		// Добавляем игры к общему списку
		allGames = append(allGames, games...)

		// Сохраняем после каждой страницы
		if err := saveToJSON(allGames, filename); err != nil {
			fmt.Printf("Ошибка сохранения после страницы %d: %v\n", i, err)
		}

		delay := rand.Intn(5) + 1 // от 1 до 5 секунд
		fmt.Printf("Задержка перед следующей страницей: %d\n\n", delay)
		time.Sleep(time.Duration(delay) * time.Second)

	}

	return allGames
}

// parseGamesFromURL парсит одну страницу и возвращает массив с данными об играх
func parseGamesFromURL(url string) []GameInfo {
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

	var games []GameInfo

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

		if productType == "" {
			productType = "Full Game"
		}

		games = append(games, GameInfo{
			Name:          name,
			Link:          "https://store.playstation.com" + href,
			ImgURL:        fixImgURL,
			OfferPrice:    offerPrice,
			OriginalPrice: originalPrice,
			ProductType:   productType,
		})
	})

	return games
}

func parseVoice(filename string) []GameInfo {
	var result []GameInfo

	gameInfo, _ := readFromJSON(filename)

	for i, game := range gameInfo {
		url := game.Link
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		respBody, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			panic(err)
		}

		ps5Voice := respBody.Find("[data-qa$='#ps5Voice-value']").Text()
		ps5ScreenLanguages := respBody.Find("[data-qa$='#ps5Subtitles-value']").Text()

		ps4Voice := respBody.Find("[data-qa$='#ps4Voice-value']").Text()
		ps4ScreenLanguages := respBody.Find("[data-qa$='#ps5Subtitles-value']").Text()

		voice := respBody.Find("[data-qa$='#voice-value']").Text()
		screenLanguages := respBody.Find("[data-qa$='#subtitles-value']").Text()

		gameInfo[i].PS5Voice = ps5Voice

		gameInfo[i].PS5ScreenLanguages = ps5ScreenLanguages

		gameInfo[i].PS4Voice = ps4Voice
		gameInfo[i].PS4ScreenLanguages = ps4ScreenLanguages

		gameInfo[i].Voice = voice
		gameInfo[i].ScreenLanguages = screenLanguages
		time.Sleep(time.Duration(1) * time.Second)
	}

	saveToJSON(gameInfo, "out_2.json")

	return result
}
