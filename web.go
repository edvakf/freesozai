package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB

func main() {
	initDB()

	r := mux.NewRouter()
	r.HandleFunc("/", topHandler).Methods("GET")
	r.HandleFunc("/", createHandler).Methods("POST")
	r.HandleFunc("/{id:[0-9]+}", postHandler).Methods("GET")
	http.Handle("/", r)

	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func initDB() {
	var err error
	m := regexp.MustCompile("^mysql://(.+?):(.+?)@(.+?)/(.+?)\\?.*").FindStringSubmatch(os.Getenv("CLEARDB_DATABASE_URL"))
	log.Printf("%+v", m)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", m[1], m[2], m[3], 3306, m[4])
	db, err = sql.Open("mysql", dsn)

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS posts (id INT(10) UNSIGNED NOT NULL AUTO_INCREMENT, text MEDIUMTEXT, PRIMARY KEY (id))")
	if err != nil {
		panic(err.Error())
	}
}

func topHandler(w http.ResponseWriter, r *http.Request) {
	var id int
	err := db.QueryRow("SELECT id FROM posts ORDER BY id DESC LIMIT 1").Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			t := template.Must(template.New("html").Parse(`<!DOCTYPE html><h1>Welcome!</h1>`))
			t.ExecuteTemplate(w, "html", nil)
			return
		} else {
			serverError(w, err)
			return
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/%d", id), http.StatusFound)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	log.Printf("%s", body)
	if err != nil {
		serverError(w, err)
		return
	}
	res, err := db.Exec("INSERT INTO posts (text) VALUES (?)", string(body))
	if err != nil {
		serverError(w, err)
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		serverError(w, err)
		return
	}
	log.Printf("record created! id: %d", id)
	w.Write([]byte(fmt.Sprintf("http://%s/%d\n", r.Host, id)))
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	var text string
	vars := mux.Vars(r)
	err := db.QueryRow("SELECT text FROM posts WHERE id = ?", vars["id"]).Scan(&text)
	if err != nil {
		serverError(w, err)
		return
	}
	log.Printf("found post")
	t := template.Must(template.New("html").Parse(`<!DOCTYPE html><pre>{{.text}}</pre>`))
	t.ExecuteTemplate(w, "html", map[string]string{
		"text": text,
	})
}

func serverError(w http.ResponseWriter, err error) {
	log.Printf("error: %s", err)
	code := http.StatusInternalServerError
	http.Error(w, http.StatusText(code), code)
}
