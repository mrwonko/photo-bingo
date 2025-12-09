package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strconv"
	"sync"
	"time"
)

type SignupData struct {
	RedirectPath string
	BaseURL      string
}

type GameData struct {
	BaseURL string
	User    PlayerName
	Board   DisplayBingoBoard
	Score   int
}

type SpaceData struct {
	BaseURL string
	Space   DisplayBingoSpace
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
	signup := mustLookup("signup.html")
	index := mustLookup("index.html")
	space := mustLookup("space.html")

	if _, err := os.Stat(imagePath); errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(imagePath, 0700)
		if err != nil {
			log.Fatalf("Failed to create image directory %q: %s", imagePath, err)
		}
	}

	err = loadState()
	if err != nil {
		log.Fatalf("failed to load state: %s", err)
	}

	saveTrigger := make(chan struct{}, 32) // keep a buffer to try to avoid blocking on high traffic

	mux := http.NewServeMux()

	fileServer := http.StripPrefix("/"+imagePath, http.FileServer(http.Dir(imagePath)))

	mux.Handle("GET /"+imagePath+"/", fileServer)

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
		gameData := GameData{
			BaseURL: basePath,
			User:    *user,
		}
		gameState.Read(func(gs GameState) {
			board := gs.Players[*user].Board
			gameData.Board = board.display()
			gameData.Score = board.score()
		})
		serveTemplate(w, index, gameData)
	})

	mux.HandleFunc("/spaces/{x}/{y}", func(w http.ResponseWriter, r *http.Request) {
		logf("%s request to %s", r.Method, r.URL)

		// TODO refactor this auth check
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

		x, err := strconv.Atoi(r.PathValue("x"))
		if err != nil || x < 0 || x >= 5 {
			serveError(w, http.StatusBadRequest, fmt.Errorf("invalid X value"))
		}
		y, err := strconv.Atoi(r.PathValue("y"))
		if err != nil || y < 0 || y >= 5 {
			serveError(w, http.StatusBadRequest, fmt.Errorf("invalid Y value"))
		}

		action := r.FormValue("action")
		uploadFileName := ""

		if action == "upload" {
			r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
			if err := r.ParseMultipartForm(maxUploadSize); err != nil {
				serveError(w, http.StatusBadRequest, err)
				return
			}
			srcFile, header, err := r.FormFile("image_file")
			if err != nil {
				serveError(w, http.StatusBadRequest, err)
				return
			}
			defer srcFile.Close()
			if ct := header.Header.Get("Content-Type"); ct != "image/jpeg" {
				serveError(w, http.StatusBadRequest, fmt.Errorf("upload must be of type image/jpeg, got %q", ct))
				return
			}
			encodedPlayerName := base64.URLEncoding.EncodeToString([]byte(*user))
			randSuffix, err := randStr(6)
			if err != nil {
				serveError(w, http.StatusInternalServerError, fmt.Errorf("generating file name: %w", err))
				return
			}
			uploadFileName = path.Join(imagePath, fmt.Sprintf("%s.%d.%d.%s.jpg", encodedPlayerName, x, y, randSuffix))
			dstFile, err := os.Create(uploadFileName)
			if err != nil {
				serveError(w, http.StatusInternalServerError, fmt.Errorf("failed to create file: %w", err))
				return
			}
			defer dstFile.Close()
			if _, err := io.Copy(dstFile, srcFile); err != nil {
				serveError(w, http.StatusInternalServerError, fmt.Errorf("failed to write file: %w", err))
				return
			}
			if err = dstFile.Close(); err != nil {
				serveError(w, http.StatusInternalServerError, fmt.Errorf("failed to finish writing file: %w", err))
				return
			}
			if err = srcFile.Close(); err != nil {
				serveError(w, http.StatusInternalServerError, fmt.Errorf("failed to close file upload: %w", err))
				return
			}
		}
		spaceData := SpaceData{
			BaseURL: basePath,
		}
		gameState.Modify(func(gs GameState) GameState {
			pd := gs.Players[*user]
			space := pd.Board.get(x, y)
			needsUpdate := true
			switch action {
			case "complete":
				space.Completed = true
			case "decomplete":
				space.Completed = false
			case "upload":
				space.Image = uploadFileName
				space.Completed = true
			default:
				needsUpdate = false
			}
			if needsUpdate {
				gs.Players[*user] = pd
			}
			spaceData.Space = space.display()
			return gs
		})
		serveTemplate(w, space, spaceData)
	})

	// TODO handle /spaces/$x/$y (with auth -> create auth middleware)

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
		// save new user
		saveTrigger <- struct{}{}
		http.Redirect(w, r, basePath+path, http.StatusSeeOther)
	})

	log.Print("serving")
	srv := &http.Server{
		Addr:    "localhost:8081",
		Handler: mux,
	}
	sigCtx, sigStop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer sigStop()
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := srv.ListenAndServe()
		logf("ListenAndServe exited: %s", err)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		saveState(sigCtx, saveTrigger)
	}()

	<-sigCtx.Done()
	log.Print("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		logf("shutdown err: %s", err)
	}
	wg.Wait()

	log.Print("goodbye")
}

func serveTemplate(w http.ResponseWriter, t *template.Template, data any) {
	//logf("serving template %q with data %#v", t.Name(), data)
	logf("serving template %q", t.Name())
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
