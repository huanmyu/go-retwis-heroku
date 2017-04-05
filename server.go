package main

import (
	"fmt"
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
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hello World!")
	})

	n := negroni.Classic() // Includes some default middlewares
	n.UseHandler(mux)
	n.Run(":" + port)
}
