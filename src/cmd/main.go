package main

import (
	"log"
	"os"
	"tickTackToe/internal/di"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil && !os.IsNotExist(err) {
		log.Fatal("Error loading .env file")
	}
	di.Build()
}
