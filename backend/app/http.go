package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"

	"github.com/tdewolff/minify/html"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tdewolff/minify"
)

const (
	Accept        = "Accept"
	Authorization = "Authorization"
	ContentType   = "Content-Type"
)

const (
	Html     = "text/html"
	Json     = "application/json;charset=UTF-8"
	PngImage = "image/png"
)

func Serve() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ReturnMessage(w, "server up", http.StatusOK)
	})
	r.Route("/v1", func(r chi.Router) {
		r.Get("/info", GetInfoHandler)
		r.Get("/free-customers", GetFreeCustomerHandler)
		r.Get("/file-upload", GetUploadHandler)
		r.Post("/file-upload/confirm", ConfirmUploadHandler)
		r.Post("/upload-csv", UploadCSVHandler)
		r.Post("/assign-referral", AssignReferralHandler)
		r.Post("/add-payment", AddPaymentHandler)
	})
	r.Route("/page", func(r chi.Router) {
		r.Get("/{page}", PageHandler)
	})

	fs := http.FileServer(http.Dir("public"))
	r.HandleFunc("/view/html", ForbiddenHandler)
	r.HandleFunc("/view/tmpl", ForbiddenHandler)
	r.Handle("/view/*", http.StripPrefix("/view/", fs))

	log.Printf("running http server on %s\n", Config().HTTPPort)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", Config().HTTPPort),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Minute,
	}
	log.Fatalln(srv.ListenAndServe())
}

func ForbiddenHandler(w http.ResponseWriter, r *http.Request) {
	ReturnMessage(w, "forbidden", http.StatusForbidden)
}

func PageHandler(w http.ResponseWriter, r *http.Request) {
	page := chi.URLParam(r, "page")
	tmpl, err := template.ParseFiles(fmt.Sprintf("public/tmpl/%s.html", page))
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusNotFound)
		return
	}

	var b bytes.Buffer
	err = tmpl.Execute(&b, struct {
		BaseURL string
	}{BaseURL: Config().BaseURL})
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusNotFound)
		return
	}

	m := minify.New()
	m.AddFunc(Html, html.Minify)
	m.AddFunc(Html, css.Minify)
	m.AddFunc(Html, js.Minify)

	mb, err := m.Bytes(Html, b.Bytes())
	if err != nil {
		ReturnMessage(w, err.Error(), http.StatusNotFound)
		return
	}

	ReturnRaw(w, mb, http.StatusOK)
}

func ReturnRaw(w http.ResponseWriter, raw []byte, statusCode int) {
	w.Header().Set(ContentType, Html)
	w.WriteHeader(statusCode)
	w.Write(raw)
}

func ReturnImage(w http.ResponseWriter, buf []byte, statusCode int) {
	w.Header().Set(ContentType, PngImage)
	w.WriteHeader(statusCode)
	w.Write(buf)
}

func ReturnJson(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set(ContentType, Json)
	w.WriteHeader(statusCode)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func ReturnMessage(w http.ResponseWriter, message string, statusCode int) {
	ReturnJson(w, messageResponse{Message: message}, statusCode)
}

func ReturnBool(w http.ResponseWriter, success bool, statusCode int) {
	ReturnJson(w, successResponse{Success: success}, statusCode)
}

type successResponse struct {
	Success bool `json:"success"`
}

type messageResponse struct {
	Message string `json:"message"`
}
