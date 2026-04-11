package rtc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	lksdk "github.com/livekit/server-sdk-go/v2"
)

// RelayMsg is the JSON protocol between speaker and listener over LiveKit Data Channel.
type RelayMsg struct {
	Type   string `json:"type"`             // "interim" or "pair"
	Text   string `json:"text,omitempty"`   // interim only
	Source string `json:"source,omitempty"` // pair only
	Target string `json:"target,omitempty"` // pair only
}

// Client wraps a LiveKit room connection for real-time data relay.
type Client struct {
	room *lksdk.Room
	ch   chan RelayMsg
}

// Connect joins a LiveKit room and returns a Client.
// The caller should defer Client.Close().
func Connect(ctx context.Context, url, apiKey, apiSecret, roomName, identity string) (*Client, error) {
	c := &Client{
		ch: make(chan RelayMsg, 64),
	}

	cb := &lksdk.RoomCallback{
		ParticipantCallback: lksdk.ParticipantCallback{
			OnDataPacket: func(data lksdk.DataPacket, params lksdk.DataReceiveParams) {
				pkt, ok := data.(*lksdk.UserDataPacket)
				if !ok {
					return
				}
				var msg RelayMsg
				if err := json.Unmarshal(pkt.Payload, &msg); err != nil {
					return
				}
				select {
				case c.ch <- msg:
				default:
				}
			},
		},
	}

	room, err := lksdk.ConnectToRoom(url, lksdk.ConnectInfo{
		APIKey:              apiKey,
		APISecret:           apiSecret,
		RoomName:            roomName,
		ParticipantIdentity: identity,
	}, cb)
	if err != nil {
		return nil, fmt.Errorf("livekit connect: %w", err)
	}

	c.room = room
	log.Printf("Connected to LiveKit room %q as %q", roomName, identity)
	return c, nil
}

// Send publishes a RelayMsg to all participants in the room (reliable delivery).
func (c *Client) Send(msg RelayMsg) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.room.LocalParticipant.PublishDataPacket(
		&lksdk.UserDataPacket{Payload: payload},
		lksdk.WithDataPublishReliable(true),
	)
}

// Messages returns a channel that receives RelayMsg from other participants.
func (c *Client) Messages() <-chan RelayMsg {
	return c.ch
}

// Close disconnects from the LiveKit room.
func (c *Client) Close() {
	if c.room != nil {
		c.room.Disconnect()
	}
}
