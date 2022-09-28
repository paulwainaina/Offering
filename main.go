package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"example.com/controller"
	"example.com/session"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
var sessionManager *session.SessionManager
func main() {	
	sessionManager=session.NewSessionManager()
	go sessionManager.DeleteSession()
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

	controller.RegisterMemberController(client, sessionManager)
	controller.RegisterFrontController(sessionManager)
	controller.RegisterUserController(client, sessionManager)
	http.ListenAndServe(string(os.Getenv("SERVER")+":"+os.Getenv("SERVER_PORT")), nil)
}
