package main

import (
	"embed"
	"fmt"
	"html/template"
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
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = fmt.Fprintf(w, "Request for %s %s\n", req.Method, req.URL)
		_, _ = fmt.Fprintf(w, "Templates:\n")
		for _, t := range templates.Templates() {
			_, _ = fmt.Fprintf(w, "%s\n", t.Name())
		}
	})

	http.ListenAndServe("localhost:8081", mux)
}
