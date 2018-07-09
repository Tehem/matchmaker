package main

import (
	"github.com/Tehem/matchmaker/gcalendar"
	"github.com/Tehem/matchmaker/match"
	"github.com/Tehem/matchmaker/util"
	"github.com/transcovo/go-chpr-logger"
	"google.golang.org/api/calendar/v3"
	"io/ioutil"
)

func main() {
	yml, err := ioutil.ReadFile("./planning.yml")
	util.PanicOnError(err, "Can't yml planning")
	sessions, err := match.LoadPlan(yml)

	cal, err := gcalendar.GetGoogleCalendarService()
	util.PanicOnError(err, "Can't get gcalendar client")

	for _, session := range sessions {
		attendees := []*calendar.EventAttendee{}

		for _, person := range session.Reviewers.People {
			attendees = append(attendees, &calendar.EventAttendee{
				Email: person.Email,
			})
		}

		_, err := cal.Events.Insert("chauffeur-prive.com_k23ttdrv7g0l5i2vjj1f3s8voc@group.calendar.google.com", &calendar.Event{
			Start: &calendar.EventDateTime{
				DateTime: gcalendar.FormatTime(session.Range.Start),
			},
			End: &calendar.EventDateTime{
				DateTime: gcalendar.FormatTime(session.Range.End),
			},
			Summary: session.GetDisplayName(),
			Attendees: attendees,
		}).Do()
		util.PanicOnError(err, "Can't create event")
		logger.Info("✔ " + session.GetDisplayName())
	}
}
