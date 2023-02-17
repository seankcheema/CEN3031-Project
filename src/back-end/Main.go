package main

import (
	"fmt"
	"net/http"
	"time"

	"encoding/json"

	"github.com/dimuska139/rawg-sdk-go"
	"github.com/gorilla/mux"
	// _ "github.com/jinzhu/gorm/dialects/sqlite"
	// _ "gorm.io/gorm"
)

func main() {

	//Creates a rounter
	router := mux.NewRouter()
	//Create RAWG SDK config and client
	config := rawg.Config{
		ApiKey:   "476cd66f8e4d44eb975aad199e0d7a07", //RAWG API key
		Language: "en",                               // English
		Rps:      5,                                  // Has to stay 5 (limit)
	}

	//Setup client to talk to database
	var client *rawg.Client = rawg.NewClient(http.DefaultClient, &config)
	users := make(map[string]*user)

	//Functions that handles the url's sent from the backend:
	router.HandleFunc("/specific-game", func(w http.ResponseWriter, r *http.Request) {
		PrintGames(w, r, client)
	}).Methods("GET")

	router.HandleFunc("/allGames", func(w http.ResponseWriter, r *http.Request) {
		PrintAllGames(w, r, client)
	}).Methods("GET")

	router.HandleFunc("/sign-up", func(w http.ResponseWriter, r *http.Request) {
		SignUp(w, r, users)
	}).Methods("GET")

	router.HandleFunc("/", Hello).Methods("GET")
	http.Handle("/", router)

	router.HandleFunc("/recent", func(w http.ResponseWriter, r *http.Request) {
		recentGames(w, r, client)
	}).Methods("GET")

	//Start and listen for requests
	http.ListenAndServe(":8080", router)

}

// Enable the front end to access backend, enables Cross-Origin Resource Sharing because frontend and backend serve from different domains
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func SignUp(w http.ResponseWriter, r *http.Request, users map[string]*user) {
	//Allows the doamin to be accessed by frontenf
	enableCors(&w)

	//Updates the header to indicate successful reach of the fuction
	w.WriteHeader(http.StatusOK)

	//User map Creation
	var username string
	var password string

	fmt.Println("Input Username:")
	fmt.Scanln(&username)
	fmt.Println("Input Password:")
	fmt.Scanln(&password)
	if _, ok := users[username]; ok {
		fmt.Fprint(w, "User ", username, " already exists!")
	} else {
		users[username] = newUser(username, password)
		fmt.Fprint(w, "User ", username, " added!")
	}

}

func Hello(w http.ResponseWriter, r *http.Request) {
	//Allows the doamin to be accessed by frontenf
	enableCors(&w)

	fmt.Fprint(w, "Hello, Welcome to the Temporary Back-End Home Page")
}

func PrintGames(w http.ResponseWriter, r *http.Request, client *rawg.Client) {
	//Allows the doamin to be accessed by frontenf
	enableCors(&w)

	//Specify status code
	w.WriteHeader(http.StatusOK)

	fmt.Println("Input game name:")

	var err error

	var name string

	fmt.Scanln(&name)

	//Update response writer
	filter := rawg.NewGamesFilter().SetPageSize(40).SetSearch(name)
	var games []*rawg.Game
	var num int
	games, num, err = client.GetGames(filter)
	for i := 0; i < 10; i++ {
		fmt.Fprint(w, "Name: ")
		fmt.Fprintln(w, games[i].Name)
		fmt.Fprint(w, "Rating: ")
		fmt.Fprintln(w, games[i].Rating)
	}

	_ = err
	_ = num
	_ = games
}

func PrintAllGames(w http.ResponseWriter, r *http.Request, client *rawg.Client) {
	//Allows the doamin to be accessed by frontenf
	enableCors(&w)

	//Specify status code
	w.WriteHeader(http.StatusOK)

	//Update response writer and request all games
	filter := rawg.NewGamesFilter().SetPageSize(40)
	var games []*rawg.Game
	var num int
	var err error

	games, num, err = client.GetGames(filter)

	//Limit of 40 games per "page" so we iterarte through all pages

	// for i := 0; i < 40; i++ {

	// fmt.Fprintln(w, games[0])
	response, err := json.Marshal(games)
	if err != nil {
		return
	}
	//}
	w.Write(response)
	// response, err = json.Marshal(games[1])
	if err != nil {
		return
	}
	//}
	// w.Write(response)
	_ = err
	_ = num
	_ = games
}

// Handles requests to get the 4 most recent games released
func recentGames(w http.ResponseWriter, r *http.Request, client *rawg.Client) {
	//Allows the doamin to be accessed by frontenf
	enableCors(&w)

	//Specify status code
	w.WriteHeader(http.StatusOK)

	//Create time frame
	start := time.Now()
	end := start.AddDate(0, -1, 0) //1 month ago from current time

	var specifiedTime rawg.DateRange
	specifiedTime.From = end
	specifiedTime.To = start

	//Set filer to search all games in the past month, ordered by release date {handled by RAWG itself}
	filter := rawg.NewGamesFilter().SetPageSize(4).SetOrdering("released")
	var games []*rawg.Game
	var num int
	var err error

	games, num, err = client.GetGames(filter)

	response, err := json.Marshal(games)
	if err != nil {
		return
	}

	w.Write(response)

	_ = err
	_ = num
	_ = games
}

type user struct {
	username string
	password string
}

func newUser(username string, password string) *user {
	u := user{username: username, password: password}
	return &u
}
