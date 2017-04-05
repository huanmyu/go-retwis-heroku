package main

import (
	"log"
	"net/http"
	"os"

	"github.com/urfave/negroni"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("public")))

	// Includes some default middlewares
	n := negroni.Classic()
	n.UseHandler(mux)
	n.Run(":" + port)
}
