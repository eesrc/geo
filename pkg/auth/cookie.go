package auth

import (
	"errors"
	"net/http"
	"time"
)

func getSessionProfileFromCookie(w http.ResponseWriter, r *http.Request, sessionStore SessionStore) (Profile, error) {
	cookie, err := r.Cookie(AuthCookieName)
	if cookie == nil || err == http.ErrNoCookie || cookie.Value == "" {
		return Profile{}, errors.New("no session cookie")
	}
	if err != nil {
		RemoveAuthCookie(w)
		return Profile{}, err
	}

	sessionID := cookie.Value

	sess, err := sessionStore.GetSession(sessionID, time.Now().UnixNano())
	if err != nil {
		RemoveAuthCookie(w)
		return Profile{}, err
	}
	return sess.Profile, nil
}

func RemoveAuthCookie(w http.ResponseWriter) {
	// Remove cookie by overwriting cookie
	rmCookie := &http.Cookie{
		Name:     AuthCookieName,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, rmCookie)
}
