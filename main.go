package main

import (
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/gorilla/context"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB

//User type=0 for guest and type=1 for host
type User struct {
	gorm.Model
	Name        string
	Email       string
	Password    string
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

type key int

const MyKey key = 0

var indexTmpl = template.Must(template.ParseFiles("base.html", "index.html"))
var loginTmpl = template.Must(template.ParseFiles("base.html", "login.html"))
var dashboardTmpl = template.Must(template.ParseFiles("base.html", "dashboard.html"))
var registerTmpl = template.Must(template.ParseFiles("base.html", "register.html"))
var exTmpl = template.Must(template.ParseFiles("base.html", "ex.html"))

func auth(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user User
		cookie, err := r.Cookie("auth")
		if err != nil {
			http.Redirect(w, r, "/login", 302)
			return
		}
		as := strings.Split(cookie.Value, "&")
		db.Where("email=?", as[0]).First(&user)
		if user.Password != as[1] {
			http.Redirect(w, r, "/login", 302)
			return
		}

		context.Set(r, MyKey, user)

		fn(w, r)
	}
}

func withuser(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user User
		cookie, err := r.Cookie("auth")
		if err == nil {
			as := strings.Split(cookie.Value, "&")
			db.Where("email=? and password=?", as[0], as[1]).First(&user)
		}

		context.Set(r, MyKey, user)
		fn(w, r)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	var email string
	cookie, err := r.Cookie("auth")
	if err != nil {
		email = ""
	} else {
		as := strings.Split(cookie.Value, "&")
		email = as[0]
	}

	var experiences []Experience
	q := r.FormValue("q")
	if q != "" {
		db.Where("location LIKE ?", "%"+q+"%").Preload("User").Find(&experiences)
	} else {
		db.Preload("User").Find(&experiences)
	}

	indexTmpl.Execute(w, map[string]interface{}{
		"experiences": &experiences,
		"email":       email,
		"user":        context.Get(r, MyKey),
	})
}

func ex(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	var exp Experience
	db.First(&exp, id)

	exTmpl.Execute(w, map[string]interface{}{
		"experience": &exp,
		"user":       context.Get(r, MyKey),
	})
}

func login(w http.ResponseWriter, r *http.Request) {
	var messsage string
	if r.Method == "POST" {
		var user User
		db.Where("email=? and password=?", r.FormValue("email"), r.FormValue("password")).First(&user)
		if user.Email == "" {
			messsage = "Email or password is incorrect"
		} else {
			http.SetCookie(w, &http.Cookie{
				Name:   "auth",
				Value:  user.Email + "&" + user.Password,
				Path:   "/",
				MaxAge: 2592000,
			})
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}

	loginTmpl.Execute(w, map[string]interface{}{
		"message": messsage,
		"user":    context.Get(r, MyKey),
	})
}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		var usertype int

		if r.FormValue("role") == "guest" {
			usertype = 0
		} else {
			usertype = 1
		}

		user := User{
			Name:     r.FormValue("name"),
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
			Type:     usertype,
		}

		db.Create(&user)
		http.SetCookie(w, &http.Cookie{
			Name:   "auth",
			Value:  user.Email + "&" + user.Password,
			Path:   "/",
			MaxAge: 2592000,
		})
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	registerTmpl.Execute(w, map[string]interface{}{
		"user": context.Get(r, MyKey),
	})
}

func logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "auth",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusFound)
	return
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
	http.HandleFunc("/", withuser(index))
	http.HandleFunc("/ex", withuser(ex))
	http.HandleFunc("/login", withuser(login))
	http.HandleFunc("/register", withuser(register))
	http.HandleFunc("/logout", logout)
	http.ListenAndServe(":8080", nil)
}
