package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("static"))

	/* http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		w.Write([]byte("Cat!"))
	})
	*/
	http.Handle("/", fs)

	log.Fatal(http.ListenAndServe(":1806", nil))
}
