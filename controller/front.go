package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"
	"time"

	"example.com/session"
)

var (
	tpl     *template.Template
	logedin =[]User{}
)

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.html"))
}

type Page struct {
	Body  []byte
	Title string
	Error error
}

type FrontController struct {
	sessionManager *session.SessionManager
}

func NewFrontController(s interface{}) *FrontController {
	return &FrontController{
		sessionManager: s.(*session.SessionManager),
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
			front.SigninPage(w, "")
		}
	case "/members":
		{
			front.MemberPage(w)
		}
	case "/loginHandler":
		{
			front.LoginHandler(w, r)
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
func (front FrontController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" {
		front.SigninPage(w, "Empty field detected")
		return
	}
	data := map[string]interface{}{"Email": email, "Password": password}
	js, err := json.Marshal(data)
	if err != nil {
		front.SigninPage(w, err.Error())
		return
	}
	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080/login", bytes.NewReader(js))
	if err != nil {
		fmt.Println(err)
		front.SigninPage(w, err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		front.SigninPage(w, err.Error())
		return
	}
	var user = User{}
	json.NewDecoder(res.Body).Decode(&user)
	logedin = append(logedin, user) 
	sess,err:=front.sessionManager.UserSession(user.ID)
	if err != nil {
		front.SigninPage(w, err.Error())
		return
	}
	cookie:=http.Cookie{Name:"Session",Value:sess}
	http.SetCookie(w,&cookie)
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
func (front FrontController) MemberPage(w http.ResponseWriter) {
	file := "member.html"
	filePath := "templates/" + file
	pageName := "Members Page"
	page, err := front.LoadPage(filePath)
	if err != nil {
		page = &Page{Title: pageName}
	}
	page.Title = pageName
	front.RenderTemplate(w, file, page)
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
func (front FrontController) SigninPage(w http.ResponseWriter, message string) {
	file := "signin.html"
	filePath := "templates/" + file
	pageName := "Signin Page"
	page, err := front.LoadPage(filePath)
	if err != nil {
		page = &Page{Title: pageName}
	}
	page.Title = pageName
	page.Error = fmt.Errorf(message)
	front.RenderTemplate(w, file, page)
}

func (front FrontController) RenderTemplate(w http.ResponseWriter, file string, page *Page) {
	err := tpl.ExecuteTemplate(w, file, page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (front *FrontController) SignedInMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val, err := r.Cookie("Session")

		if err != nil {
			http.Redirect(w, r, "/signin", http.StatusMovedPermanently)
			return
		}
		if !(*front).sessionManager.SessionExist(val.Value) {
			http.Redirect(w, r, "/signin", http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RegisterFrontController(s interface{}) {
	frontcontroller := NewFrontController(s)
	http.Handle("/", frontcontroller.SignedInMiddleware(http.HandlerFunc(frontcontroller.ServeHttp)))
	http.Handle("/members", frontcontroller.SignedInMiddleware(http.HandlerFunc(frontcontroller.ServeHttp)))
	http.HandleFunc("/signin", frontcontroller.ServeHttp)
	http.HandleFunc("/loginHandler", frontcontroller.ServeHttp)
}
