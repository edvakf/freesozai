package main

import (
	"database/sql"
	"fmt"
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
	r.HandleFunc("/{id:[0-9+]}", postHandler).Methods("GET")
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
	t := template.Must(template.New("html").Parse(`
<!DOCTYPE html>
<form method="post">
<textarea name="text"></textarea>
<button>post</button>
</form>`))
	t.ExecuteTemplate(w, "html", nil)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	res, err := db.Exec("INSERT INTO posts (text) VALUES (?)", r.PostFormValue("text"))
	if err != nil {
		serverError(w, err)
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		serverError(w, err)
		return
	}
	log.Printf("record created! redirecting to /%d", id)
	http.Redirect(w, r, fmt.Sprintf("/%d", id), http.StatusFound)
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
	t := template.Must(template.New("html").Parse(`
<!DOCTYPE html>
<pre>
{{.text}}
</pre>`))
	t.ExecuteTemplate(w, "html", map[string]string{
		"text": text,
	})
}

func serverError(w http.ResponseWriter, err error) {
	log.Printf("error: %s", err)
	code := http.StatusInternalServerError
	http.Error(w, http.StatusText(code), code)
}
