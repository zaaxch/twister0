package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/tkanos/gonfig"
	"github.com/tsuna/gohbase"
)

// Configuration of application
type Configuration struct {
	AccessToken       string
	AccessTokenSecret string
	ConsumerKey       string
	ConsumerSecret    string
}

// Response of application
type Response struct {
	Data   interface{} `json:"data"`
	Error  interface{} `json:"error"`
	Status int         `json:"status"`
}

var trends []string
var store = sessions.NewCookieStore([]byte("something-very-secret"))
var configuration = Configuration{}
var client = gohbase.NewClient("localhost")

func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}

func fetchTrends() (trends []string) {
	api := anaconda.NewTwitterApi(configuration.AccessToken, configuration.AccessTokenSecret)
	result, _ := api.GetTrendsByPlace(23424977, nil)
	for _, trend := range result.Trends {
		trends = append(trends, trend.Name)
	}
	return
}

func like() {
	trends = fetchTrends()
	go doEvery(1*time.Hour, func(t time.Time) {
		trends = fetchTrends()
	})
	go doEvery(10*time.Second, func(arg2 time.Time) {
		api := anaconda.NewTwitterApi(configuration.AccessToken, configuration.AccessTokenSecret)
		searchResult, error := api.GetSearch(string(trends[rand.Intn(len(trends))]), nil)
		if error != nil {
			fmt.Println("Error fetching search results.")
			return
		}
		status := searchResult.Statuses[rand.Intn(len(searchResult.Statuses))]
		fmt.Printf("Attempting to like status with ID: %s.\n", status.IdStr)
		tweet, error := api.Favorite(status.Id)
		if error != nil {
			fmt.Printf("Error liking status with ID: %s.\n", status.IdStr)
			return
		}
		fmt.Printf("Successfully liked status with ID: %s.\n", tweet.IdStr)
	})
}

func respond(writer http.ResponseWriter, data interface{}, error interface{}, status int) {
	response := Response{Data: data, Error: error, Status: status}
	j, _ := json.Marshal(response)
	fmt.Fprint(writer, string(j))
}

func main() {
	_ = gonfig.GetConf("conf.json", &configuration)
	gob.Register(&oauth.Credentials{})
	anaconda.SetConsumerKey(configuration.ConsumerKey)
	anaconda.SetConsumerSecret(configuration.ConsumerSecret)
	r := mux.NewRouter()
	r.HandleFunc("/oauth/callback", OAuthCallback)
	r.HandleFunc("/oauth/init", OAuthInit)
	r.HandleFunc("/oauth/self", OAuthSelf)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, error := ioutil.ReadFile("index.html")
		if error != nil {
			respond(w, nil, error.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(body))
	})
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.Handle("/", r)
	http.ListenAndServe(":8081", r)
}
