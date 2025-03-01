package main

import (
	"log"
	"project2/server"
	"project2/agent"
)

func main() {
	go func() {
		log.Println("Запуск сервера...")
		server.StartServer()
	}()

	go func() {
		log.Println("Запуск агента...")
		agent.StartAgent()
	}()

	select {}
}
