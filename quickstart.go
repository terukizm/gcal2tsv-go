package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

const FORMAT_YYMMDD = "2006-01-02"
const FORMAT_YYMMDD_HHMISS = "2006-01-02 15:04:05"

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("calendar-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// dump2tsv
func dump2tsv(events *calendar.Events) {
	if len(events.Items) < 1 {
		fmt.Printf("No upcoming events found.\n")
	} else {
		for _, i := range events.Items {
			// 「終日」の予定は、TSV出力の対象外とする
			if i.Start.DateTime == "" {
				continue
			}

			summary := i.Summary
			st, _ := time.Parse(time.RFC3339, i.Start.DateTime)
			ed, _ := time.Parse(time.RFC3339, i.End.DateTime)
			startTime := st.Format(FORMAT_YYMMDD_HHMISS)
			endTime := ed.Format(FORMAT_YYMMDD_HHMISS)

			fmt.Printf("%s (%s - %s)\n", summary, startTime, endTime)
		}
	}
}

// @see https://developers.google.com/google-apps//alendar/quickstart/go
func main() {

	client_secret := "./client_secret.json"
	calender_id := "uik1nf72sm3t6vtmnu75k1hni8@group.calendar.google.com"
	scope := calendar.CalendarReadonlyScope

	// credential(client_secret)読み込み
	b, err := ioutil.ReadFile(client_secret)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/calendar-go-quickstart.json
	config, err := google.ConfigFromJSON(b, scope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	ctx := context.Background()
	client := getClient(ctx, config)

	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve calendar Client %v", err)
	}

	st, _ := time.Parse(FORMAT_YYMMDD, "2017-08-01")
	ed, _ := time.Parse(FORMAT_YYMMDD, "2017-08-31")
	fmt.Printf("start=%v, end=%v", st, ed)
	events, err := srv.Events.List(calender_id).
		TimeMin(st.Format(time.RFC3339)).
		TimeMax(ed.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events. %v", err)
	}

	fmt.Println("Upcoming events:")
	dump2tsv(events)
}