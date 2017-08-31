package main

import (
	"encoding/csv"
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

	"github.com/BurntSushi/toml"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

const FORMAT_YYMMDD = "2006-01-02"
const FORMAT_YYMMDD_HHMISS = "2006-01-02 15:04:05"

////////////////////////////////////////////////////////////////////////////////

type Config struct {
	Date   DateConfig
	Output OutputConfig
	GCal   GCalConfig
}

type DateConfig struct {
	Start string
	End   string
}

type OutputConfig struct {
	Filename string
	Type     string
}

type GCalConfig struct {
	ClientSecret string `toml:"client_secret"`
	CalendarId   string `toml:"calender_id"`
}

type WorkLog struct {
	start   time.Time
	end     time.Time
	summary string
}

func newWorkLog(start time.Time, end time.Time, summary string) *WorkLog {
	w := new(WorkLog)
	w.start = start
	w.end = end
	w.summary = summary
	return w
}

func (w WorkLog) toRecord() []string {
	startTime := w.start.Format(FORMAT_YYMMDD_HHMISS)
	endTime := w.end.Format(FORMAT_YYMMDD_HHMISS)
	workHour := w.end.Sub(w.start).Hours()
	return []string{startTime, endTime, w.summary, fmt.Sprint(workHour)}
}

////////////////////////////////////////////////////////////////////////////////

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

////////////////////////////////////////////////////////////////////////////////

// Google Calendar APIから取得した情報を、TSV形式(SJIS)で書き出し
func dump2tsv(events *calendar.Events, outfile string) {
	if len(events.Items) < 1 {
		fmt.Printf("No upcoming events found.\n")
		return
	}

	file, err := os.OpenFile(outfile, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatalf("Unable to open outfile: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(transform.NewWriter(file, japanese.ShiftJIS.NewEncoder()))
	writer.UseCRLF = true
	writer.Comma = '\t'

	header := []string{"開始時刻", "終了時刻", "作業概要", "作業時間"}
	writer.Write(header)
	for _, i := range events.Items {
		// 「終日」になっているカレンダーの予定は、出力対象外
		if i.Start.DateTime == "" {
			continue
		}

		st, _ := time.Parse(time.RFC3339, i.Start.DateTime)
		ed, _ := time.Parse(time.RFC3339, i.End.DateTime)
		worklog := newWorkLog(st, ed, i.Summary)

		record := worklog.toRecord()

		fmt.Printf("%v\n", record)
		writer.Write(worklog.toRecord())
	}

	writer.Flush()
}

////////////////////////////////////////////////////////////////////////////////

// @see https://developers.google.com/google-apps/calendar/quickstart/go
func main() {
	// config.tomlから各種情報を読み込み
	var conf Config
	_, err := toml.DecodeFile("./config.toml", &conf)
	if err != nil {
		log.Fatalf("Unable to read config.toml: %v", err)
		panic(err)
	}

	startDate := conf.Date.Start
	endDate := conf.Date.End

	outfile := conf.Output.Filename

	client_secret := conf.GCal.ClientSecret
	calender_id := conf.GCal.CalendarId
	scope := calendar.CalendarReadonlyScope

	// credential(client_secret.json)読み込み
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

	// 開始日、終了日を指定してGoogle Calendarから取得
	fmt.Printf("start=%s, end=%s \n", startDate, endDate)
	st, _ := time.Parse(FORMAT_YYMMDD, startDate)
	ed, _ := time.Parse(FORMAT_YYMMDD, endDate)
	events, err := srv.Events.List(calender_id).
		TimeMin(st.Format(time.RFC3339)).
		TimeMax(ed.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events. %v", err)
	}

	dump2tsv(events, outfile)
}
