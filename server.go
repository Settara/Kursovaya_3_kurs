package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

// Отслеживаем подключенных клиентов c помощью go карты(ключ-значение:IP-ПОРТ)
var clients = make(map[string]clientInfo)

// Обеспечиваем безопасный доступ к карте клиентов при одновременном доступе или изменении ее несколькими программами.
var mu sync.Mutex

// Структура в которой хранится информация о клиенте
type clientInfo struct {
	//Имя клиента
	username string
	//Сетевое подключение
	conn net.Conn
}

// Эта функция запускается отдельно для каждого подключенного клиента
func handleClient(conn net.Conn) {
	//Соеденение будет закрыто при завершении работы функции
	defer conn.Close()
	//Получает IP и порт в виде строки
	clientAddr := conn.RemoteAddr().String()

	//Запрос пароля
	conn.Write([]byte("Введите пароль от сервера: "))
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		password := scanner.Text()
		if password != ServerPassword {
			conn.Write([]byte("Пароль неверный, соединение разорвано.\n"))
			log.Printf("Пользователь %s не смог ввести правильный пароль.\n", clientAddr)
			return
		}
	}

	//Получение имени
	conn.Write([]byte("Введите ваше имя: "))
	var username string
	if scanner.Scan() {
		username = scanner.Text()
	}

	log.Printf("Пользователь подключен: %s (Имя пользователя: %s)\n", clientAddr, username)
	fmt.Printf("Новый пользователь подключен: %s (Имя пользователя: %s)\n", clientAddr, username)

	//Сохранение информации о клиентах
	mu.Lock()
	clients[clientAddr] = clientInfo{username: username, conn: conn}
	mu.Unlock()

	for scanner.Scan() {
		message := scanner.Text()
		switch {
		case strings.HasPrefix(message, "/msg"):
			handlePrivateMessage(clientAddr, message)
		case message == "/list":
			handleListClients(conn)
		default:
			log.Printf("Неизвестная команда от %s: %s", clientAddr, message)
			conn.Write([]byte("Неизвестная команда. Используйте /help чтобы посмотреть список доступных команд.\n"))
		}
	}

	//Обработка отключения клиента
	mu.Lock()
	delete(clients, clientAddr)
	mu.Unlock()
	log.Printf("Пользователь отключился: %s (Имя пользователя: %s)\n", clientAddr, username)
	fmt.Printf("Пользователь отключился: %s (Имя пользователя: %s)\n", clientAddr, username)
}

// Функция личного сообщения по команде /msg
func handlePrivateMessage(senderAddr, message string) {
	//Разбивает сообщение на каоманду, адрес получателя, и текст сообщения
	parts := strings.SplitN(message, " ", 3)
	if len(parts) < 3 {
		log.Printf("Неверный формат сообщения от %s: %s", senderAddr, message)
		return
	}
	recipientAddr := parts[1]
	msgText := parts[2]
	//Блокирует карту клиентов, чтобы безопасно удалить из нее клиента после его отключения
	mu.Lock()
	recipientInfo, exists := clients[recipientAddr]
	mu.Unlock()

	if exists {
		_, err := recipientInfo.conn.Write([]byte(fmt.Sprintf("Сообщение от %s: %s\n", clients[senderAddr].username, msgText)))
		if err != nil {
			log.Printf("Ошибка отправки сообщения к %s: %v", recipientAddr, err)
		} else {
			log.Printf("Сообщение отправлено от %s к %s: %s", clients[senderAddr].username, recipientInfo.username, msgText)
		}
	} else {
		log.Printf("Получатель %s не найден", recipientAddr)
		senderConn := clients[senderAddr].conn
		senderConn.Write([]byte(fmt.Sprintf("Получатель %s не найден. Проверьте адрес.\n", recipientAddr)))
	}
}

// Вывод списка пользователей, подключенных к серверу
func handleListClients(conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	clientList := "Подключенный пользователи:\n"
	for addr, info := range clients {
		clientList += fmt.Sprintf("Имя пользователя: %s, Адрес: %s\n", info.username, addr)
	}
	conn.Write([]byte(clientList))
}

// Запускает сервер на указанном порту
func startServer(port string) {
	//Запускает tcp-сервер, прослушивающий указанный порт
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
	defer listener.Close()

	log.Printf("Сервер запущен по порту %s\n", port)
	fmt.Printf("Сервер запущен по порту %s. Ожидаем подключения пользователей...\n", port)

	for {
		//Ожидает входящих клиентских подключений
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Ошибка при приеме соединения: %v", err)
			continue
		}
		//Вызываем функцию при каждом новом подключении
		go handleClient(conn)
	}
}

func main() {
	//Сервер записывает события в файл с именем server.log.
	logFile, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	//Стандартный порт
	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	startServer(port)
}
