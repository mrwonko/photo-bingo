package main

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"log"
	"net/http"
)

//go:embed templates
var templateFS embed.FS

func main() {
	templates, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %s", err)
	}
	index := templates.Lookup("index.html")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		var buf bytes.Buffer
		err := index.Execute(&buf, nil)
		if err != nil {
			handleError(err, w)
			return
		}
		_, _ = io.Copy(w, &buf)
	})

	http.ListenAndServe("localhost:8081", mux)
}

func handleError(err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}
