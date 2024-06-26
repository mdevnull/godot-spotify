package lib

import (
	"fmt"
	"net/http"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var (
	state = "abc123"
)

func MakeAuth(clientID string, clientSecret string) *spotifyauth.Authenticator {
	return spotifyauth.New(
		spotifyauth.WithRedirectURL("http://localhost:8188"),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate),
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
		spotifyauth.WithScopes("user-read-playback-state"),
	)
}

func WebServer(auth *spotifyauth.Authenticator, setClientChan chan<- *spotify.Client) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Got a request in")

		token, err := auth.Token(r.Context(), state, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusNotFound)
			return
		}

		// create a client using the specified token
		client := spotify.New(auth.Client(r.Context(), token))
		setClientChan <- client
	})

	server := &http.Server{Addr: ":8188", Handler: mux}
	go server.ListenAndServe()
	fmt.Println("webserver should be up")

	return server
}

func GetAuthURL(auth *spotifyauth.Authenticator) string {
	authURL := auth.AuthURL(state)
	fmt.Println(authURL)
	return authURL
}
