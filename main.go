package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type game struct {
	ID   string `json:"_id" bson:"_id"`
	Name string `json:"name" bson:"name"`
	URL  string `json:"url" bson:"url"`
	Like int    `json:"like" bson:"like"`
}

var games = []game{}

func listGames(context *gin.Context) {
	getGamesFromBase()
	//IntendedJSON pretvara nasu strukturu u format JSON
	context.IndentedJSON(http.StatusOK, games)
}
func addGame(context *gin.Context) {
	var newGame game
	//BindJSON pretvara JSON u format nase strukture
	if err := context.BindJSON(&newGame); err != nil {
		return
	}
	check := addToBase(newGame)
	if check == -1 {
		context.IndentedJSON(http.StatusCreated, gin.H{"message": "URL not valid!"})
	} else if check == -2 {
		context.IndentedJSON(http.StatusCreated, gin.H{"message": "Item with tah index alrady exists!"})
	} else {
		context.IndentedJSON(http.StatusCreated, newGame)
	}
}
func getGame(context *gin.Context) {
	name := context.Param("name")
	getGameFromBase(name)
	if len(games) == 0 {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "Game not found!"})
		return
	}
	context.IndentedJSON(http.StatusOK, games)
}
func deleteGame(context *gin.Context) {
	name := context.Param("name")
	if deleteGameFromBase(name) == 0 {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "Game not found!"})
		return
	}
	context.IndentedJSON(http.StatusOK, gin.H{"message": "Game successfuly deleted."})
}
func likeIncrease(context *gin.Context) {
	name := context.Param("name")
	game, check := updateGameFromBase(name)
	if check == 0 {
		context.IndentedJSON(http.StatusNotFound, gin.H{"message": "Game not found!"})
		return
	}
	context.IndentedJSON(http.StatusOK, game)
}

func addToBase(newGame game) int {
	//checking if the url is valid
	_, err := url.ParseRequestURI(newGame.URL)
	if err != nil {
		return -1
	}
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	context, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err = client.Connect(context)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context)

	database := client.Database("gamesdb")
	gamesCollection := database.Collection("games")

	//Verovatno moze na aelegantniji nacin
	var game1 game
	gamesCollection.FindOne(context, bson.M{"_id": newGame.ID}).Decode(&game1)
	if err != nil {
		log.Fatal(err)
	}
	if game1.ID == newGame.ID {
		return -2
	}
	game, err := gamesCollection.InsertOne(context, bson.D{
		{Key: "_id", Value: newGame.ID},
		{Key: "name", Value: newGame.Name},
		{Key: "url", Value: newGame.URL},
		{Key: "like", Value: newGame.Like},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(game.InsertedID)
	return 0
}
func getGamesFromBase() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	context, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err = client.Connect(context)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context)

	database := client.Database("gamesdb")
	gamesCollection := database.Collection("games")

	cursor, err := gamesCollection.Find(context, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	if err = cursor.All(context, &games); err != nil {
		log.Fatal(err)
	}
	fmt.Println(games)
}
func getGameFromBase(name string) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	context, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err = client.Connect(context)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context)

	database := client.Database("gamesdb")
	gamesCollection := database.Collection("games")

	cursor, err := gamesCollection.Find(context, bson.M{"name": name})
	if err != nil {
		log.Fatal(err)
	}
	if err = cursor.All(context, &games); err != nil {
		log.Fatal(err)
	}
	fmt.Println(games)
}
func deleteGameFromBase(name string) int {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	context, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err = client.Connect(context)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context)

	database := client.Database("gamesdb")
	gamesCollection := database.Collection("games")

	cursor, err := gamesCollection.DeleteOne(context, bson.M{"name": name})
	if err != nil {
		log.Fatal(err)
	}
	return int(cursor.DeletedCount)
}
func updateGameFromBase(name string) (game, int) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	context, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err = client.Connect(context)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context)

	database := client.Database("gamesdb")
	gamesCollection := database.Collection("games")
	if err != nil {
		log.Fatal(err)
	}
	var game game
	if err := gamesCollection.FindOne(context, bson.M{"name": name}).Decode(&game); err != nil {
		log.Fatal(err)
	}
	cursor, err := gamesCollection.UpdateOne(context, bson.M{"name": name}, bson.D{{Key: "$set", Value: bson.D{{Key: "like", Value: game.Like + 1}}}})
	if err != nil {
		log.Fatal(err)
	}
	gamesCollection.FindOne(context, bson.M{"name": name}).Decode(&game)
	return game, int(cursor.ModifiedCount)
}
func main() {
	//pravimo server i pokrecemo ga na localhost:9090
	router := gin.Default()
	//pravimo prvi endpoint
	router.GET("/games", listGames)
	router.GET("/games/:name", getGame)
	router.DELETE("/games/:name", deleteGame)
	router.PATCH("/games/:name", likeIncrease)
	router.POST("/addGame", addGame)
	router.Run("localhost:9090")
}
