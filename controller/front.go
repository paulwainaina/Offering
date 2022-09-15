package controller

import "net/http"

type FrontController struct{

}

func NewFrontController()(*FrontController){
	return &FrontController{}
}

func (front FrontController)ServeHttp(w http.ResponseWriter,r * http.Request){
	switch r.URL.Path{
	case "/":{
		w.Write([]byte("welcome"))
	}
	
	}
}

func RegisterFrontController() {
	frontcontroller := NewFrontController()
	http.HandleFunc("/",frontcontroller.ServeHttp)	
}
