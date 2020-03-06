package main

import (
	"log"
	"net/http"
	"text/template"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

//User ...
type User struct {
	gorm.Model
	Name        string
	Email       string
	Type        int
	Reservation []Reservation
	Experience  []Experience
}

//Experience ...
type Experience struct {
	gorm.Model
	UserID      uint
	Title       string
	Location    string
	Duration    string
	Description string
	Type        string
}

//Reservation ...
type Reservation struct {
	gorm.Model
	UserID      uint
	ArrivalDate string
}

var indexTmpl = template.Must(template.ParseFiles("base.html", "index.html"))
var loginTmpl = template.Must(template.ParseFiles("base.html", "login.html"))

func index(w http.ResponseWriter, r *http.Request) {
	indexTmpl.Execute(w, nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	loginTmpl.Execute(w, nil)
}

func main() {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.AutoMigrate(&User{}, &Experience{}, &Reservation{})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))
	http.HandleFunc("/", index)
	http.HandleFunc("/login", login)
	http.ListenAndServe(":8080", nil)
}
