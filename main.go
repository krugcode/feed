package main

import (
	"database/sql"
	"log"
	"net/http"

	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Post struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`
	Subtitle string    `json:"subtitle"`
	Tags     []string  `json:"tags"`
	Content  string    `json:"content"`
	Slug     string    `json:"slug"`
	Created  time.Time `json:"created"`
}

type FrontMatter struct {
	Title    string   `yaml:"title"`
	Subtitle string   `yaml:"subtitle"`
	Tags     []string `yaml:"tags"`
}

type App struct {
	db *sql.DB
}

func main() {
	app := &App{}
	app.initDB()

	r := mux.NewRouter()

	// static assets
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))

	// routes
	r.HandleFunc("/", app.homePage).Methods("GET")

	log.Println("Server starting on :8090")
	log.Fatal(http.ListenAndServe(":8090", r))
}

func (app *App) initDB() {
	var err error
	app.db, err = sql.Open("sqlite3", "./blog.db")
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		subtitle TEXT,
		tags TEXT,
		content TEXT NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		created DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = app.db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

func (app *App) homePage(w http.ResponseWriter, r *http.Request) {
	component := HomePage()
	component.Render(r.Context(), w)
}
