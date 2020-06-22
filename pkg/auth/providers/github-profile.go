package providers

import (
	"encoding/json"
	"strconv"

	"github.com/eesrc/geo/pkg/auth"
)

// GithubUserProfile contains the GitHub profile. It's much bigger,
// but only the fields used are documented for now.
type GithubUserProfile struct {
	ID        int64  `json:"id"`
	AvatarURL string `json:"avatarUrl"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Login     string `json:"login"`
}

// profileFromGithubUser reads the profile from json bytes returned from the Github user endpoint
func profileFromGithubUser(jsonBytes []byte) auth.Profile {
	var githubProfile GithubUserProfile
	profile := auth.Profile{}

	err := json.Unmarshal(jsonBytes, &githubProfile)

	if err != nil {
		return profile
	}

	profile.LoginID = strconv.FormatInt(githubProfile.ID, 10)

	profile.Name = githubProfile.Name

	profile.Email = githubProfile.Email
	// GitHub requires you to verify your mail so we assume the email is verified
	profile.EmailVerified = true

	profile.AvatarURL = githubProfile.AvatarURL
	profile.UserObject = githubProfile
	profile.Provider = auth.AuthGithub

	return profile
}
