package auth

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

// GitHubUser represents an authenticated GitHub user
type GitHubUser struct {
	ID    int64
	Login string
	Email string
}

// GitHubAuth handles GitHub OAuth authentication
type GitHubAuth struct {
	oauthConfig *oauth2.Config
	ghOrg       string
}

// NewGitHubAuth creates a new GitHub OAuth authenticator
func NewGitHubAuth(clientID, clientSecret, org string) *GitHubAuth {
	return &GitHubAuth{
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  "", // Empty - uses callback URL registered in GitHub App settings
			Scopes:       []string{"read:org", "user:email"},
			Endpoint:     githuboauth.Endpoint,
		},
		ghOrg: org,
	}
}

// StartLogin initiates the OAuth flow
func (g *GitHubAuth) StartLogin() (state, redirectURL string) {
	// Generate random state parameter
	state = generateRandomState()
	redirectURL = g.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
	return state, redirectURL
}

// HandleCallback exchanges the OAuth code for a token and validates org membership
func (g *GitHubAuth) HandleCallback(ctx context.Context, code string) (*GitHubUser, error) {
	// Exchange code for token
	token, err := g.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Create GitHub client
	client := github.NewClient(g.oauthConfig.Client(ctx, token))

	// Get user info
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check org membership
	isMember, _, err := client.Organizations.IsMember(ctx, g.ghOrg, user.GetLogin())
	if err != nil {
		return nil, fmt.Errorf("failed to check org membership: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("user %s is not a member of organization %s", user.GetLogin(), g.ghOrg)
	}

	// Get user email
	var email string
	emails, _, err := client.Users.ListEmails(ctx, nil)
	if err == nil && len(emails) > 0 {
		for _, e := range emails {
			if e.GetPrimary() {
				email = e.GetEmail()
				break
			}
		}
		if email == "" {
			email = emails[0].GetEmail()
		}
	}

	return &GitHubUser{
		ID:    user.GetID(),
		Login: user.GetLogin(),
		Email: email,
	}, nil
}

func generateRandomState() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
