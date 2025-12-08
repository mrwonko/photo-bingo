package main

import (
	"bytes"
	"embed"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
)

//go:embed templates
var templateFS embed.FS

func main() {
	templates, err := template.New("").Funcs(template.FuncMap{
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call with odd number of args")
			}
			dict := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}).ParseFS(templateFS, "templates/*.html")
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
