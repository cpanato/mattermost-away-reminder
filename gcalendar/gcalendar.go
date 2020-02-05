package gcalendar

import (
	"context"
	"fmt"
	"log"

	"github.com/cpanato/mattermost-away-reminder/model"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func AddEventToGCal(userName, text, fromDate, toDate, calendarId, calendarAPIKey string) (string, error) {

	calendarService, err := calendar.NewService(context.Background(), option.WithAPIKey(calendarAPIKey))
	if err != nil {
		log.Fatalf("Unable to create Calendar service: %v", err)
	}

	fmt.Println("Adding one Event")

	msg := fmt.Sprintf("%v: %v", userName, text)
	eventToSave := &calendar.Event{
		Summary:     msg,
		Description: msg,
		Start: &calendar.EventDateTime{
			Date: fromDate,
		},
		End: &calendar.EventDateTime{
			Date: toDate,
		},
	}

	event, err := calendarService.Events.Insert(calendarId, eventToSave).Do()
	if err != nil {
		log.Fatalf("Unable to create event. %v\n", err)
	}
	fmt.Printf("Event created: %s\n", event.Id)

	return event.Id, nil

}

func RemoveEventFromGCal(id, calendarId, calendarAPIKey string) error {
	calendarService, err := calendar.NewService(context.Background(), option.WithAPIKey(calendarAPIKey))
	if err != nil {
		log.Fatalf("Unable to create Calendar service: %v", err)
	}

	fmt.Println("Removing Event")

	event := calendarService.Events.Delete(calendarId, id).Do()
	if event != nil {
		msg := fmt.Sprintf("Unable to delete event. %v", event)
		fmt.Println(msg)
		return model.NewLocAppError("Mattermost Time Away", msg, nil, "")
	}
	fmt.Printf("Event Deleted\n")
	return nil
}
