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
func (s *server) selectUsers() []UserInfo {
	rows, err := s.db.Query("select id, userName, age from users;")
	if err != nil {
		log.Fatal(err)
	}

	var users []UserInfo
	for rows.Next() {
		var user UserInfo
		err := rows.Scan(&user.UserId, &user.UserName, &user.Age)
		if err != nil {
			log.Fatal("selectUsers", err)
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		log.Fatal("selectUsers2", err)
	}

	// fmt.Println(users)

	return users
}

func (s *server) selectUser(id int) UserInfo {
	rows := s.db.QueryRow("select id, userName, age from users where id=?;", id)

	var user UserInfo
	err := rows.Scan(&user.UserId, &user.UserName, &user.Age)
	if err != nil {
		log.Fatal("selectUsers", err)
	}

	return user
}

func (s *server) allUsersHandle(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./static/allUsers.html")
	if err != nil {
		log.Fatal("allUsersHandle", err)
	}

	allUsers := s.selectUsers()
	errExecute := t.Execute(w, allUsers)
	fmt.Println(allUsers[0].UserName)
	if errExecute != nil {
		log.Fatal("allUsersHandle2", err)
	}
}
func (s *server) updateUserByID(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	id := r.FormValue("id")
	idInt, err := strconv.Atoi(id)
	userName := r.FormValue("name")
	age := r.FormValue("age")
	updateUser(userName, age, idInt, s)
	http.Redirect(w, r, "/users", http.StatusSeeOther)
}

func (s *server) updateUserForm(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./static/update.html")
	if err != nil {
		log.Fatal("allUsersHandle", err)
	}

	err = r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	id := r.FormValue("id")
	idInt, err := strconv.Atoi(id)
	user := s.selectUser(idInt)

	t.Execute(w, user)
}

func (s *server) allUserChangeHandle(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./static/update.html")
	if err != nil {
		log.Fatal("allUsersHandle", err)
	}

	allUsers := s.selectUsers()
	errExecute := t.Execute(w, allUsers)
	// fmt.Println(allUsers[0].FullName)
	if errExecute != nil {
		log.Fatal("allUsersHandle2", err)
	}
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

/*func getUsers(db *sql.DB) ([]UserInfo, error) {
	rows, err := db.Query(`SELECT * FROM "users"`)
	if err != nil {
		log.Fatal(err)
	}

	var usersInfo []UserInfo
	for rows.Next() {
		var user UserInfo
		err := rows.Scan(&user.UserId, &user.UserName, &user.Age)
		if err != nil {
			return nil, err
		}
		usersInfo = append(usersInfo, user)
	}

	return usersInfo, nil
}
*/

func (s *server) updateUser(userName string, age int, id int) int {
	res, err := s.db.Exec("update users set userName=?, age=? where id=?", userName, age, id)
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

/*
	func (s *server) getUserHandle(w http.ResponseWriter, r *http.Request){
		users, err := getUsers(s.db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		t, err := template.ParseFiles("./static/allUsers.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := t.Execute(w, users); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

}
*/
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
	http.HandleFunc("/users", s.allUsersHandle)
	http.HandleFunc("/change", s.allUserChangeHandle)
	http.HandleFunc("/update", s.updateUserForm)
	http.HandleFunc("/delete", s.deleteUserHandle)
	// http.HandleFunc("/update", s.updateUser)
	http.HandleFunc("/update", s.updateUserForm)
	http.HandleFunc("/updateUserByID", s.updateUserByID)
	fmt.Print("Server is up and running...")
	defer s.db.Close()
	http.ListenAndServe(":1806", nil)
}

