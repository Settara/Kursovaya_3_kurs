package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

// Функция выводит сообщение о том, какие команды можно ввести.
func printHelp() {
	helpMessage := `
Команды:
- /msg : Отправить сообщение другому пользователю.
- /list : Посмотреть список пользователей, подключенных к серверу.
- /help : Посмотреть список доступных команд.
- /exit : Выйти из мессенджера.
`
	fmt.Println(helpMessage)
}

// Основная функция которая управляет клиентом
func startClient(serverAddr string) {

	//Попытка установить tcp-соединение с сервером по переданному адресу
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Не получилось подключиться к серверу: %v", err)
	}

	//Соеденение будет закрыто при завершении работы функции
	defer conn.Close()

	//Горутина для получения сообщений от сервера
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			//Выводит каждое сообщение, получаемое от сервера.
			fmt.Printf("\n%s\n", scanner.Text())
		}
	}()

	//Для считывания пользовательского ввода с консоли
	scanner := bufio.NewScanner(os.Stdin)

	//Спрашиваем знает ли пользователь пароль от сервера
	fmt.Print("Введите пароль от сервера: ")
	if scanner.Scan() {
		password := scanner.Text()
		//Отправляем введенный пароль на сервер
		_, err := conn.Write([]byte(password + "\n"))
		if err != nil {
			log.Fatalf("Пароль не правильный: %v", err)
		}
	}

	//Спрашиваем у пользователя его имя
	fmt.Print("Введите ваше имя: ")
	var username string
	if scanner.Scan() {
		username = scanner.Text()
		//Отправляем имя на сервер
		_, err := conn.Write([]byte(username + "\n"))
		if err != nil {
			log.Fatalf("Что-то с именем пошло не так: %v", err)
		}
	}

	fmt.Println("Вы подключены к серверу под именем: ", username)
	fmt.Println("Введите '/help' для получения списка доступных команд.")

	//Бесконечный цикл который крутится, пока пользователь не завершит работу
	for {
		fmt.Print("Введите сообщение или команду (/help для списка команд, /exit чтобы выйти): ")
		if scanner.Scan() {
			command := scanner.Text()

			switch {
			case command == "/exit":
				fmt.Println("Вы отключились от сервера.")
				return
			case command == "/help":
				printHelp()
			case command == "/list":
				//Отправляет команду /list на сервер
				_, err := conn.Write([]byte("/list\n"))
				if err != nil {
					log.Printf("Ошибка в отправке команды /list : %v", err)
				}
			case strings.HasPrefix(command, "/msg"):
				//Спрашиваем куда отправить сообщине
				fmt.Print("Введите IP:Port получателя: ")
				if scanner.Scan() {
					recipient := scanner.Text()
					//Проверяем что сообщение не пустое
					message := strings.TrimSpace(command[4:])
					if message == "" {
						fmt.Println("Сообщение не может быть пустым.")
					} else {
						fullMessage := fmt.Sprintf("/msg %s %s\n", recipient, message)
						//Отправляем сообщение на сервер
						_, err := conn.Write([]byte(fullMessage))
						if err != nil {
							log.Printf("Ошибка отправки сообщения: %v", err)
						} else {
							fmt.Println("Сообщение отправлено.")
						}
					}
				}
			default:
				//Для обычных сообщений(не команд)
				fmt.Print("Введите IP:Port получателя: ")
				if scanner.Scan() {
					recipient := scanner.Text()

					//Проверяем что сообщение не пустое
					if strings.TrimSpace(command) == "" {
						fmt.Println("Сообщение не может быть пустым.")
					} else {
						fullMessage := fmt.Sprintf("/msg %s %s\n", recipient, command)
						//Отправляем сообщение на сервер
						_, err := conn.Write([]byte(fullMessage))
						if err != nil {
							log.Printf("Ошибка отправки сообщения: %v", err)
						} else {
							fmt.Println("Сообщение отправлено.")
						}
					}
				}
			}
		}
	}
}

func main() {
	//Проверяем указал ли пользователь адрес сервера для запуска программы
	if len(os.Args) < 2 {
		fmt.Println("Введите: go run client.go <адрес сервера:порт>")
		return
	}
	//Записывает в serverAddr адрес сервера и вызывает основную функцию
	serverAddr := os.Args[1]
	startClient(serverAddr)
}
