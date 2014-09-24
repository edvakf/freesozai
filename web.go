package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"text/template"

	"github.com/fzzy/radix/redis"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", topHandler).Methods("GET")
	r.HandleFunc("/", createHandler).Methods("POST")
	r.HandleFunc("/{key:[0-9a-f]+}", postHandler).Methods("GET")
	http.Handle("/", r)

	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func redisClient() (*redis.Client, error) {
	m := regexp.MustCompile("^redis://redistogo:(.+?)@(.+?)/").FindStringSubmatch(os.Getenv("REDISTOGO_URL"))
	log.Printf("%+v", m)
	cli, err := redis.Dial("tcp", m[2])
	if err != nil {
		return nil, err
	}
	_, err = cli.Cmd("AUTH", m[1]).Bool()
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func topHandler(w http.ResponseWriter, r *http.Request) {
	cli, err := redisClient()
	if err != nil {
		serverError(w, err)
		return
	}
	defer cli.Close()
	rep := cli.Cmd("RANDOMKEY")
	if rep.Type == redis.NilReply {
		log.Printf("empty database")
		t := template.Must(template.New("html").Parse(`<!DOCTYPE html><h1>Welcome!</h1>`))
		t.ExecuteTemplate(w, "html", nil)
		return
	}
	key, err := rep.Str()
	if err != nil {
		serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/%s", key), http.StatusFound)
}

func md5hash(text []byte) string {
	h := md5.New()
	h.Write(text)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		serverError(w, err)
		return
	}
	cli, err := redisClient()
	if err != nil {
		serverError(w, err)
		return
	}
	defer cli.Close()
	key := md5hash(body)
	cli.Cmd("MULTI")
	cli.Cmd("SET", key, body)
	cli.Cmd("EXPIRE", key, 60*60*24*7)
	rep := cli.Cmd("EXEC")
	log.Printf("key: %s %s", key, rep.String())
	if rep.Err != nil {
		serverError(w, err)
		return
	}
	log.Printf("record created! id: %s", key)
	w.Write([]byte(fmt.Sprintf("http://%s/%s\n", r.Host, key)))
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	cli, err := redisClient()
	if err != nil {
		serverError(w, err)
		return
	}
	defer cli.Close()
	vars := mux.Vars(r)
	text, err := cli.Cmd("GET", vars["key"]).Str()
	if err != nil {
		serverError(w, err)
		return
	}
	log.Printf("found post for key %s", vars["key"])
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
