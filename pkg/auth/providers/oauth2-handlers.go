package providers

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/auth"
	"golang.org/x/oauth2"
)

const (
	// sessionCheckInterval is the frequency of session checks
	sessionCheckInterval = time.Minute * 5
	// defaultSessionLength is the initial length of each session. It's much longer
	// than the check interval to spread the number of checks out. It's the fallback if the
	// token doesn't have a expiry
	defaultSessionLength = sessionCheckInterval * 12
)

func handleOAuth2Login(w http.ResponseWriter, r *http.Request, p *OAuth2Provider) {
	if cookie, err := r.Cookie(auth.AuthCookieName); err == nil {
		log.Warn("Trying to log in while still having an active auth cookie. Removing old session")
		_ = p.sessions.RemoveSession(cookie.Value)
	}

	// Create and persist state
	state := newState()
	if err := p.sessions.PutState(state); err != nil {
		log.WithError(err).Error("Unable to persist state for OAuth token")
		p.serverErrorHandleFunc(w, r, auth.PersistStateError)
		return
	}
	url := p.authConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleOauth2Logout(w http.ResponseWriter, r *http.Request, p *OAuth2Provider) {

	// Remove session cookie, delete tokens
	cookie, err := r.Cookie(auth.AuthCookieName)
	if err != nil {
		log.WithError(err).Error("Got error retrieving cookie")
		http.Redirect(w, r, p.Config.LogoutSuccessURL, http.StatusTemporaryRedirect)
		return
	}

	sessionID := cookie.Value
	_, err = p.sessions.GetSession(sessionID, time.Now().UnixNano())
	if err == nil {
		if err := p.sessions.RemoveSession(sessionID); err != nil {
			log.WithError(err).Error("Couldn't remove session")
		}
	}
	auth.RemoveAuthCookie(w)
	http.Redirect(w, r, p.Config.LogoutSuccessURL, http.StatusTemporaryRedirect)
}

func handleOAuth2Callback(w http.ResponseWriter, r *http.Request, p *OAuth2Provider) {
	if errCode := r.URL.Query().Get("error"); errCode != "" {
		p.serverErrorHandleFunc(w, r, createAuthError(errCode))
		return
	}

	requestState := r.URL.Query().Get("state")

	if err := p.sessions.RemoveState(requestState); err != nil {
		p.serverErrorHandleFunc(w, r, auth.UnknownStateError)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		p.serverErrorHandleFunc(w, r, auth.MissingCodeError)
		return
	}

	ctx := context.Background()

	token, err := p.authConfig.Exchange(ctx, code)
	if err != nil {
		p.serverErrorHandleFunc(w, r, auth.RemoteServerError)
		return
	}

	client := p.authConfig.Client(ctx, token)
	resp, err := client.Get(p.Config.UserEndpointURL)
	if err != nil {
		p.serverErrorHandleFunc(w, r, auth.RemoteServerError)
		return
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		p.serverErrorHandleFunc(w, r, auth.RemoteServerError)
		return

	}

	profile := p.userToProfileFunc(buf)

	var expiry int64
	if token.Expiry.IsZero() {
		expiry = time.Now().UnixNano() + defaultSessionLength.Nanoseconds()
	} else {
		expiry = token.Expiry.UnixNano()
	}

	// Create session. The access token may never expire, but the access might be revoked by
	// the user. If the user has revoked the access token disable the session. The session
	// check is done separately dependent on the provider.

	sessionID, err := p.sessions.CreateSession(token.AccessToken, expiry, profile)
	if err != nil {
		log.WithError(err).Errorf("Could not create session for user %+v", profile)
		p.serverErrorHandleFunc(w, r, auth.StoreError)
		return
	}

	cookie := &http.Cookie{
		Name:     auth.AuthCookieName,
		Value:    sessionID,
		HttpOnly: true,
		MaxAge:   0,
		Path:     "/",
		Secure:   p.Config.SecureCookie,
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, p.Config.LoginSuccessURL, http.StatusTemporaryRedirect)
}

func createAuthError(errorCode string) auth.AuthError {
	// TODO: Add all known errors
	switch errorCode {
	case "redirect_uri_mismatch":
		return auth.RedirectURIMismatchError
	default:
		return fmt.Errorf("Unknown error code received from provider: %s", errorCode)
	}
}
