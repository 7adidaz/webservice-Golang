package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var (
	db *sql.DB
)

type Blog struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}

type Error struct {
	message string
}

func logger(f http.HandlerFunc) http.HandlerFunc { // middleware
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s -> %s \n", r.Method, r.URL.Path)
		f(w, r)
	}
}

func main() {
	dbLoader()
	defer db.Close()

	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/api/v1/blog", logger(allBlogs)).Methods("GET")
	r.HandleFunc("/api/v1/blog", logger(postBlog)).Methods("POST")
	r.HandleFunc("/api/v1/blog/{blogId}", logger(getBlogById)).Methods("GET")

	http.ListenAndServe(":3000", r)
}

func dbLoader() {
	connectionString := "postgresql://dbeaver:0@localhost:5432/blogs"
	_db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("error with postgresql connection", err)
	}
	db = _db
}

func allBlogs(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * from blogs")
	if err != nil {
		log.Fatal("error with postgresql connection", err)
	}

	var blogs []Blog

	var blog Blog
	for rows.Next() {
		if err := rows.Scan(&blog.ID, &blog.Content); err != nil {
			var error Error
			error.message = "Error getting blogs"

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(error)
		}
		blogs = append(blogs, blog)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blogs)
}

func postBlog(w http.ResponseWriter, r *http.Request) {
	var blog Blog

	err := json.NewDecoder(r.Body).Decode(&blog)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println("blog", blog.Content)

	_, err = db.Exec("INSERT INTO blogs (content) VALUES ($1)", blog.Content)
	if err != nil {
		log.Fatalln("something went wrong inserting", err)
	}

	http.Redirect(w, r, "/api/v1/blogs", http.StatusOK)
}

func getBlogById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blogId := vars["blogId"]

	var blog Blog

	row := db.QueryRow("SELECT * from blogs WHERE id=$1", blogId)
	if err := row.Scan(&blog.ID, &blog.Content); err != nil {
		var error Error
		error.message = "Error getting blog with id " + blogId

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(error)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blog)
}
