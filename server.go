package main

import (
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/bowenchen6/go-retwis-heroku/model"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

func registerHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Redirect(w, r, "/index.html", http.StatusFound)
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	password2 := r.FormValue("password2")

	if username == "" {
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	}

	if password == "" && password2 == "" {
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	}

	if password != password2 {
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	}
	user := model.User{
		Username: username,
		Password: password,
	}

	err = user.CreateUser()
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/index.html", http.StatusFound)
		return
	}

	t, _ := template.ParseFiles("public/register.html")
	t.Execute(w, &user)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	mux := mux.NewRouter()
	mux.Handle("/", http.FileServer(http.Dir("public")))
	mux.HandleFunc("/register", registerHandler)

	// Includes some default middlewares
	n := negroni.Classic()
	n.UseHandler(mux)
	n.Run(":" + port)
}
