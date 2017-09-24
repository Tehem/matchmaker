package match

import (
	"sort"
	"math/rand"
)

func generateSquads(people []*Person, busyTimes []*BusyTime) []*Squad {
	masters := filterPersons(people, true)
	disciples := filterPersons(people, false)

	squads := []*Squad{}
	for _, master := range masters {
		for _, disciple := range disciples {
			masterExclusivity := master.GetExclusivity()
			discipleExclusivity := disciple.GetExclusivity()
			languagesCompatible := masterExclusivity == ExclusivityNone ||
				discipleExclusivity == ExclusivityNone ||
				masterExclusivity == discipleExclusivity
			characterCompatible := !(master.IsAnnoying && disciple.IsAnnoying)
			if languagesCompatible && characterCompatible {
				people := []*Person{master, disciple}
				squads = append(squads, &Squad{
					People:     people,
					BusyRanges: mergeBusyRanges(busyTimes, people),
				})
			}
		}
	}

	for i := range squads {
		j := rand.Intn(i + 1)
		squads[i], squads[j] = squads[j], squads[i]
	}

	sort.Sort(byExclusivity(squads))

	return squads
}

type byExclusivity []*Squad

func (a byExclusivity) Len() int      { return len(a) }
func (a byExclusivity) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byExclusivity) Less(i, j int) bool {
	iExclusivity := a[i].GetExclusivity()
	jExclusivity := a[j].GetExclusivity()
	switch iExclusivity {
	case ExclusivityMobile:
		return jExclusivity != ExclusivityMobile
	case ExclusivityBack:
		return jExclusivity == ExclusivityNone
	}
	return false
}

func filterPersons(persons []*Person, wantedIsGoodReviewer bool) []*Person {
	result := []*Person{}
	for _, person := range persons {
		if person.IsGoodReviewer == wantedIsGoodReviewer {
			result = append(result, person)
		}
	}
	return result
}

func mergeBusyRanges(busyTimes []*BusyTime, people []*Person) []*Range {
	mergedBusyRanges := []*Range{}
	for _, busyTime := range busyTimes {
		for _, person := range people {
			if busyTime.Person == person {
				mergedBusyRanges = mergeRangeListWithRange(mergedBusyRanges, busyTime.Range)
			}
		}
	}
	return mergedBusyRanges
}

func mergeRangeListWithRange(ranges []*Range, extraRange *Range) []*Range {
	mergedRangeList := []*Range{}
	rangeToAdd := extraRange
	for _, currentRange := range ranges {
		if haveIntersection(currentRange, extraRange) {
			rangeToAdd = mergeRanges(currentRange, rangeToAdd)
		} else {
			mergedRangeList = append(mergedRangeList, currentRange)
		}
	}
	return append(mergedRangeList, rangeToAdd)
}

func mergeRanges(range1 *Range, range2 *Range) *Range {
	result := &Range{}
	if range1.Start.Before(range2.Start) {
		result.Start = range1.Start
	} else {
		result.Start = range2.Start
	}
	if range1.End.After(range2.End) {
		result.End = range1.End
	} else {
		result.End = range2.End
	}
	return result
}

func haveIntersection(range1 *Range, range2 *Range) bool {
	return range1.End.After(range2.Start) && range2.End.After(range1.Start)
}
