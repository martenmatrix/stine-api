package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"stine-push-notifications/pkg/stine"
)

func main() {
	fmt.Println("Starting")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Unable to load .env file")
	}

	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	err = stine.GetSession(username, password)
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to create STiNE session: %s", err))
	}
}
