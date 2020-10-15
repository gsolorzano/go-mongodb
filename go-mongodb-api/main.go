package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection
var ctx = context.TODO()

// Book struct
type Book struct {
	ID     int    `json:"ID"`
	Title  string `json:"Title"`
	Author string `json:"Author"`
	ISBN   string `json:"ISBN"`
}

func init() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("library").Collection("books")
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the HomePage!")
	fmt.Println("Endpoint Hit: homePage")
}

func createBook(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var book Book
	json.Unmarshal(reqBody, &book)
	_, err := collection.InsertOne(ctx, book)
	if err != nil {
		fmt.Println(err)
	}
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	reqBody, _ := ioutil.ReadAll(r.Body)
	var book Book

	json.Unmarshal(reqBody, &book)
	filter := bson.D{primitive.E{Key: "id", Value: id}}

	// One way
	update := bson.M{"$set": book}

	// Another way
	// update := bson.D{primitive.E{Key: "$set", Value: book}}

	collection.FindOneAndUpdate(ctx, filter, update)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	filter := bson.D{primitive.E{Key: "id", Value: id}}

	res, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		fmt.Println(err)
	}

	if res.DeletedCount == 0 {
		fmt.Println("No tasks were deleted")
	}
}

func returnAllBooks(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnAllBooks")

	filter := bson.D{{}}
	json.NewEncoder(w).Encode(filterBooks(filter))
}

func returnOneBook(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: returnOneBook")

	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	filter := bson.D{
		primitive.E{Key: "id", Value: id},
	}

	json.NewEncoder(w).Encode(filterBooks(filter))
}

func filterBooks(filter interface{}) []*Book {
	// A slice of tasks for storing the decoded documents
	var books []*Book

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		fmt.Println(err)
		return books
	}

	for cur.Next(ctx) {
		var b Book
		err := cur.Decode(&b)
		if err != nil {
			fmt.Println(err)
			return books
		}

		books = append(books, &b)
	}

	if err := cur.Err(); err != nil {
		fmt.Println(err)
		return books
	}

	// once exhausted, close the cursor
	cur.Close(ctx)

	if len(books) == 0 {
		fmt.Println(mongo.ErrNoDocuments)
		return books
	}

	return books
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/books", returnAllBooks)
	myRouter.HandleFunc("/book", createBook).Methods("POST")
	myRouter.HandleFunc("/book/{id}", updateBook).Methods("PUT")
	myRouter.HandleFunc("/book/{id}", deleteBook).Methods("DELETE")
	myRouter.HandleFunc("/book/{id}", returnOneBook)
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func main() {
	handleRequests()
}
