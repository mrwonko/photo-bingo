package main

import (
	"bytes"
	"embed"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
)

type SignupData struct {
	Path string
}

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
	mustLookup := func(name string) *template.Template {
		res := templates.Lookup(name)
		if res == nil {
			log.Fatalf("Template %q is missing", name)
		}
		return res
	}
	index := mustLookup("index.html")
	signup := mustLookup("signup.html")

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logf("%s request to %s", r.Method, r.URL)

		//TODO if req.CookiesNamed("session_id")
		serveTemplate(w, signup, SignupData{
			Path: url.PathEscape(r.URL.Path),
		})
	})
	mux.HandleFunc("POST /signup", func(w http.ResponseWriter, r *http.Request) {
		logf("signup %q", r.FormValue("username"))
		// TODO
		serveTemplate(w, index, nil)
	})

	log.Print("serving")
	http.ListenAndServe("localhost:8081", mux)
}

func serveTemplate(w http.ResponseWriter, t *template.Template, data any) {
	logf("serving template %q with data %#v", t.Name(), data)
	var buf bytes.Buffer
	err := t.Execute(&buf, data)
	if err != nil {
		handleError(err, w)
		return
	}
	_, _ = io.Copy(w, &buf)
}

func handleError(err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

func logf(fmt string, values ...any) {
	if verbose {
		log.Printf(fmt, values...)
	}
}
