package main

import (
	"google.golang.org/api/calendar/v3"
	"github.com/transcovo/go-chpr-logger"
	"time"
	"github.com/rossille/matchmaker/gcalendar"
	"github.com/rossille/matchmaker/match"
	logrus2 "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"github.com/rossille/matchmaker/util"
)

func FirstDayOfISOWeek() time.Time {
	date := time.Now()
	date = time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), 0, 0, 0, date.Location())

	// iterate to Monday
	for !(date.Weekday() == time.Monday && date.Hour() == 0) {
		date = date.Add(time.Hour)
	}

	date = date.AddDate(0, 0, 0)

	return date
}

func GetWorkRange(beginOfWeek time.Time, day int, startHour int, startMinute int, endHour int, endMinute int) *match.Range {
	start := time.Date(
		beginOfWeek.Year(),
		beginOfWeek.Month(),
		beginOfWeek.Day()+day,
		startHour,
		startMinute,
		0,
		0,
		beginOfWeek.Location(),
	)
	end := time.Date(
		beginOfWeek.Year(),
		beginOfWeek.Month(),
		beginOfWeek.Day()+day,
		endHour,
		endMinute,
		0,
		0,
		beginOfWeek.Location(),
	)
	return &match.Range{
		Start: start,
		End:   end,
	}
}

func GetWeekWorkRanges(beginOfWeek time.Time) chan *match.Range {
	ranges := make(chan *match.Range)

	go func() {
		for day := 0; day < 5; day ++ {
			ranges <- GetWorkRange(beginOfWeek, day, 10, 0, 12, 30)
			ranges <- GetWorkRange(beginOfWeek, day, 13, 30, 19, 0)
		}
		close(ranges)
	}()

	return ranges
}

func parseTime(dateStr string) time.Time {
	result, err := time.Parse(time.RFC3339, dateStr)
	util.PanicOnError(err, "Impossible to parse date "+dateStr)
	return result
}

func ToSlice(c chan *match.Range) []*match.Range {
	s := make([]*match.Range, 0)
	for r := range c {
		s = append(s, r)
	}
	return s
}

func loadProblem() *match.Problem {
	people, err := match.LoadPersons("./persons.yml")
	util.PanicOnError(err, "Can't load people")
	logger.WithField("count", len(people)).Info("People loaded")

	cal, err := gcalendar.GetGoogleCalendarService()
	util.PanicOnError(err, "Can't get gcalendar client")
	logger.Info("Connected to google calendar")

	beginOfWeek := FirstDayOfISOWeek()

	workRanges := ToSlice(GetWeekWorkRanges(beginOfWeek))
	busyTimes := []*match.BusyTime{}
	for _, person := range people {
		personLogger := logger.WithField("person", person.Email)
		personLogger.Info("Loading busy detail")
		for _, workRange := range workRanges {
			personLogger.WithFields(logrus2.Fields{
				"start": workRange.Start,
				"end":   workRange.End,
			}).Info("Loading busy detail on range")
			result, err := cal.Freebusy.Query(&calendar.FreeBusyRequest{
				TimeMin: gcalendar.FormatTime(workRange.Start),
				TimeMax: gcalendar.FormatTime(workRange.End),
				Items: []*calendar.FreeBusyRequestItem{
					{
						Id: person.Email,
					},
				},
			}).Do()
			util.PanicOnError(err, "Can't retrive free/busy data for "+person.Email)
			busyTimePeriods := result.Calendars[person.Email].Busy
			println(person.Email + ":")
			for _, busyTimePeriod := range busyTimePeriods {
				println("  - " + busyTimePeriod.Start + " -> " + busyTimePeriod.End)
				busyTimes = append(busyTimes, &match.BusyTime{
					Person: person,
					Range: &match.Range{
						Start: parseTime(busyTimePeriod.Start),
						End:   parseTime(busyTimePeriod.End),
					},
				})
			}
		}
	}
	return &match.Problem{
		People:         people,
		WorkRanges:     workRanges,
		BusyTimes:      busyTimes,
		TargetCoverage: 2,
	}
}

func main() {
	problem := loadProblem()
	yml, _ := problem.ToYaml()
	ioutil.WriteFile("./problem.yml", yml, os.FileMode(0644))
}
