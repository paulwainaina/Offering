package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"example.com/controller"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading the .env file")
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(string("mongodb://"+os.Getenv("MONGO_HOST")+":"+os.Getenv("MONGO_PORT"))))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))

	controller.RegisterMemberController(client)
	controller.RegisterFrontController()
	controller.RegisterUserController(client)
	http.ListenAndServe(string(os.Getenv("SERVER")+":"+os.Getenv("SERVER_PORT")), nil)
}
