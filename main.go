package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var port int

func init() {
	flag.IntVar(&port, "p", 9000, "port to serve")
}

type payload struct {
	Ref    string
	Pusher struct {
		Name, Email string
	}
	Commit struct {
		ID, Message, Timestamp string
	} `json:"head_commit"`
}

type app struct {
	Webhooks []*hook
}

func main() {
	flag.Parse()
	app := load(flag.Args())

	http.Handle("/", recoverable(timing(http.HandlerFunc(app.gitwebhooks))))

	log.Println("listening for github webhooks on:", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func load(args []string) *app {
	if len(args) == 0 {
		panic("need to provide webhook configuration json, as first argument")
	}

	a := &app{}
	file, err := os.Open(args[0])
	if err != nil {
		panic(err)
	}

	if err := json.NewDecoder(file).Decode(&a); err != nil {
		panic(err)
	}

	if len(a.Webhooks) == 0 {
		panic("could not find any webhooks, in given configuration file: " + args[0])
	}

	log.Println("note, that every path og webhook, needs to have id - loaded webhooks:")
	for _, h := range a.Webhooks {
		log.Println("  ->", h.ID, "-", h.Command.Exec, "at", h.Command.Workdir)
	}

	return a
}

func (a *app) gitwebhooks(w http.ResponseWriter, req *http.Request) {
	// ensure request method is POST
	if req.Method != "POST" {
		http.Error(w, http.StatusText(405), 405)
		return
	}

	// ensure hook id matches
	id := strings.Trim(req.URL.Path, "/")
	h := a.find(id)
	if h == nil {
		log.Println("could not locate webhook, by given id:", id)
		http.Error(w, http.StatusText(404), 404)
		return
	}

	// ensure json payload
	if !strings.Contains(req.Header.Get("Content-Type"), "json") {
		log.Println("only json payload is supported")
		http.Error(w, http.StatusText(406), 406)
		return
	}

	// read body data from request
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println("failed to read request body:", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// ensure is authorized
	if err := h.authorized(req, body); err != nil {
		log.Println("unauthorized:", err)
		http.Error(w, http.StatusText(401), 401)
		return
	}

	// read payload to struct
	var commit payload
	if err := json.Unmarshal(body, &commit); err != nil {
		log.Println("failed to unmarshal payload to struct:", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// only master branch changes are taken to account
	ref := strings.Split(commit.Ref, "/")
	if ref[len(ref)-1] != "master" {
		log.Println("skipping command, since not a master branch:", commit.Ref)
		w.Write([]byte("Skipped"))
		return
	}

	// execute the hook command
	if err := h.run(&commit); err != nil {
		log.Println("failed to execute hook", id, "command:", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.Write([]byte("OK"))
}

func (a *app) find(id string) *hook {
	for _, h := range a.Webhooks {
		if h.ID == id {
			return h
		}
	}
	return nil
}
