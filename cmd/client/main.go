package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	endpoint := "http://localhost:8080/"

	// контейнер данных для запроса
	data := url.Values{}
	// приглашение в консоли
	fmt.Println("Введите длинный URL")
	// открываем потоковое чтение из консоли
	reader := bufio.NewReader(os.Stdin)
	// читаем строку из консоли
	long, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	long = strings.TrimSuffix(long, "\n")
	// заполняем контейнер данными
	data.Set("url", long)
	// добавляем HTTP-клиент

	client := &http.Client{}
	// If CheckRedirect is nil, the Client uses its default policy, which is to stop after 10 consecutive requests.
	// As a special case, if CheckRedirect returns ErrUseLastResponse,
	// then the most recent response is returned with its body unclosed, along with a nil error.
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	// пишем запрос
	// запрос методом POST должен, помимо заголовков, содержать тело
	// тело должно быть источником потокового чтения io.Reader
	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		panic(err)
	}
	// в заголовках запроса указываем кодировку
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// отправляем запрос и получаем ответ
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	// выводим код ответа
	fmt.Println("Статус-код ", response.Status)
	defer response.Body.Close()
	// читаем поток из тела ответа
	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	// и печатаем его
	fmt.Println(string(body))

	/*
		// GET

		// приглашение в консоли
		fmt.Println("Введите укороченный URL")
		// открываем потоковое чтение из консоли
		reader := bufio.NewReader(os.Stdin)
		// читаем строку из консоли
		long, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		//long = strings.TrimSuffix(long, "\n")

		request, err := http.NewRequest(http.MethodGet, strings.TrimSpace(endpoint+long), nil)
		if err != nil {
			panic(err)
		}

		response, err := client.Do(request)
		if err != nil {
			panic(err)
		}
		defer response.Body.Close()

		// выводим код ответа
		fmt.Println("Статус-код ", response.Status)
		fmt.Println("Location: ", response.Header.Get("Location"))
	*/
}
