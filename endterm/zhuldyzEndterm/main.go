package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"net/http"
)

type server struct {
	db *sql.DB
}

type Recipe struct {
	ID       int
	Url      string
	Name     string
	Duration int
	Portion  int
}

type Post struct {
	User_id int
	Post_id int
	Name    string
}

const (
	host     = "localhost"
	port     = 5433
	user     = "postgres"
	password = "12345"
	dbname   = "recipe"
)

func database() *server {
	dbconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", dbconn)
	if err != nil {
		log.Fatal(err)
	}
	return &server{db: db}
}

func main() {
	db := database()
	defer db.db.Close()

	Fl := http.FileServer(http.Dir("./static"))
	http.Handle("/", Fl)
	http.HandleFunc("/recipes", db.recipesPage)
	http.HandleFunc("/add", db.addPage)
	http.HandleFunc("/update", db.updatePage)
	http.HandleFunc("/delete", db.deleteRecipe)

	http.HandleFunc("/register", db.registerPage)
	http.HandleFunc("/authorization", db.authorizationPage)
	http.HandleFunc("/deleteuser", db.deleteUser)
	http.HandleFunc("/updateuser", db.updateUser)

	http.HandleFunc("/save", db.savePage)
	http.HandleFunc("/deletepost", db.deletePost)

	http.HandleFunc("/admin", adminPape)

	http.ListenAndServe(":1234", nil)
}

func (s *server) recipesPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		var recipes []Recipe
		result, err := s.db.Query("SELECT * FROM recipes;")
		if err != nil {
			log.Fatal(err)
		}
		for result.Next() {
			var recipe Recipe
			result.Scan(&recipe.ID, &recipe.Url, &recipe.Name, &recipe.Duration, &recipe.Portion)
			recipes = append(recipes, recipe)
		}
		t, _ := template.ParseFiles("static/html/recipes.html")
		t.Execute(w, recipes)
		return
	}
	fmt.Print(r.FormValue("id"))
}

func (s *server) addPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		t, err := template.ParseFiles("static/html/add.html")
		if err != nil {
			log.Fatal(err)
		}
		t.Execute(w, nil)
		return
	}
	if err := r.ParseForm(); err != nil {
		log.Fatal(err)
	}
	photo := r.FormValue("photo")
	name := r.FormValue("name")
	dur := r.FormValue("dur")
	por := r.FormValue("por")
	if _, err := s.db.Exec("INSERT INTO recipes(url, name, duration, portion) VALUES($1, $2, $3, $4)", photo, name, dur, por); err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/recipes", http.StatusSeeOther)
}

func (s *server) updatePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		t, _ := template.ParseFiles("static/html/update.html")
		t.Execute(w, nil)
		return
	}
	id := r.FormValue("id")
	url := r.FormValue("url")
	name := r.FormValue("name")
	dur := r.FormValue("dur")
	por := r.FormValue("por")
	if _, err := s.db.Exec("UPDATE recipes set url=$1, name=$2, duration=$3, portion=$4 where id=$5", url, name, dur, por, id); err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/recipes", http.StatusSeeOther)
}

func (s *server) deleteRecipe(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		t, _ := template.ParseFiles("static/html/delete.html")
		t.Execute(w, nil)
		return
	}
	id := r.FormValue("id")
	if _, err := s.db.Exec("delete from recipes where id=$1", id); err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/recipes", http.StatusSeeOther)
}

func (s *server) registerPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		t, _ := template.ParseFiles("static/html/register.html")
		t.Execute(w, nil)
		return
	}
	username := r.FormValue("username")
	age := r.FormValue("age")
	if _, err := s.db.Exec("INSERT INTO users(username, age) VALUES($1, $2)", username, age); err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/authorization", http.StatusSeeOther)
}

func (s *server) authorizationPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		t, _ := template.ParseFiles("static/html/authorization.html")
		t.Execute(w, nil)
		return
	}
	username := r.FormValue("username")
	if username == "Zhuldyz" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	var user_id int
	if err := s.db.QueryRow("select id from users where username=$1", username).Scan(&user_id); err != nil {
		log.Fatal(err)
	}

	var posts []Post
	res, err := s.db.Query("SELECT * from saved_posts where user_id=$1", user_id)
	if err != nil {
		log.Fatal(err)
	}
	for res.Next() {
		var post Post
		res.Scan(&post.User_id, &post.Post_id)
		if err := s.db.QueryRow("SELECT name FROM recipes where id=$1", post.Post_id).Scan(&post.Name); err != nil {
			log.Fatal(err)
		}

		posts = append(posts, post)
	}

	t, _ := template.ParseFiles("static/html/posts.html")
	t.Execute(w, posts)
}

func (s *server) savePage(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("username") == "" {
		post_id := r.FormValue("id")
		data := map[string]interface{}{"id": post_id}
		t, _ := template.ParseFiles("static/html/save.html")
		t.Execute(w, data)
		return
	}
	var user_id int
	post_id := r.FormValue("id")
	username := r.FormValue("username")
	err := s.db.QueryRow("SELECT id FROM users WHERE username=$1", username).Scan(&user_id)
	if err != nil {
		log.Fatal(err)
	}
	_, err = s.db.Exec("INSERT INTO saved_posts(user_id, post_id) VALUES($1, $2)", user_id, post_id)
	http.Redirect(w, r, "/recipes", http.StatusSeeOther)
}

func (s *server) deletePost(w http.ResponseWriter, r *http.Request) {
	user_id := r.FormValue("user")
	post_id := r.FormValue("post")
	if _, err := s.db.Exec("delete from saved_posts where user_id=$1 and post_id=$2", user_id, post_id); err != nil {
		log.Fatal(err)
	}

	var posts []Post
	res, err := s.db.Query("SELECT * from saved_posts where user_id=$1", user_id)
	if err != nil {
		log.Fatal(err)
	}
	for res.Next() {
		var post Post
		res.Scan(&post.User_id, &post.Post_id)
		if err := s.db.QueryRow("SELECT name FROM recipes where id=$1", post.Post_id).Scan(&post.Name); err != nil {
			log.Fatal(err)
		}

		posts = append(posts, post)
	}

	t, _ := template.ParseFiles("static/html/posts.html")
	t.Execute(w, posts)
}

func adminPape(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("static/html/admin.html")
	t.Execute(w, nil)
}

func (s *server) deleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		t, _ := template.ParseFiles("static/html/deleteUser.html")
		t.Execute(w, nil)
		return
	}
	username := r.FormValue("username")
	if _, err := s.db.Exec("delete from users where username=$1", username); err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (s *server) updateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		t, _ := template.ParseFiles("static/html/updateUser.html")
		t.Execute(w, nil)
		return
	}
	id := r.FormValue("id")
	username := r.FormValue("username")
	age := r.FormValue("age")
	if _, err := s.db.Exec("update users set username=$1, age=$2 where id=$3", username, age, id); err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
