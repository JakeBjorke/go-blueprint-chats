package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"trace"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

//ServeHTTP handles the http Request
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})

	t.templ.Execute(w, r)
}

func main() {
	//get the address from the command line arguments
	var addr = flag.String("addr", ":8080", "The addr of the application")
	flag.Parse()

	r := newRoom()
	r.tracer = trace.New(os.Stdout)

	//direct to the chat client
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	//route to the login page
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	//Bind the room path to the r object
	http.Handle("/room", r)

	//start the room in a go-routine
	go r.run()

	log.Println("Starting web server on ", *addr)

	//start the web server
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:  ", err)
	}
}
