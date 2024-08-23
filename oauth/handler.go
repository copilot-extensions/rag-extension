package oauth

import (
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"net/http"
)

// Service provides endpoints to allow this agent to be authorized.
type Service struct {
	conf *oauth2.Config
}

func NewService(clientID, clientSecret, callback string) *Service {
	return &Service{
		conf: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  callback,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},
		},
	}
}

const (
	STATE_COOKIE = "oauth_state"
)

// PreAuth is the landing page that the user arrives at when they first attempt
// to use the agent while unauthorized.  You can do anything you want here,
// including making sure the user has an account on your side.  At some point,
// you'll probably want to make a call to the authorize endpoint to authorize
// the app.
func (s *Service) PreAuth(w http.ResponseWriter, r *http.Request) {
	// In our example, we're not doing anything except going through the
	// authorization flow.  This is standard Oauth2.

	verifier := oauth2.GenerateVerifier()
	state := uuid.New()

	url := s.conf.AuthCodeURL(state.String(), oauth2.AccessTypeOnline, oauth2.S256ChallengeOption(verifier))
	stateCookie := &http.Cookie{
		Name:     STATE_COOKIE,
		Value:    state.String(),
		MaxAge:   10 * 60, // 10 minutes in seconds
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, stateCookie)
	w.Header().Set("location", url)
	w.WriteHeader(http.StatusFound)
}

// PostAuth is the landing page where the user lads after authorizing.  As
// above, you can do anything you want here.  A common thing you might do is
// get the user information and then perform some sort of account linking in
// your database.
func (s *Service) PostAuth(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	stateCookie, err := r.Cookie(STATE_COOKIE)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("state cookie not found"))
		return
	}

	// Important:  Compare the state!  This prevents CSRF attacks
	if state != stateCookie.Value {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid state"))
		return
	}

	_, err = s.conf.Exchange(r.Context(), code)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("error exchange code for token: %v", err)))
		return
	}

	// Response contains an access token, now the world is your oyster.  Get user information and perform account linking, or do whatever you want from here.

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("All done!  Please return to the app"))
}
