package providers

import (
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/auth"
)

func getConnectTokenCheckURL(config Config) string {
	if strings.Contains(config.AuthEndpointURL, "staging") {
		return "https://connect.staging.telenordigital.com/oauth/tokeninfo"
	}

	return "https://connect.telenordigital.com/oauth/tokeninfo"
}

func connectSessionChecker(sessions auth.SessionStore, config Config) {
	client := http.Client{}
	for {
		timeToCheck := time.Now().UnixNano() + int64(sessionCheckInterval)
		activeSessions, err := sessions.GetSessions(timeToCheck)

		if err != nil {
			log.Errorf("Got error checking for expired sessions: %v", err)
			continue
		}
		for _, session := range activeSessions {
			// Only check Connect sessions
			if session.Profile.Provider != auth.AuthConnect {
				continue
			}

			// Check token status
			req, err := http.NewRequest("GET", getConnectTokenCheckURL(config), nil)
			req.URL.Query().Add("access_token", session.AccessToken)
			if err != nil {
				continue
			}
			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			// Read the entire response, then discard. This ensures the http.Client is
			// reused properly.
			_, err = io.Copy(ioutil.Discard, resp.Body)

			if err != nil {
				log.Error("Failed to discard http client response. This may lead to sessions not being removed.")
			}

			resp.Body.Close()

			// If the token is expired the Connect auth server will return 400. If so we remove the session.
			// TODO: Add refresh token and check if refresh is possible.
			if resp.StatusCode == http.StatusBadRequest {
				log.Info("Deauthorizing session due to not found in connect app")

				if err := sessions.RemoveSession(session.ID); err != nil {
					log.Errorf("Unable to remove session %v: %v", session.ID, err)
				}

				continue
			}

			if err := sessions.RefreshSession(session.ID, int64(defaultSessionLength)); err != nil {
				log.Errorf("Unable to update session %v: %v", session.ID, err)
			}
		}
		time.Sleep(sessionCheckInterval)
	}
}
