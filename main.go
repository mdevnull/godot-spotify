package main

import (
	"context"
	"fmt"
	"main/lib"

	"github.com/zmb3/spotify/v2"
	"grow.graphics/gd"
	"grow.graphics/gd/gdextension"
)

type GodotSpotify struct {
	gd.Class[GodotSpotify, gd.Node]

	SpotifyClientID     gd.String `gd:"spotify_client_id"`
	SpotifyClientSecret gd.String `gd:"spotify_client_secret"`
	SpotifyAuthURL      gd.String `gd:"spotify_auth_url"`
	PollInterval        gd.Int    `default:"5"`
	Poll                gd.Bool

	IsPlaying   gd.Bool   `gd:"is_playing"`
	AlbumName   gd.String `gd:"album_name"`
	TrackName   gd.String `gd:"track_name"`
	ArtistsName gd.String `gd:"artist_names"`
	CoverURL    gd.String `gd:"cover_url"`
	ProgressMS  gd.Int    `gd:"progress_ms"`

	running    bool
	updateChan <-chan lib.PlayStateUpdate
	endChan    chan<- bool
}

// Ready implements the Godot Node2D _ready interface (virtual function).
func (h *GodotSpotify) Ready(godoCtx gd.Context) {
	clientID := h.SpotifyClientID.String()
	clientSecret := h.SpotifyClientSecret.String()

	if clientID == "" || clientSecret == "" {
		godoCtx.Printerr(godoCtx.Variant("Missing spotify client id or secret"))
		return
	}

	auth := lib.MakeAuth(clientID, clientSecret)
	h.SpotifyAuthURL = h.Pin().String(lib.GetAuthURL(auth))

	h.AlbumName = h.Pin().String("")
	h.TrackName = h.Pin().String("")
	h.ArtistsName = h.Pin().String("")
	h.CoverURL = h.Pin().String("")

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

func (h *GodotSpotify) Process(godoCtx gd.Context, delta gd.Float) {
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

		h.IsPlaying = gd.Bool(updateMsg.IsPlaying)

		h.AlbumName.Free()
		h.AlbumName = h.Pin().String(updateMsg.AlbumName)

		h.TrackName.Free()
		h.TrackName = h.Pin().String(updateMsg.TrackName)

		h.ArtistsName.Free()
		h.ArtistsName = h.Pin().String(updateMsg.ArtistsName)

		h.CoverURL.Free()
		h.CoverURL = h.Pin().String(updateMsg.CoverURL)

		h.ProgressMS = gd.Int(updateMsg.ProgressMS)
	}
}

func (h *GodotSpotify) OnSet(godoCtx gd.Context, propName gd.StringName, propValue gd.Variant) {
	fmt.Printf("Godot Spotify Prop update: %s\n", propName.String())
}

func (h *GodotSpotify) ToString(godoCtx gd.Context) gd.String {
	return godoCtx.String("GodotSpotify")
}

func main() {
	godot, ok := gdextension.Link()
	if !ok {
		panic("Unable to link to godot")
	}
	gd.Register[GodotSpotify](godot)
}
