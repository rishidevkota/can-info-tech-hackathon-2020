//Rishi Ram Devkota
//2020-03-07
//Hackathon Nepalgunj
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
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
	Name         string
	Email        string
	Password     string
	Type         int
	Reservations []Reservation
	Experience   *Experience
	Comments     []Comment
}

//Experience ...
type Experience struct {
	gorm.Model
	UserID       uint
	Title        string
	Location     string
	Duration     string
	Price        string
	Description  string
	Type         string
	User         User
	Reservations []Reservation
	Comments     []Comment
}

//Reservation ...
type Reservation struct {
	gorm.Model
	UserID       uint
	ExperienceID uint
	ArrivalDate  string
	Experience   Experience
	User         User
}

//Comment ...
type Comment struct {
	gorm.Model
	ExperienceID uint
	UserID       uint
	Text         string
	User         User
	Experience   Experience
}

//TimeFormat ...
func (c *Comment) TimeFormat() string {
	return c.CreatedAt.Format("02 Jan 2006")
}

type key int

const MyKey key = 0

var indexTmpl = template.Must(template.ParseFiles("base.html", "index.html"))
var loginTmpl = template.Must(template.ParseFiles("base.html", "login.html"))
var dashboardTmpl = template.Must(template.ParseFiles("base.html", "dashboard.html"))
var registerTmpl = template.Must(template.ParseFiles("base.html", "register.html"))
var exTmpl = template.Must(template.ParseFiles("base.html", "ex.html"))
var mrevTmpl = template.Must(template.ParseFiles("base.html", "mrev.html"))

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
	user := context.Get(r, MyKey)
	indexTmpl.Execute(w, map[string]interface{}{
		"experiences": experiences,
		"email":       email,
		"user":        user,
	})
}

func ex(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	var exp Experience
	db.First(&exp, id)
	db.Model(&exp).Related(&exp.Comments)
	for i := 0; i < len(exp.Comments); i++ {
		db.Model(&exp.Comments[i]).Related(&exp.Comments[i].User)
	}
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

func dashboard(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, MyKey).(User)
	var ex Experience
	db.Model(&user).Related(&ex)
	db.Model(&ex).Related(&ex.Reservations)
	for i := 0; i < len(ex.Reservations); i++ {
		db.Model(&ex.Reservations[i]).Related(&ex.Reservations[i].User)
	}

	fmt.Println(ex)

	if r.Method == "POST" {
		//db.Model(&user).Association("Experience").Append(&Experience{
		ex.Title = r.FormValue("title")
		ex.Location = r.FormValue("location")
		ex.Duration = r.FormValue("duration")
		ex.Price = r.FormValue("price")
		ex.Description = r.FormValue("description")
		if ex.ID == 0 {
			db.Model(&user).Association("Experience").Append(ex)
		} else {
			db.Save(ex)
		}
		//})
	}

	dashboardTmpl.Execute(w, map[string]interface{}{
		"exprience": ex,
		"user":      user,
	})
}

func mrev(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, MyKey).(User)
	db.Model(&user).Related(&user.Reservations)
	for i := 0; i < len(user.Reservations); i++ {
		db.Model(&user.Reservations[i]).Related(&user.Reservations[i].Experience)
	}

	mrevTmpl.Execute(w, map[string]interface{}{
		"user": user,
	})
}

func reserve(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, MyKey).(User)
	var ex Experience
	exid, _ := strconv.Atoi(r.FormValue("exid"))
	db.Find(&ex, exid)

	db.Model(&user).Association("Reservations").Append(&Reservation{
		ArrivalDate: r.FormValue("checkin"),
		Experience:  ex,
	})

	http.Redirect(w, r, "/myreservations", http.StatusFound)
	return
}

func upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("uploadfile")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()

	f, err := os.OpenFile("./public/img/"+r.FormValue("filename"), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()
	io.Copy(f, file)
	http.Redirect(w, r, "/dashboard", http.StatusFound)
	return
}

func comment(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, MyKey).(User)
	exid, _ := strconv.Atoi(r.FormValue("exid"))
	var ex Experience
	db.Find(&ex, exid)
	comment := Comment{
		Text:       r.FormValue("comment"),
		User:       user,
		Experience: ex,
	}
	db.Create(&comment)

	http.Redirect(w, r, "/ex?id="+r.FormValue("exid"), http.StatusFound)
}

func main() {
	var err error
	db, err = gorm.Open("sqlite3", "data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.AutoMigrate(&User{}, &Experience{}, &Reservation{}, &Comment{})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))
	http.HandleFunc("/", withuser(index))
	http.HandleFunc("/ex", withuser(ex))
	http.HandleFunc("/dashboard", withuser(dashboard))
	http.HandleFunc("/myreservations", withuser(mrev))
	http.HandleFunc("/reserve", withuser(reserve))
	http.HandleFunc("/login", withuser(login))
	http.HandleFunc("/register", withuser(register))
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/comment", withuser(comment))
	http.ListenAndServe(":8080", nil)
}
