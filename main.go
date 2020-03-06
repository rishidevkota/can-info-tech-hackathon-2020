package main

import (
	"log"
	"net/http"
	"text/template"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB

//User type=0 for guest and type=1 for host
type User struct {
	gorm.Model
	Name        string
	Email       string
	Type        int
	Reservation []Reservation
	Experiences []Experience
}

//Experience ...
type Experience struct {
	gorm.Model
	UserID       uint
	Title        string
	Location     string
	Duration     string
	Description  string
	Type         string
	User         User
	Reservations []Reservation
}

//Reservation ...
type Reservation struct {
	gorm.Model
	UserID       uint
	ExperienceID uint
	ArrivalDate  string
}

var indexTmpl = template.Must(template.ParseFiles("base.html", "index.html"))
var loginTmpl = template.Must(template.ParseFiles("base.html", "login.html"))
var dashboardTmpl = template.Must(template.ParseFiles("base.html", "dashboard.html"))
var registerTmpl = template.Must(template.ParseFiles("base.html", "register.html"))

func auth(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	var experiences []Experience
	q := r.FormValue("q")
	if q != "" {
		db.Where("location LIKE ?", "%"+q+"%").Preload("User").Find(&experiences)
	} else {
		db.Preload("User").Find(&experiences)
	}

	indexTmpl.Execute(w, &experiences)
}

func login(w http.ResponseWriter, r *http.Request) {
	loginTmpl.Execute(w, nil)
}

func register(w http.ResponseWriter, r *http.Request) {
	registerTmpl.Execute(w, nil)
}

func dasboard(w http.ResponseWriter, r *http.Request) {
	dashboardTmpl.Execute(w, nil)
}

func main() {
	var err error
	db, err = gorm.Open("sqlite3", "data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.AutoMigrate(&User{}, &Experience{}, &Reservation{})

	// user := &User{
	// 	Name:  "hari",
	// 	Email: "h@g.com",
	// 	Type:  1,
	// 	Experiences: []Experience{
	// 		{
	// 			Title:       "travel palpa with local food",
	// 			Location:    "palpa",
	// 			Description: "asd daf a asdfl lasdf",
	// 		},
	// 	},
	// }

	//db.Create(&user)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))
	http.HandleFunc("/", index)
	http.HandleFunc("/login", login)
	http.HandleFunc("/register", register)
	http.ListenAndServe(":8080", nil)
}
