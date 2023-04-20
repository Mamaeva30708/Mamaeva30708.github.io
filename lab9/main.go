package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"recipe/sql"

	_ "github.com/mattn/go-sqlite3"
)

type UserInfo struct {
	UserId   int    `json:"user_id"`
	UserName string `json:"user_name"`
	Age      int    `json:"age"`
}

type server struct {
	db *sql.DB
}

func dbConnect() server {
	db, err := sql.Open("sqlite3", "recipe.sql")
	fmt.Println("Opening database")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully connected to database")

	return &server{db: db}

}

func (s *server) formHandle(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userName := r.FormValue("name")
	age := r.FormValue("age")
	userID, err := createUser(userName, age, s.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	person := UserInfo{
		UserID:   userID,
		UserName: userName,
		Age:      age,
	}

	fmt.Println(person)
	outputHTML(w, "./static/qwe.html", person)
}

func createUser(userName string, age string, db *sql.DB) (int, error) {
	res, err := db.Exec("INSERT INTO users(userName, age) VALUES (?, ?)", userName, age)
	if err != nil {
		return 0, err
	}

	user_id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(userID), nil
}

func outputHTML(w http.ResponseWriter, filename string, person UserInfo) {
	t, err := template.ParseFiles(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := t.Execute(w, person); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
func main() {
	s := dbConnect()
	defer s.db.Close()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	http.HandleFunc("/form", s.formHandle)

	fmt.Print("Server is up and running...")
	http.ListenAndServe(":1806", nil)
}
