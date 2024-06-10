package lib

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/zmb3/spotify/v2"
)

type (
	poller struct {
		pollInterval   time.Duration
		client         *spotify.Client
		updateSendChan chan<- PlayStateUpdate

		lastID         string
		wasPlaying     bool
		lastProgressMS int
	}
	PlayStateUpdate struct {
		IsPlaying   bool
		AlbumName   string
		TrackName   string
		ArtistsName string
		CoverURL    string
		ProgressMS  int
	}
)

func NewPoller(intervalSec int, client *spotify.Client) (chan<- bool, <-chan PlayStateUpdate) {
	outChan := make(chan PlayStateUpdate, 1)
	endChan := make(chan bool)
	p := &poller{
		client:         client,
		pollInterval:   time.Second * time.Duration(intervalSec),
		updateSendChan: outChan,
	}
	go p.start(endChan)

	return endChan, outChan
}

func (p *poller) start(endChan <-chan bool) {
	go func() {
		for {
			select {
			case <-endChan:
				// stop polling
				fmt.Println("polling stopped")
				return

			case <-time.After(p.pollInterval):
				// poll new sstate from spotify
				p.tick()
			}
		}
	}()
}

func (p *poller) tick() {
	if p.client == nil {
		// client not set yet. wait for next tick
		return
	}

	playerState, err := p.client.PlayerState(context.Background())
	if err != nil {
		log.Printf("error getting player state: %s\n", err.Error())
		return
	}

	if !p.hasChanged(playerState) {
		log.Printf("nothing changed")
		return
	}

	if p.hasChanged(playerState) {
		select {
		case p.updateSendChan <- PlayStateUpdate{
			IsPlaying:   playerState.Playing,
			ProgressMS:  int(playerState.Progress),
			AlbumName:   getAlbumName(playerState),
			ArtistsName: getArtistsName(playerState),
			TrackName:   getTrackName(playerState),
			CoverURL:    GetCoverImageURL(playerState),
		}:
		case <-time.After(p.pollInterval / 2):
			fmt.Println("sending poll data timed out")
		}
	}
}

func (p *poller) hasChanged(playerState *spotify.PlayerState) bool {
	if playerState.Item != nil {
		if p.lastID != playerState.Item.ID.String() {
			return true
		}
	}

	if p.wasPlaying != playerState.Playing {
		return true
	}

	if p.lastProgressMS != int(playerState.Progress) {
		return true
	}

	return false
}

func getAlbumName(playerState *spotify.PlayerState) string {
	if playerState.Item == nil {
		return ""
	}

	// album is not a pointer so its safe to use
	return playerState.Item.Album.Name
}

func getArtistsName(playerState *spotify.PlayerState) string {
	if playerState.Item == nil {
		return ""
	}

	names := make([]string, len(playerState.Item.Artists))
	for i, spotifyArtist := range playerState.Item.Artists {
		names[i] = spotifyArtist.Name
	}

	return strings.Join(names, ", ")
}

func getTrackName(playerState *spotify.PlayerState) string {
	if playerState.Item == nil {
		return ""
	}

	return playerState.Item.Name
}

func GetCoverImageURL(playerState *spotify.PlayerState) string {
	if playerState.Item == nil {
		return ""
	}

	if len(playerState.Item.Album.Images) > 0 {
		return playerState.Item.Album.Images[0].URL
	}

	return ""
}
