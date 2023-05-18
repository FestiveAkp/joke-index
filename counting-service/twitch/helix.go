package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/FestiveAkp/jji/counting-service/utils"
	"github.com/nicklaw5/helix/v2"
)

func GetNewHelixClient() *helix.Client {
	client, err := helix.NewClient(&helix.Options{
		ClientID:     "qj3a9kuig1ir6djwh53effyhqau1x3",
		ClientSecret: "8sd5sg0d31cz7b9g288b6encatfxrk",
	})
	utils.Check(err)

	resp, err := client.RequestAppAccessToken([]string{"user:read:email"})
	utils.Check(err)
	client.SetAppAccessToken(resp.Data.AccessToken)

	return client
}

func GetUserIDByChannelName(client *helix.Client, channel string) string {
	resp, err := client.GetUsers(&helix.UsersParams{
		Logins: []string{channel},
	})
	utils.Check(err)

	return resp.Data.Users[0].ID
}

func IsStreamLive(client *helix.Client, channel string) bool {
	resp, err := client.GetStreams(&helix.StreamsParams{
		UserLogins: []string{channel},
	})
	utils.Check(err)

	// fmt.Printf("%+v\n", resp)
	return len(resp.Data.Streams) > 0
}

func GetEventSubSubscriptions(client *helix.Client) []helix.EventSubSubscription {
	resp, err := client.GetEventSubSubscriptions(&helix.EventSubSubscriptionsParams{
		Status: helix.EventSubStatusEnabled,
	})
	utils.Check(err)

	fmt.Printf("%+v\n", resp)
	// fmt.Printf("%+v\n", resp.Data.EventSubSubscriptions)
	return resp.Data.EventSubSubscriptions
}

func CreateEventSubSubscriptionStreamOnline(client *helix.Client, userID string) *helix.EventSubSubscriptionsResponse {
	resp, err := client.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    helix.EventSubTypeStreamOnline,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: userID,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: "https://localhost:443/webhooks/twitch-callback",
			Secret:   "secretword",
		},
	})
	utils.Check(err)
	return resp
}

func CreateEventSubSubscriptionStreamOffline(client *helix.Client, userID string) *helix.EventSubSubscriptionsResponse {
	resp, err := client.CreateEventSubSubscription(&helix.EventSubSubscription{
		Type:    helix.EventSubTypeStreamOffline,
		Version: "1",
		Condition: helix.EventSubCondition{
			BroadcasterUserID: userID,
		},
		Transport: helix.EventSubTransport{
			Method:   "webhook",
			Callback: "https://localhost:443/webhooks/twitch-callback",
			Secret:   "secretword",
		},
	})
	utils.Check(err)
	return resp
}

type eventSubNotification struct {
	Subscription helix.EventSubSubscription `json:"subscription"`
	Challenge    string                     `json:"challenge"`
	Event        json.RawMessage            `json:"event"`
}

func HandleEventSubOnline(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	defer r.Body.Close()

	// Verify that the notification came from Helix using the secret
	if !helix.VerifyEventSubNotification("secretword", r.Header, string(body)) {
		log.Println("No valid signature on EventSub subscription, rejecting")
		return
	} else {
		log.Println("Verified signature for EventSub subscription")
	}

	var vals eventSubNotification

	err = json.NewDecoder(bytes.NewReader(body)).Decode(&vals)
	if err != nil {
		log.Println(err)
		return
	}

	// If there's a challenge in the request, respond with only the challenge to verify the EventSub
	if vals.Challenge != "" {
		w.Write([]byte(vals.Challenge))
		return
	}

	var onlineEvent helix.EventSubStreamOnlineEvent
	err = json.NewDecoder(bytes.NewReader(vals.Event)).Decode(&onlineEvent)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Got stream online webhook: %s %s", onlineEvent.Type, onlineEvent.BroadcasterUserLogin)
	w.WriteHeader(200)
	w.Write([]byte("ok"))
}

// ID:90acb2fc-8fee-41c7-997f-498ced0f4647
