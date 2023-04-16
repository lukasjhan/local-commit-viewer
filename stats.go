package main

import (
	"fmt"
	"sort"
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

const outOfRange = 99999
const daysInLastSixMonths = 183
const weeksInLastSixMonths = 26

type column []int

func stats() {
	commits := getCommitMapFromRepos()
	printCommitsStats(commits)
}

func getBeginningOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	startOfDay := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
	return startOfDay
}

func countDaysSinceDate(date time.Time) int {
	days := 0
	now := getBeginningOfDay(time.Now())
	for date.Before(now) {
		date = date.Add(time.Hour * 24)
		days++
		if days > daysInLastSixMonths {
			return outOfRange
		}
	}
	return days
}

func fillCommits(path string, commits map[int]int) map[int]int {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return commits
	}
	ref, err := repo.Head()
	if err != nil {
		return commits
	}
	iterator, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return commits
	}
	offset := calcOffset()
	iterator.ForEach(func(c *object.Commit) error {
		daysAgo := countDaysSinceDate(c.Author.When) + offset
		if daysAgo != outOfRange {
			commits[daysAgo]++
		}
		return nil
	})

	return commits
}

func initCommitMap() map[int]int {
	commits := make(map[int]int, daysInLastSixMonths)
	for i := daysInLastSixMonths; i >= 0; i-- {
		commits[i] = 0
	}
	return commits
}

func getCommitMapFromRepos() map[int]int {
	filePath := getDotFilePath()
	repos := parseFileLinesToSlice(filePath)
	commits := initCommitMap()

	for _, path := range repos {
		commits = fillCommits(path, commits)
	}

	return commits
}

func calcOffset() int {
	offsets := map[time.Weekday]int{
		time.Sunday:    7,
		time.Monday:    6,
		time.Tuesday:   5,
		time.Wednesday: 4,
		time.Thursday:  3,
		time.Friday:    2,
		time.Saturday:  1,
	}

	weekday := time.Now().Weekday()
	return offsets[weekday]
}

func printCell(val int, today bool) {
	escape := "\033[0;37;30m"
	switch {
	case val > 0 && val < 5:
		escape = "\033[1;30;47m"
	case val >= 5 && val < 10:
		escape = "\033[1;30;43m"
	case val >= 10:
		escape = "\033[1;30;42m"
	}

	if today {
		escape = "\033[1;37;45m"
	}

	if val == 0 {
		fmt.Printf(escape + "  - " + "\033[0m")
		return
	}

	str := "  %d "
	switch {
	case val >= 10:
		str = " %d "
	case val >= 100:
		str = "%d "
	}

	fmt.Printf(escape+str+"\033[0m", val)
}

func printCommitsStats(commits map[int]int) {
	keys := sortMapIntoSlice(commits)
	cols := buildCols(keys, commits)
	printCells(cols)
}

func sortMapIntoSlice(m map[int]int) []int {
	var keys []int
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func buildCols(keys []int, commits map[int]int) map[int]column {
	cols := make(map[int]column)
	col := column{}

	for _, k := range keys {
		week := int(k / 7)
		dayinweek := k % 7

		if dayinweek == 0 {
			col = column{}
		}

		col = append(col, commits[k])

		if dayinweek == 6 {
			cols[week] = col
		}
	}

	return cols
}

func printCells(cols map[int]column) {
	printMonths()
	for j := 6; j >= 0; j-- {
		for i := weeksInLastSixMonths + 1; i >= 0; i-- {
			if i == weeksInLastSixMonths+1 {
				printDayCol(j)
			}
			if col, ok := cols[i]; ok {
				if i == 0 && j == calcOffset()-1 {
					printCell(col[j], true)
					continue
				} else {
					if len(col) > j {
						printCell(col[j], false)
						continue
					}
				}
			}
			printCell(0, false)
		}
		fmt.Printf("\n")
	}
}

func printMonths() {
	week := getBeginningOfDay(time.Now()).Add(-(daysInLastSixMonths * time.Hour * 24))
	month := week.Month()
	fmt.Printf("         ")
	for {
		if week.Month() != month {
			fmt.Printf("%s ", week.Month().String()[:3])
			month = week.Month()
		} else {
			fmt.Printf("    ")
		}

		week = week.Add(7 * time.Hour * 24)
		if week.After(time.Now()) {
			break
		}
	}
	fmt.Printf("\n")
}

func printDayCol(day int) {
	empty := "     "
	days := []string{empty, " Mon ", empty, " Wed ", empty, " Fri ", empty}
	fmt.Printf("%s", days[day])
}
