package main

import (
	tgClient "bot/LBot/src/internal/clients/telegram"
	eventconsumer "bot/LBot/src/internal/consumer/event-consumer"
	"bot/LBot/src/internal/events/telegram"
	"bot/LBot/src/internal/storage/files"
	"flag"
	"log"
)

const(
	tgBotHost = "api.telegram.org"
	storagePath = "storage"
	batchSize = 100
)

func main() {
	strg := files.New(storagePath)
	eventsProcessor := telegram.New(tgClient.New(tgBotHost, mustToken()), &strg)

	log.Print("service started")

	evcons := eventconsumer.New(eventsProcessor, eventsProcessor, batchSize)
	if err := evcons.Start(); err != nil {

	}
}

func mustToken() string {
	token := flag.String("token", "", "telegram-token to access bot")
	flag.Parse()

	if *token == "" {
		log.Fatal("invalid token")
	}
	return *token
}