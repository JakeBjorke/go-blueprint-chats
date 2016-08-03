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

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/google"
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

	gomniauth.SetSecurityKey("wBpme39jkn8gwwgJFkPLUns4HSVvr2h3AtTRsNv32VvYrxrmavNAtjAd0Ae269V")
	gomniauth.WithProviders(
		google.New("827837591987-hbq5h5hh9f15f50b5fromqs2fsd2rg6t.apps.googleusercontent.com",
			"JTR6g8-6UlzCPVl2XCWd_QVt",
			"http://localhost:8080/auth/callback/google"),
	)

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
