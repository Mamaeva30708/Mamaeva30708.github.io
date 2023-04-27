package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

type UserInfo struct {
	UserId   int    `json:"user_id"`
	UserName string `json:"user_name"`
	Age      int    `json:"age"`
}

type server struct {
	db *sql.DB
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "12345"
	dbname   = "recipes"
)

func dbConnect() *server {
	dbconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", dbconn)
	fmt.Println("Opening database")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
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
	ageStr := r.FormValue("age")
	userID, err := createUser(userName, ageStr, s.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	age, err := strconv.Atoi(ageStr)
	if err != nil {
		log.Fatal("age", err)
	}

	person := UserInfo{
		UserId:   userID,
		UserName: userName,
		Age:      age,
	}

	fmt.Println(person)
	outputHTML(w, "./static/qweFinal.html", person)
}

func createUser(userName string, age string, db *sql.DB) (int, error) {
	user_id := 0
	err := db.QueryRow(`INSERT INTO "users"("username", "age") VALUES ($1, $2) returning id`, userName, age).Scan(&user_id)
	if err != nil {
		return 0, err
	}

	// user_id, err := res.LastInsertId()
	// if err != nil {
	// 	return 0, err
	// }

	return user_id, nil
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
	defer s.db.Close()
	http.ListenAndServe(":1806", nil)
}
