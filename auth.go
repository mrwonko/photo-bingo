package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const authCookie = "session_id"

var authEncoding = base64.URLEncoding

// checkAuth looks for the [authCookie] and validates it, if present.
// If not present, it returns nil, nil.
func checkAuth(r *http.Request) (*PlayerName, error) {
	cookies := r.CookiesNamed(authCookie)
	if len(cookies) == 0 {
		return nil, nil
	}
	var errs []error
	for i, c := range cookies {
		tokenJSON, err := authEncoding.DecodeString(c.Value)
		if err != nil {
			errs = append(errs, fmt.Errorf("%d: decoding session ID: %w", i, err))
			continue
		}
		var token InsecurePlaintextAuthToken
		err = json.Unmarshal(tokenJSON, &token)
		if err != nil {
			errs = append(errs, fmt.Errorf("%d: parsing session ID: %w", i, err))
			continue
		}
		gameState.Read(func(gs GameState) {
			p, ok := gs.Players[token.User]
			if !ok {
				err = fmt.Errorf("%d: unknown user name: %s", i, token.User)

			} else if token.Password != p.Password { // FIXME this obviously cannot stay, add some kind of SSO or something
				err = fmt.Errorf("%d: invalid password", i)
			}
		})
		if err != nil {
			errs = append(errs, err)
			continue
		}
		token.Password = ""
		return &token.User, nil
	}
	return nil, errors.Join(errs...)
}

func signUp(w http.ResponseWriter, r *http.Request) error {
	token := InsecurePlaintextAuthToken{
		User: PlayerName(r.FormValue("username")),
	}
	if token.User == "" {
		return errors.New("no username provided")
	}
	var pwBytes [10]byte
	_, err := rand.Read(pwBytes[:])
	if err != nil {
		return fmt.Errorf("generating password: %w", err)
	}
	token.Password = InsecurePlaintextPassword(authEncoding.EncodeToString(pwBytes[:]))
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("encoding session ID: %w", err)
	}
	tokenB64 := base64.URLEncoding.EncodeToString(tokenJSON)
	gameState.Modify(func(gs GameState) GameState {
		_, ok := gs.Players[token.User]
		if ok {
			err = fmt.Errorf("user name %q already taken", token.User)
			return gs
		}
		ps := PlayerState{
			Password: token.Password,
			Approved: false,
			Board:    generateBoard(),
		}
		if gs.Players == nil {
			gs.Players = map[PlayerName]PlayerState{
				token.User: ps,
			}
		} else {
			gs.Players[token.User] = ps
		}
		return gs
	})
	if err != nil {
		return err
	}
	logf("Signed up user %q", token.User)
	http.SetCookie(w, &http.Cookie{
		Name:     authCookie,
		Value:    tokenB64,
		HttpOnly: true, // inaccessible to JS
		Secure:   true, // HTTPS-only (to slightly mitigate how insecure it is)
		Path:     basePath,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().AddDate(0, 1, 0),
	})
	return nil
}

type InsecurePlaintextAuthToken struct {
	User     PlayerName                `json:"u"`
	Password InsecurePlaintextPassword `json:"p"`
}
