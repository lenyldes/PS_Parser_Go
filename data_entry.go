package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

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
