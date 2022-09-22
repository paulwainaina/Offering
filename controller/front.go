package controller

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"

	"example.com/session"
)

var (
	tpl *template.Template
)

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.html"))
}

type Page struct {
	Body  []byte
	Title string
}

type FrontController struct {
	sessionManager *session.SessionManager
}

func NewFrontController(s *session.SessionManager) *FrontController {
	return &FrontController{
		sessionManager: s,
	}
}

func (front FrontController) ServeHttp(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		{
			front.IndexPage(w)
		}
	case "/signin":
		{
			front.SigninPage(w)
		}

	}
}

func (front FrontController) LoadPage(file string) (*Page, error) {
	body, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return &Page{Body: body}, nil
}

func (front FrontController) IndexPage(w http.ResponseWriter) {
	file := "index.html"
	filePath := "templates/" + file
	pageName := "Home Page"
	page, err := front.LoadPage(filePath)
	if err != nil {
		page = &Page{Title: pageName}
	}
	page.Title = pageName
	front.RenderTemplate(w, file, page)
}
func (front FrontController) SigninPage(w http.ResponseWriter) {
	file := "signin.html"
	filePath := "templates/" + file
	pageName := "Signin Page"
	page, err := front.LoadPage(filePath)
	if err != nil {
		page = &Page{Title: pageName}
	}
	page.Title = pageName
	front.RenderTemplate(w, file, page)
}

func (front FrontController) RenderTemplate(w http.ResponseWriter, file string, page *Page) {
	err := tpl.ExecuteTemplate(w, file, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (front FrontController) SignedInMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(front.sessionManager.Sessions)
		val, err := r.Cookie("Session")
		if err != nil {
			http.Redirect(w, r, "/signin", http.StatusMovedPermanently)
			return
		}
		if !front.sessionManager.SessionExist(val.Value) {
			
			http.Redirect(w, r, "/signin", http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RegisterFrontController(s *session.SessionManager) {
	frontcontroller := NewFrontController(s)
	http.Handle("/", frontcontroller.SignedInMiddleware(http.HandlerFunc(frontcontroller.ServeHttp)))
	http.HandleFunc("/signin", frontcontroller.ServeHttp)
}
