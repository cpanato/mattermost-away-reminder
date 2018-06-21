package gcalendar

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"

	"github.com/cpanato/mattermost-away-reminder/model"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	fmt.Println(err)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}

	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	fileName := FindClientAndTokenFile(file)
	fmt.Println("Loading " + fileName)
	f, err := os.Open(fileName)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	json.NewEncoder(f).Encode(token)
}

func FindClientAndTokenFile(fileName string) string {
	if _, err := os.Stat("/tmp/" + fileName); err == nil {
		fileName, _ = filepath.Abs("/tmp/" + fileName)
	} else if _, err := os.Stat("./config/" + fileName); err == nil {
		fileName, _ = filepath.Abs("./config/" + fileName)
	} else if _, err := os.Stat("../config/" + fileName); err == nil {
		fileName, _ = filepath.Abs("../config/" + fileName)
	} else if _, err := os.Stat("./client/" + fileName); err == nil {
		fileName, _ = filepath.Abs("./client/" + fileName)
	} else if _, err := os.Stat("./token/" + fileName); err == nil {
		fileName, _ = filepath.Abs("./token/" + fileName)
	} else if _, err := os.Stat(fileName); err == nil {
		fileName, _ = filepath.Abs(fileName)
	}

	return fileName
}

func getGClient() *calendar.Service {
	fileName := FindClientAndTokenFile("client_secret.json")
	fmt.Println("Loading " + fileName)

	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved client_secret.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	srv, err := calendar.New(getClient(config))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	return srv
}

func AddEventToGCal(text, fromDate, toDate, calendarId string) (string, error) {

	srv := getGClient()
	fmt.Println("Adding one Event")

	event := &calendar.Event{
		Summary:     text,
		Description: text,
		Start: &calendar.EventDateTime{
			Date: fromDate,
		},
		End: &calendar.EventDateTime{
			Date: toDate,
		},
	}

	event, err := srv.Events.Insert(calendarId, event).Do()
	if err != nil {
		log.Fatalf("Unable to create event. %v\n", err)
	}
	fmt.Printf("Event created: %s\n", event.Id)

	return event.Id, nil

}

func RemoveEventFromGCal(id, calendarId string) error {
	srv := getGClient()
	fmt.Println("Removing Event")

	event := srv.Events.Delete(calendarId, id).Do()
	if event != nil {
		msg := fmt.Sprintf("Unable to delete event. %v", event)
		fmt.Println(msg)
		return model.NewLocAppError("Mattermost Time Away", msg, nil, "")
	}
	fmt.Printf("Event Deleted\n")
	return nil
}
