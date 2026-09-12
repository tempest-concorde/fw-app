package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

// Sentinel errors returned by HandleCallback so callers can distinguish
// non-membership from upstream GitHub failures and missing configuration.
var (
	// ErrNotMember indicates the authenticated user is not a member of the
	// configured organization.
	ErrNotMember = errors.New("user is not a member of the configured organization")

	// ErrAuthFailed indicates GitHub rejected the code exchange or the user
	// lookup failed.
	ErrAuthFailed = errors.New("github authentication failed")

	// ErrConfig indicates the organization is not configured.
	ErrConfig = errors.New("github organization is not configured")
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
	// Fail closed if the organization is not configured.
	if g.ghOrg == "" {
		return nil, ErrConfig
	}

	// Exchange code for token
	token, err := g.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to exchange code: %w", ErrAuthFailed, err)
	}

	// Create GitHub client
	client := github.NewClient(g.oauthConfig.Client(ctx, token))

	// Get user info
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get user: %w", ErrAuthFailed, err)
	}

	// Check org membership
	isMember, _, err := client.Organizations.IsMember(ctx, g.ghOrg, user.GetLogin())
	if err != nil {
		return nil, fmt.Errorf("%w: failed to check org membership: %w", ErrAuthFailed, err)
	}
	if !isMember {
		return nil, fmt.Errorf("%w: user %s is not a member of organization %s",
			ErrNotMember, user.GetLogin(), g.ghOrg)
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
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failure is catastrophic - OAuth cannot proceed safely
		panic(fmt.Sprintf("crypto/rand.Read failed: %v", err))
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
