package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/eesrc/geo/pkg/auth"
)

type githubSessionCheckPayload struct {
	AccessToken string `json:"access_token"`
}

func githubSessionChecker(sessions auth.SessionStore, config Config) {
	client := http.Client{}
	for {
		timeToCheck := time.Now().UnixNano() + int64(sessionCheckInterval)
		activeSessions, err := sessions.GetSessions(timeToCheck)

		if err != nil {
			log.Errorf("Got error checking for expired sessions: %v", err)
			continue
		}
		for _, session := range activeSessions {
			// Only check github sessions
			if session.Profile.Provider != auth.AuthGithub {
				continue
			}

			// GitHub token check endpoint wants a payload with the token
			githubSession := githubSessionCheckPayload{AccessToken: session.AccessToken}
			payloadBytes, _ := json.Marshal(githubSession)

			// Check profile status
			checkURL := fmt.Sprintf(githubTokenCheckURL, config.ClientID)
			req, err := http.NewRequest("POST", checkURL, bytes.NewBuffer(payloadBytes))
			if err != nil {
				continue
			}
			req.SetBasicAuth(config.ClientID, config.ClientSecret)
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

			if resp.StatusCode == http.StatusNotFound {
				log.Info("Deauthorizing session due to not found in github app")
				if err := sessions.RemoveSession(session.ID); err != nil {
					log.Errorf("Unable to remove session %v: %v", session.ID, err)
				}

				continue
			}

			if resp.StatusCode != http.StatusOK {
				log.Errorf("Got an error when trying to check the github session. %v. Deauthorizing github session just to be safe.", resp.Status)
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
