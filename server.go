package main

import (
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/gorilla/mux"
	"github.com/jasonlvhit/gocron"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	dbops "sky-meter/dbops"
	httpreponser "sky-meter/httpres"
	models "sky-meter/models"
)

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to sky-meter")
}

func getStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	httpresdata, _ := httpreponser.GetHttpdata("https://bing.com")
	w.Write(httpresdata)
	return
}

func httpSyntheticCheck(endpoint string, time uint64) {
	gocron.Every(time).Second().Do(callEndpoint, endpoint)
	<-gocron.Start()
}

func callEndpoint(endpoint string) {
	httpresdata, _ := httpreponser.GetHttpdata(endpoint)
	log.Println(string(httpresdata))
}

func main() {
	senenv := os.Getenv("sentry_dsn")
	senterr := sentry.Init(sentry.ClientOptions{
		Dsn: senenv,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
	})
	if senterr != nil {
		log.Fatalf("sentry.Init: %s", senterr)
	}

	jsonFile, err := os.Open("input.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var endpoints models.JsonInput

	json.Unmarshal(byteValue, &endpoints)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  "host=localhost user=postgres password=postgres dbname=postgres port=5433 sslmode=disable TimeZone=Asia/Shanghai",
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		log.Println(err)
	}

	dbops.InitialMigration(db)
	dbops.InsertUrlsToDb(db, endpoints)

	log.Println("listening on port 8080")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeLink)
	router.HandleFunc("/stats", getStats).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))

}
