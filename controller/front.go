package controller

import (
	"io/ioutil"
	"net/http"
	"text/template"
)

var (
	tpl * template.Template
)

func init(){
	tpl=template.Must(template.ParseGlob("templates/*.html"))
}

type Page struct{
	Body []byte
	Title string
}

type FrontController struct{

}

func NewFrontController()(*FrontController){
	return &FrontController{}
}

func (front FrontController)ServeHttp(w http.ResponseWriter,r * http.Request){
	switch r.URL.Path{
	case "/":{
		front.IndexPage(w)
	}

	}
}

func (front FrontController) LoadPage(file string)(*Page,error){
	body,err:=ioutil.ReadFile(file)
	if err!=nil{
		return nil,err
	}
	return &Page{Body:body},nil
}

func (front FrontController) IndexPage(w http.ResponseWriter){
	file:="index.html"
	filePath:="templates/"+file
	pageName:="Home Page"
	page,err:=front.LoadPage(filePath)
	if err !=nil{
		page=&Page{Title:pageName}
	}
	page.Title=pageName
	front.RenderTemplate(w,file,page)
}

func (front FrontController)RenderTemplate(w http.ResponseWriter,file string,page *Page){
	err:=tpl.ExecuteTemplate(w,file,page)
	if err !=nil{
		http.Error(w,err.Error(),http.StatusInternalServerError)
	}
}

func RegisterFrontController() {
	frontcontroller := NewFrontController()
	http.HandleFunc("/",frontcontroller.ServeHttp)	
}
