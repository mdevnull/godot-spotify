package main

import (
	"context"
	"fmt"
	"main/lib"
	"os/exec"
	"strings"

	"github.com/zmb3/spotify/v2"
	"graphics.gd/classdb"
	"graphics.gd/classdb/Engine"
	"graphics.gd/classdb/Node"
	"graphics.gd/classdb/OS"
	"graphics.gd/startup"
	"graphics.gd/variant/Float"
)

type GodotSpotify struct {
	classdb.Extension[GodotSpotify, Node.Instance]

	SpotifyClientID     string `gd:"spotify_client_id"`
	SpotifyClientSecret string `gd:"spotify_client_secret"`
	SpotifyAuthURL      string `gd:"spotify_auth_url"`
	PollInterval        int    `default:"5"`
	Poll                bool

	IsPlaying   bool   `gd:"is_playing"`
	AlbumName   string `gd:"album_name"`
	TrackName   string `gd:"track_name"`
	ArtistsName string `gd:"artist_names"`
	CoverURL    string `gd:"cover_url"`
	ProgressMS  int    `gd:"progress_ms"`
	LengthMS    int    `gd:"length_ms"`

	running    bool
	updateChan <-chan lib.PlayStateUpdate
	endChan    chan<- bool
}

// Ready implements the Godot Node2D _ready interface (virtual function).
func (h *GodotSpotify) Ready() {
	clientID := h.SpotifyClientID
	clientSecret := h.SpotifyClientSecret

	if clientID == "" || clientSecret == "" {
		Engine.Log("Missing spotify client id or secret")
		return
	}

	auth := lib.MakeAuth(clientID, clientSecret)
	h.SpotifyAuthURL = lib.GetAuthURL(auth)

	h.AlbumName = ""
	h.TrackName = ""
	h.ArtistsName = ""
	h.CoverURL = ""

	clientChan := make(chan *spotify.Client)
	webServer := lib.WebServer(auth, clientChan)
	go func() {
		client := <-clientChan
		webServer.Shutdown(context.Background())
		h.endChan, h.updateChan = lib.NewPoller(int(h.PollInterval), client)
		h.running = true
		h.Poll = true
	}()
}

func (h *GodotSpotify) Process(delta Float.X) {
	if h.running && !bool(h.Poll) && h.endChan != nil {
		h.endChan <- true
		h.endChan = nil
		h.running = false
	}

	if !h.running && bool(h.Poll) {
		h.Poll = false
	}

	// process max one queued msg
	if len(h.updateChan) > 0 {
		updateMsg := <-h.updateChan

		h.IsPlaying = updateMsg.IsPlaying
		h.ProgressMS = updateMsg.ProgressMS
		h.LengthMS = updateMsg.TrackLengthMS

		if h.AlbumName != updateMsg.AlbumName {
			h.AlbumName = updateMsg.AlbumName
		}

		if h.TrackName != updateMsg.TrackName {
			h.TrackName = updateMsg.TrackName
		}

		if h.ArtistsName != updateMsg.ArtistsName {
			h.ArtistsName = updateMsg.ArtistsName
		}

		if h.CoverURL != updateMsg.CoverURL {
			h.CoverURL = updateMsg.CoverURL
		}
	}
}

func (h *GodotSpotify) OnSet(propName string, propValue any) {
	fmt.Printf("Godot Spotify Prop update: %s\n", propName)
}

func (h *GodotSpotify) ToString() string {
	return "GodotSpotify"
}

func (h *GodotSpotify) OpenAuthInBrowser() {
	var openCmd *exec.Cmd
	switch strings.ToLower(OS.GetName()) {
	case "windows":
		winQuotedURL := strings.ReplaceAll(h.SpotifyAuthURL, "&", "^&")
		openCmd = exec.Command("cmd", "/c", "start", winQuotedURL)
	case "macos":
		openCmd = exec.Command("open", h.SpotifyAuthURL)
	case "linux":
		openCmd = exec.Command("xdg-open", h.SpotifyAuthURL)
	default:
		Engine.Log("unable to open browser on current platform")
		return
	}

	if err := openCmd.Run(); err != nil {
		Engine.Log("error opening browser for auth: %s", err.Error())
	}
}

func main() {
	classdb.Register[GodotSpotify]()
	startup.Engine()
}
