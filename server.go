package main

import (
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/bowenchen6/go-retwis-heroku/model"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

type errs struct {
	RegisterErr string
	LoginErr    string
}

type userInfo struct {
	User           *model.User
	FollowingCount int
	FollowersCount int
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	u, err := IsLogin(w, r)
	if err != nil {
		err := r.ParseForm()
		if err != nil {
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, err.Error())
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" {
			errInfo := errs{
				LoginErr: "username must be set",
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
			return
		}

		if password == "" {
			errInfo := errs{
				LoginErr: "password must be set",
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
			return
		}

		u = model.User{
			Username: username,
		}

		err = u.GetUserByName()
		if err != nil {
			errInfo := errs{
				LoginErr: "user not found",
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
			return
		}

		if model.Md5PasswordWithSalt(u.Salt, password) != u.Password {
			errInfo := errs{
				LoginErr: "password not correct",
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
			return
		}

		oneYearAgo := time.Now().Add(time.Hour * 24 * 365)
		if cookie, err := r.Cookie("auth"); err != nil {
			http.SetCookie(w, &http.Cookie{Name: "auth", Value: u.Auth, Expires: oneYearAgo})
		} else {
			cookie.Value = u.Auth
			cookie.Expires = oneYearAgo
			http.SetCookie(w, cookie)
		}
	}

	data := userInfo{
		&u,
		len(u.Following),
		len(u.Followers),
	}

	t, _ := template.ParseFiles("template/home.html")
	t.Execute(w, data)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	u, err := IsLogin(w, r)
	if err != nil {
		err := r.ParseForm()
		if err != nil {
			errInfo := errs{
				RegisterErr: err.Error(),
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
		}

		username := r.FormValue("username")
		password := r.FormValue("password")
		password2 := r.FormValue("password2")

		if username == "" {
			errInfo := errs{
				RegisterErr: "username must be set",
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
			return
		}

		if password == "" || password2 == "" {
			errInfo := errs{
				RegisterErr: "password or repeat password must be set",
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
			return
		}

		if password != password2 {
			errInfo := errs{
				RegisterErr: "password not same",
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
			return
		}

		u := model.User{
			Username: username,
			Password: password,
		}

		err = u.CreateUser()
		if err != nil {
			errInfo := errs{
				RegisterErr: err.Error(),
			}
			t, _ := template.ParseFiles("template/index.html")
			t.Execute(w, &errInfo)
			return
		}

		oneYearAgo := time.Now().Add(time.Hour * 24 * 365)
		if cookie, err := r.Cookie("auth"); err != nil {
			http.SetCookie(w, &http.Cookie{Name: "auth", Value: u.Auth, Expires: oneYearAgo})
		} else {
			cookie.Value = u.Auth
			cookie.Expires = oneYearAgo
			http.SetCookie(w, cookie)
		}

		t, _ := template.ParseFiles("template/register.html")
		t.Execute(w, &u)
		return
	}
	data := userInfo{
		&u,
		len(u.Following),
		len(u.Followers),
	}

	t, _ := template.ParseFiles("template/home.html")
	t.Execute(w, data)
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/favicon.ico")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	u, err := IsLogin(w, r)
	if err != nil {
		t, _ := template.ParseFiles("template/index.html")
		t.Execute(w, nil)
		return
	}

	data := userInfo{
		&u,
		len(u.Following),
		len(u.Followers),
	}

	t, _ := template.ParseFiles("template/home.html")
	t.Execute(w, data)
}

func IsLogin(w http.ResponseWriter, r *http.Request) (model.User, error) {
	u := model.User{}
	if cookie, err := r.Cookie("auth"); err != nil {
		return u, err
	} else {
		u.Auth = cookie.Value
		err = u.GetUserByAuth()
		if err != nil {
			return u, err
		}

		return u, nil
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	mux := mux.NewRouter()
	mux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/favicon.ico", faviconHandler)
	mux.HandleFunc("/register", registerHandler)
	mux.HandleFunc("/login", loginHandler)

	// Includes some default middlewares
	n := negroni.Classic()
	n.UseHandler(mux)
	n.Run(":" + port)
}
