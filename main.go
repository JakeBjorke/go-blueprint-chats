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
	"github.com/stretchr/objx"
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

	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	t.templ.Execute(w, r)
}

func main() {
	//get the address from the command line arguments
	var addr = flag.String("addr", ":8080", "The addr of the application")
	flag.Parse()

	gomniauth.SetSecurityKey("wHR9a0Kt20jjzVQnpkdBFSPdhEmyuV6guB2Nn5LVFSH0yv0Q9skPOBK8wcKahXC")
	gomniauth.WithProviders(
		google.New("578717759270-h4oc92fuisc83e75ql95et0960j0l376.apps.googleusercontent.com",
			"aAqJpqf8bGPKZk1UznGL6FbF",
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
