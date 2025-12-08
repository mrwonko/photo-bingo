package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
)

type SignupData struct {
	RedirectPath string
	BaseURL      string
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

		user, err := checkAuth(r)
		if err != nil {
			serveError(w, http.StatusUnauthorized, err)
			return
		}
		if user == nil {
			serveTemplate(w, signup, SignupData{
				RedirectPath: url.PathEscape(r.URL.Path),
				BaseURL:      basePath,
			})
			return
		}
		logf("Authorized user %q", *user)
		serveTemplate(w, index, nil)
	})
	mux.HandleFunc("POST /signup", func(w http.ResponseWriter, r *http.Request) {
		logf("signup %q", r.FormValue("username"))
		path, err := url.PathUnescape(r.URL.Query().Get("path"))
		if err != nil {
			serveError(w, http.StatusBadRequest, fmt.Errorf("unescaping redirect destination: %w", err))
			return
		}
		if err := signUp(w, r); err != nil {
			serveError(w, http.StatusBadRequest, err)
			return
		}
		http.Redirect(w, r, basePath+path, http.StatusSeeOther)
	})

	log.Print("serving")
	http.ListenAndServe("localhost:8081", mux)
}

func serveTemplate(w http.ResponseWriter, t *template.Template, data any) {
	logf("serving template %q with data %#v", t.Name(), data)
	var buf bytes.Buffer
	err := t.Execute(&buf, data)
	if err != nil {
		serveError(w, http.StatusInternalServerError, err)
		return
	}
	_, err = io.Copy(w, &buf)
	if err != nil {
		logf("Failed to serve %q: %s", t.Name(), err)
	}
}

func serveError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte(err.Error()))
}

func logf(fmt string, values ...any) {
	if verbose {
		log.Printf(fmt, values...)
	}
}
