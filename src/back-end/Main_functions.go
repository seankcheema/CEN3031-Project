package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dimuska139/rawg-sdk-go"
	//"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User Struct
type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex"`
	Email    string `gorm:"uniqueIndex"`
	Password string
}

// Review Struct
type Review struct {
	gorm.Model
	GameName    string  //Names of game being reviewed
	Rating      float32 //Rating (out of 5) of the game
	Description string  //Description of the game played
	Username    string  //Name of the account
	PlayStatus  string  //PLAYING, DROPPED, COMPLETED, ON HOLD
}

type GameRanking struct {
	gorm.Model
	GameName      string  `gorm:"uniqueIndex"` // Name of game
	AverageRating float32 // Average Rating (out of 5) of the game
	NumReviews    int     // Number of times a game has been reviewed
}

// Enable the front end to access backend, enables Cross-Origin Resource Sharing because frontend and backend serve from different domains
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

// Placeholder that handles base hanlder "/"
func Hello(w http.ResponseWriter, r *http.Request) {
	//Allows the domain to be accessed by frontend
	enableCors(&w)

	fmt.Fprint(w, "Hello, Welcome to the Temporary Back-End Home Page")
}

// Handles creation of user struct and stores in the database {W-I-P}
func SignUp(w http.ResponseWriter, r *http.Request) *User {
	//Allows the domain to be accessed by frontend
	enableCors(&w)

	//Open the database
	db, err := gorm.Open(sqlite.Open("currentUsers.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	//Migrate the format of the user struct to Gorm's database
	db.AutoMigrate(&User{})

	//Create a user
	var user User

	//Recieve username and password from front, using the parameters listed in the passed in json file
	pass := r.ContentLength
	if pass > 0 {
		json.NewDecoder(r.Body).Decode(&user)
	} else { // For unit testing
		user.Username = "UnitTest2"
		user.Email = "UnitTest2@gmail.com"
		user.Password = "PASSWORD"
	}

	//Check that the username doesn't already exist in the database
	hasUser := db.Where("username = ?", user.Username).First(&user).Error
	hasEmail := db.Where("email = ?", user.Email).First(&user).Error

	if hasUser == nil { //If the user already exists, return an error
		fmt.Println("User ", user.Username, " already exists!")
		w.WriteHeader(http.StatusInternalServerError) //Error is sent
		return nil
	} else if hasEmail == nil { //If the email already exists, return an error
		fmt.Println("Email ", user.Email, " already exists!")
		w.WriteHeader(http.StatusInternalServerError) //Error is sent
		return nil
	} else { //If its a new user, add the user and its associated information to the database
		db.Create(&User{Username: user.Username, Email: user.Email, Password: user.Password})
		w.WriteHeader(http.StatusCreated)
		return &user
	}
}

// Signs in the user, and tell the front end that the user
func SignIn(w http.ResponseWriter, r *http.Request, currentlyActiveUser *string) *User {
	//Allows the domain to be accessed by frontend
	enableCors(&w)

	//Open the database
	db, err := gorm.Open(sqlite.Open("currentUsers.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	//Migrate the format of the user struct to Gorm's database
	db.AutoMigrate(&User{})

	//Create a user
	var user User

	//Recieve username and password from front, using the parameters listed in the passed in json file
	pass := r.ContentLength
	if pass > 0 {
		json.NewDecoder(r.Body).Decode(&user)
	} else { // For unit testing
		user.Username = "UnitTestSprint5"
		user.Password = "PASSWORD"
	}

	//Check that the username doesn't already exist in the database
	var currUser User
	hasUser := db.Where("username = ?", user.Username).First(&currUser).Error

	if hasUser != nil { //If the user doesn't exist, return error
		fmt.Println("User ", user.Username, " doesn't exist!")
		w.WriteHeader(http.StatusInternalServerError) //IDK What this status does
		return nil
	} else { //If its a new user, add the user and the information to the database
		if currUser.Password != user.Password {
			fmt.Println("User ", user.Username, " doesn't exist!")
			w.WriteHeader(http.StatusInternalServerError) //IDK What this status does
			return nil
		}
		currentlyActiveUser = &currUser.Username
		w.WriteHeader(http.StatusOK)
		return &currUser
	}
}

// Create an object of a user made review using info passed in by the front end and store it in the databse
func WriteAReview(w http.ResponseWriter, r *http.Request, currentlyActiveUser *string) *Review {
	//Allows the domain to be accessed by frontend
	enableCors(&w)

	//Open the database
	db, err := gorm.Open(sqlite.Open("Reviews.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	//Migrate the format of the user struct to Gorm's database
	db.AutoMigrate(&Review{})

	var review Review // new review

	pass := r.ContentLength
	if pass > 0 {
		json.NewDecoder(r.Body).Decode(&review)
	} else { // For unit testing
		review.GameName = "Final Fantasy 3"
		review.Rating = 5.0
		review.Description = "Game guud"
		review.Username = "UnitTest"
		review.PlayStatus = "PLAYING"
	}

	var old Review
	hasReview := db.Where("username = ?", review.Username).Where("game_name = ?", review.GameName).First(&old).Error
	if hasReview == nil { // if review already exists, overwrite it
		UserGameRankings(&old, false)
		old.Description = review.Description
		old.Rating = review.Rating
		old.PlayStatus = review.PlayStatus
		db.Save(&old)
		UserGameRankings(&review, true)
		w.WriteHeader(http.StatusOK)
		return &review
	} else { // else create new review
		db.Create(&Review{GameName: review.GameName, Rating: review.Rating, Description: review.Description, Username: review.Username, PlayStatus: review.PlayStatus})
		UserGameRankings(&review, true)
		w.WriteHeader(http.StatusCreated)
		return &review
	}
}

// Helper function for WriteAReview
func UserGameRankings(review *Review, add bool) {
	// Open rankings db
	db, err := gorm.Open(sqlite.Open("UserGameRankings.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	//Migrate the GameRanking struct
	db.AutoMigrate(&GameRanking{})

	// Find the first (should be only) instance of that game
	var temp GameRanking
	hasGame := db.Where("game_name = ?", review.GameName).First(&temp).Error
	if hasGame == nil && add { // game already exists and we're adding
		num := temp.AverageRating * float32(temp.NumReviews)
		num += review.Rating
		temp.NumReviews++
		temp.AverageRating = (num / float32(temp.NumReviews))
		db.Save(&temp)
	} else if hasGame == nil && !add { // game already exists and we're subtracting
		num := temp.AverageRating * float32(temp.NumReviews)
		num -= review.Rating
		temp.NumReviews--
		temp.AverageRating = (num / float32(temp.NumReviews))
		db.Save(&temp)
	} else { // if game doesn't exist, add it
		db.Create(&GameRanking{GameName: review.GameName, AverageRating: review.Rating, NumReviews: 1})
	}
}

// Returns to front end a JSON of all of a specified user's game reviews
func GetReviews(w http.ResponseWriter, r *http.Request, currentlyActiveUser *string) []*Review {
	//Allows the domain to be accessed by frontend
	enableCors(&w)

	//Open the database
	db, err := gorm.Open(sqlite.Open("Reviews.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	//Migrate the format of the user struct to Gorm's database to get the username
	db.AutoMigrate(&User{})

	//Create a user and search the REVIEW database
	var user User
	pass := r.ContentLength
	if pass > 0 {
		json.NewDecoder(r.Body).Decode(&user)
	} else { // For unit testing
		user.Username = "UnitTest"
		user.Password = "PASSWORD"
	}
	var reviews []*Review
	db.Where("username = ?", user.Username).Find(&reviews)

	if len(reviews) == 0 { // if there are no reviews, write an error to the header
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	} else { // else send reviews to front-end
		response, err := json.Marshal(reviews)
		if err != nil {
			return nil
		}
		w.WriteHeader(http.StatusOK)
		w.Write(response)

		if err != nil {
			return nil
		}

		return reviews
	}
}

// Takes the handler, get the game requested, and returns json
func Game(w http.ResponseWriter, r *http.Request, client *rawg.Client) []*rawg.Game {
	//Allows the doamin to be accessed by frontend
	enableCors(&w)

	//Specify status code
	w.WriteHeader(http.StatusOK)

	//Pull slug from query params
	slug := r.URL.Query().Get("slug")

	//--------------------

	//Update response writer
	filter := rawg.NewGamesFilter().SetPageSize(10).SetSearch(slug)
	var games []*rawg.Game
	var num int
	var err error
	games, num, err = client.GetGames(filter)

	response, err := json.Marshal(games)
	if err != nil {
		return nil
	}

	w.Write(response)
	if err != nil {
		return nil
	}

	_ = err
	_ = num
	_ = games

	return games
}

// Takes the handler's page, and returns all games of that page (40 max)
func Games(w http.ResponseWriter, r *http.Request, client *rawg.Client) []*rawg.Game {
	//Allows the doamin to be accessed by frontend
	enableCors(&w)

	//Specify status code
	w.WriteHeader(http.StatusOK)

	//Page iterator
	tempCurrPage := r.URL.Query().Get("page")
	tempPageSize := r.URL.Query().Get("pageSize")

	var err error
	//cast to int
	currPage, err := strconv.Atoi(tempCurrPage)
	if err != nil {
		currPage = 1
	}
	pageSize, err := strconv.Atoi(tempPageSize)
	if err != nil {
		pageSize = 40
	}

	//Update response writer and request all games
	filter := rawg.NewGamesFilter().SetPage(currPage).SetPageSize(pageSize)
	var games []*rawg.Game
	var num int

	games, num, err = client.GetGames(filter)

	//Limit of 40 games per "page"
	response, err := json.Marshal(games)
	if err != nil {
		return nil
	}

	w.Write(response)
	if err != nil {
		return nil
	}

	_ = err
	_ = num
	_ = games

	return games
}

// Handles requests to get the 4 most recent games released
// Due to the constantly changing nature of the games that are updated per day, this cannot be concretely unit tested with set values
func RecentGames(w http.ResponseWriter, r *http.Request, client *rawg.Client) {
	//Allows the doamin to be accessed by frontend
	enableCors(&w)

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

	//Specify status code
	w.WriteHeader(http.StatusOK)
	w.Write(response)

	_ = err
	_ = num
	_ = games
}

// Returns up to 5? top games to be displayed on the homepage
func TopGames(w http.ResponseWriter, r *http.Request, client *rawg.Client) []GameRanking {
	enableCors(&w)

	// Open GameRankings db
	db, err := gorm.Open(sqlite.Open("UserGameRankings.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	//get all game rankings and store in review
	var reviews []GameRanking
	db.Where("average_rating >= ?", float32(0)).Find(&reviews)

	// Collect the names of the 5 highest rated games

	reviews = QuickSortDesc(reviews)

	// Turn these game names into rawg.Game objects
	var topGames []*rawg.Game
	for i := 0; i < len(reviews); i++ {
		//fmt.Print(reviews[i].GameName)
		filter := rawg.NewGamesFilter().SetPageSize(1).SetSearch(reviews[i].GameName)
		temp, _, _ := client.GetGames(filter)

		topGames = append(topGames, temp[0])
	}

	if len(topGames) == 0 { // if 0 games, write an error to the header
		w.WriteHeader(http.StatusInternalServerError)
	} else { // else send games to front-end
		response, _ := json.Marshal(topGames)
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	}

	return reviews
}

// Quick-sort algoritm
func QuickSortDesc(arr []GameRanking) []GameRanking {
	if len(arr) <= 1 {
		return arr
	}

	pivot := arr[0]
	var left []GameRanking
	var right []GameRanking

	for _, num := range arr[1:] {
		if num.AverageRating >= pivot.AverageRating {
			left = append(left, num)
		} else {
			right = append(right, num)
		}
	}

	left = QuickSortDesc(left)
	right = QuickSortDesc(right)

	return append(append(left, pivot), right...)
}

// Returns upcoming games that haven't been released yet
func UpcomingGames(w http.ResponseWriter, r *http.Request, client *rawg.Client) {
	//Allows the doamin to be accessed by frontend
	enableCors(&w)

	//Create time frame
	now := time.Now()
	tomorrow := time.Now().AddDate(0, 0, -1)
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	start := tomorrow
	end := lastOfMonth //1 month from now

	var specifiedTime rawg.DateRange
	specifiedTime.From = start
	specifiedTime.To = end

	//Set filer to search all games in the next, ordered by release date {handled by RAWG itself}
	filter := rawg.NewGamesFilter().SetPageSize(40).SetDates(&specifiedTime).SetOrdering("rating_top")
	var games []*rawg.Game
	var num int
	var err error

	games, num, err = client.GetGames(filter)

	response, err := json.Marshal(games)
	if err != nil {
		return
	}

	//Specify status code
	w.WriteHeader(http.StatusOK)
	w.Write(response)

	_ = err
	_ = num
	_ = games
}

// Returns the usernames most similar to the username passed in through the URL
func GetUsers(w http.ResponseWriter, r *http.Request) []User {
	//Allows the doamin to be accessed by frontend
	enableCors(&w)

	//Open database
	db, err := gorm.Open(sqlite.Open("currentUsers.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	//Determine if its a unit test and get the string we will be working with
	var username string
	pass := r.ContentLength
	if pass > 0 {
		username = r.URL.Query().Get("user")
	} else { // For unit testing
		username = "UnitTest"
	}

	//Migrate the given json file and fomrat it in terms of the User struct so we may work with it
	db.AutoMigrate(&User{})

	//Create a user array to return all user sthat somewhat match the name given to us by the front end
	var users []User

	//Search our databse for all the possible users that are "LIKE" the given string
	hasUsers := db.Where("username LIKE ?", username).Find(&users).Error

	//If we the found material is some error, then it doesn't exist and we write an Internal Server error to the writer
	//Otherwise, we return the marshalled user array and return successful
	if hasUsers != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response, _ := json.Marshal(users)
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	}

	return users
}

// Returns an array of reviews from the last month arranged from most recent to least recent
func RecentReviews(w http.ResponseWriter, r *http.Request) []Review {
	//Allows the domain to be accessed by front-end
	enableCors(&w)

	//Open database
	db, err := gorm.Open(sqlite.Open("Reviews.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	//Create time frame
	start := time.Now()
	end := start.AddDate(0, -1, 0) //1 month ago from current time

	db.AutoMigrate(&Review{})

	var latestReviews []Review

	// Get the reveiews from the last month
	recentReviews := db.Where("updated_at > ?", end).Find(&latestReviews).Error

	if recentReviews != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		latestReviews = reverseArray(latestReviews)
		response, _ := json.Marshal(latestReviews)
		w.WriteHeader(http.StatusOK)
		w.Write(response)
		return latestReviews
	}
	return nil
}

// reverse order of review array
func reverseArray(arr []Review) []Review {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

// Returns the game with the most number of reviews in the last month
func GetFeaturedGame(w http.ResponseWriter, r *http.Request, client *rawg.Client) *rawg.Game {
	//Allows the doamin to be accessed by frontend
	enableCors(&w)

	//Open database
	db, err := gorm.Open(sqlite.Open("UserGameRankings.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	//Create time frame
	start := time.Now()
	end := start.AddDate(0, -1, 0) //1 month ago from current time

	db.AutoMigrate(&GameRanking{})

	var gameRankings []GameRanking

	// Get all games updated in the last month
	recentRankings := db.Where("updated_at > ?", end).Find(&gameRankings).Error

	if recentRankings != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		var featuredGame string // the game with the highest ranking from []gameRankings
		max := gameRankings[0].NumReviews
		for i := 0; i < len(gameRankings); i++ {
			if gameRankings[i].NumReviews > max {
				featuredGame = gameRankings[i].GameName
				max = gameRankings[i].NumReviews
			}
		}
		// get rawg.Game with name featuredGame
		filter := rawg.NewGamesFilter().SetPageSize(1).SetSearch(featuredGame)

		games, _, _ := client.GetGames(filter)

		game := games[0]
		response, _ := json.Marshal(game)
		w.WriteHeader(http.StatusOK)
		w.Write(response)
		return game
	}
	return nil
}
