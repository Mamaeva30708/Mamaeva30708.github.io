package main

import (
	"database/sql"
	"encoding/json"
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

func getUser(userId int, db *sql.DB) (*UserInfo, error) {
	row := db.QueryRow(`SELECT * FROM "users" WHERE id=$1`, userId)

	user := &UserInfo{}
	err := row.Scan(&user.UserId, &user.UserName, &user.Age)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func updateUser(fullName string, role string, id int, s *server) int {
	res, err := s.db.Exec("update users set fullName=?, role=? where id=?", fullName, role, id)
	if err != nil {
		log.Fatal(err)
	}

	user_id, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}

	return int(user_id)
}

func (s *server) deleteUserHandle(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.URL.Query().Get("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = deleteUser(userId, s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func deleteUser(userId int, s *server) error {
	_, err := s.db.Exec(`DELETE FROM "users" WHERE "id"=$1`, userId)
	if err != nil {
		return err
	}

	return nil
}
func (s *server) getUserHandle(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.URL.Query().Get("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := getUser(userId, s.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
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
	http.HandleFunc("/user", s.getUserHandle)
	http.HandleFunc("/delete/", s.deleteUserHandle)
	fmt.Print("Server is up and running...")
	defer s.db.Close()
	http.ListenAndServe(":1806", nil)
}
