package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
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
	baseURL := url[:lastSlash]

	var allGames []GameInfo
	var mu sync.Mutex

	for batchStart := startPage; batchStart <= endPage; batchStart += 4 {
		batchEnd := batchStart + 3
		if batchEnd > endPage {
			batchEnd = endPage
		}

		var wg sync.WaitGroup
		batchEmpty := false

		for i := batchStart; i <= batchEnd; i++ {
			wg.Add(1)
			go func(pageNum int) {
				defer wg.Done()

				pageURL := fmt.Sprintf("%s/%d", baseURL, pageNum)
				games := parseGamesFromURL(pageURL)

				mu.Lock()
				if len(games) == 0 {
					batchEmpty = true
				} else {
					allGames = append(allGames, games...)
				}
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		mu.Lock()
		if err := saveToJSON(allGames, filename); err != nil {
			fmt.Printf("Ошибка сохранения после страниц %d-%d: %v\n", batchStart, batchEnd, err)
		}
		mu.Unlock()

		if batchEmpty {
			fmt.Printf("На страницах %d-%d обнаружена пустая страница. Останавливаемся.\n", batchStart, batchEnd)
			break
		}

		delay := rand.Intn(5) + 1
		fmt.Printf("Задержка перед следующим batch: %d сек.\n\n", delay)
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
	gameInfo, _ := readFromJSON(filename)
	var mu sync.Mutex

	for batchStart := 0; batchStart < len(gameInfo); batchStart += 4 {
		batchEnd := batchStart + 4
		if batchEnd > len(gameInfo) {
			batchEnd = len(gameInfo)
		}

		var wg sync.WaitGroup

		for i := batchStart; i < batchEnd; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				game := gameInfo[idx]
				url := game.Link
				resp, err := http.Get(url)
				if err != nil {
					fmt.Printf("Ошибка HTTP для игры %d: %v\n", idx+1, err)
					return
				}
				defer resp.Body.Close()

				respBody, err := goquery.NewDocumentFromReader(resp.Body)
				if err != nil {
					fmt.Printf("Ошибка парсинга для игры %d: %v\n", idx+1, err)
					return
				}

				ps5Voice := respBody.Find("[data-qa$='#ps5Voice-value']").Text()
				ps5ScreenLanguages := respBody.Find("[data-qa$='#ps5Subtitles-value']").Text()

				ps4Voice := respBody.Find("[data-qa$='#ps4Voice-value']").Text()
				ps4ScreenLanguages := respBody.Find("[data-qa$='#ps4Subtitles-value']").Text()

				voice := respBody.Find("[data-qa$='#voice-value']").Text()
				screenLanguages := respBody.Find("[data-qa$='#subtitles-value']").Text()

				mu.Lock()
				gameInfo[idx].PS5Voice = ps5Voice
				gameInfo[idx].PS5ScreenLanguages = ps5ScreenLanguages
				gameInfo[idx].PS4Voice = ps4Voice
				gameInfo[idx].PS4ScreenLanguages = ps4ScreenLanguages
				gameInfo[idx].Voice = voice
				gameInfo[idx].ScreenLanguages = screenLanguages
				fmt.Printf("Обработана игра %d/%d: %s\n", idx+1, len(gameInfo), game.Name)
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		mu.Lock()
		if err := saveToJSON(gameInfo, filename); err != nil {
			fmt.Printf("Ошибка сохранения после batch %d-%d: %v\n", batchStart+1, batchEnd, err)
		}
		mu.Unlock()

		delay := rand.Intn(5) + 1
		fmt.Printf("Задержка перед следующим batch: %d сек.\n\n", delay)
		time.Sleep(time.Duration(delay) * time.Second)
	}

	return gameInfo
}
