package main

import (
	"context"
	"encoding/json"
	"github.com/bclindner/iasipgenerator/iasipgen"
	"github.com/microcosm-cc/bluemonday"
	"html"
	"github.com/mattn/go-mastodon"
	"image/jpeg"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"fmt"
)

// iasipbot should only respond to a message if it and only it is directly mentioned before any non-mention text.
// examples:
// - if a user tags 5 accounts at the very start of the message and
// iasipbot is one of them, then it will trigger the generator
// - if a user tags X accounts but iasipbot is mentioned after any
// initial tags, then iasipbot will not trigger (making the assumption
// that the bot is merely being discussed)
//
// in proper tryhard fashion, i wrote a dumb and probably naive regex
// to do all of that, because why not
//
// first group matches mentions, second group is actual text to make an IASIP for
var (
	statusRegexp = regexp.MustCompile(`(?ms)^((?:(?:@[\w\-]+(?:@[\w\-\.]+)?) ?)+)(.+)$`)
	striptags = bluemonday.StrictPolicy()
)

// Config is the base schema for this app's JSON configuration file.
type Config struct {
	FontPath    string         `json:"fontpath"`
	Credentials MastodonConfig `json:"credentials"`
}

// MastodonConfig is identical to mastodon.Config in structure but includes
// JSON directives and a function to instantiate a client directly from it.
type MastodonConfig struct {
	Server       string `json:"server"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	AccessToken  string `json:"accessToken"`
}

func (cfg MastodonConfig) GetClient() *mastodon.Client {
	return mastodon.NewClient(&mastodon.Config{
		Server:       cfg.Server,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		AccessToken:  cfg.AccessToken,
	})
}

func main() {
	// read config from file
	cfgFile, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Printf("Failed to read config file at config.json. Is it unavailable?\nFull error: %s\n", err)
		os.Exit(1)
	}
	// parse config
	var cfg Config
	err = json.Unmarshal(cfgFile, &cfg)
	if err != nil {
		fmt.Printf("Failed to parse config file at %s.\nFull error: %s\n", "./config.json", err)
		os.Exit(1)
	}
	// load font into iasipgen
	err = iasipgen.LoadFont(cfg.FontPath)
	if err != nil {
		fmt.Printf("Failed to parse font at %s.\nFull error: %s\n", cfg.FontPath, err)
		os.Exit(1)
	}
	fmt.Printf("Found font at %s\n", cfg.FontPath)
	// create mastodon client, establish context, and get info about the bot user
	client := cfg.Credentials.GetClient()
	ctx, ctxCancel := context.WithCancel(context.Background())
	self, err := client.GetAccountCurrentUser(ctx)
	if err != nil {
		fmt.Printf("Failed to get account information: %s\n", err)
		os.Exit(1)
	}
	// establish websocket connection
	wsclient := client.NewWSClient()
	evtStream, err := wsclient.StreamingWSUser(ctx)
	if err != nil {
		fmt.Printf("Failed to establish websocket connection.\nFull error:%s\n", err)
		os.Exit(1)
	}
	// enter event loop
	fmt.Println("Entering event loop.")
	for genericEvent := range evtStream {
		switch evt := genericEvent.(type) {
		// notification events contain the mention events we want to get
		case *mastodon.NotificationEvent:
			// only get mentions - skip all others
			switch evt.Notification.Type {
			case "mention":
				status := evt.Notification.Status
				// parse content to a human- (and in this case bot-) readable string
				content := html.UnescapeString(striptags.Sanitize(status.Content))
				// try to match content using regexp
				matches := statusRegexp.FindStringSubmatch(content)
				// a nil match means this message is invalid
				if matches == nil {
					break
				}
				mentions := strings.TrimSpace(matches[1])
				message := strings.TrimSpace(matches[2])
				// if this account is not the only mention, skip the message
				if "@" + self.Username != mentions {
					break
				}
				// try to make an IASIP
				img, err := iasipgen.Generate(message)
				if err != nil {
					fmt.Printf("Failed to make IASIP for message by %s:\n\n%s\n\nError: %s\n", status.Account.Acct, content, err)
					break
				}
				/*
					this next block is a little silly - i have to make a temp file and
					write the JPEG to it since there is no way to upload media directly
					from an io.Reader interface or similar in go-mastodon. maybe I should
					raise a github issue about that.
				*/
				// create temp file and write the IASIP to it as a jpeg
				tmpfile, err := ioutil.TempFile("", "iasipbot_*.jpg")
				if err != nil {
					fmt.Printf("Failed to create temp file for message by %s:\n\n%s\n\nError: %s\n", status.Account.Acct, content, err)
					break
				}
				defer tmpfile.Close()
				// encode the jpeg to the tempfile
				err = jpeg.Encode(tmpfile, img, &jpeg.Options{
					Quality: 100,
				})
				if err != nil {
					fmt.Printf("Failed to encode image for message by %s:\n\n%s\n\nError: %s\n", status.Account.Acct, content, err)
					break
				}
				// upload the temp file
				attach, err := client.UploadMedia(ctx, tmpfile.Name())
				if err != nil {
					fmt.Printf("Failed to upload image for message by %s:\n\n%s\n\nError: %s\n", status.Account.Acct, content, err)
					break
				}
				_, err = client.PostStatus(ctx, &mastodon.Toot{
					InReplyToID: status.ID,
					Sensitive:   true,
					SpoilerText: "bot-generated IASIP",
					Status: "@"+status.Account.Acct,
					MediaIDs:    []mastodon.ID{attach.ID},
					// match the post's visibility (i.e. if the bot was DM'd then DM them back, if the toot is public then also make this public)
					Visibility: status.Visibility,
				})
				if err != nil {
					fmt.Printf("Failed to post toot for message by %s:\n\n%s\n\nError: %s\n", status.Account.Acct, content, err)
					break
				}
			default:
				break
			}
		// catch any errors the websocket gives us
		case *mastodon.ErrorEvent:
			fmt.Printf("Error in websocket event loop: %s", evt.Error())
			ctxCancel()
			break
		// ignore all others
		default:
			continue
		}
	}
}
