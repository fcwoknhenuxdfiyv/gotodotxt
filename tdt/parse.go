/*
Copyright © 2022 Jason Quigley <jason@jasonquigley.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package tdt

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	DoneRegex       = regexp.MustCompile(`^x\s+\d{4}-\d\d-\d\d\s+`)
	PriorityRegex   = regexp.MustCompile(`^\(([A-Z])\)\s+`)
	CreatedRegex    = regexp.MustCompile(`^(\d{4}-\d\d-\d\d)\s+`)
	ThresholdRegex  = regexp.MustCompile(`\s+t:(\d{4}-\d\d-\d\d?|[\+*[0-9]*[a-z]+?)\b`)
	DueRegex        = regexp.MustCompile(`\s+due:(\d{4}-\d\d-\d\d?|[\+*[0-9]*[a-z]+?)\b`)
	RecurrenceRegex = regexp.MustCompile(`\s+rec:(\+*)(\d*)([dwmy]+?)\b`)
	ProjectsRegex   = regexp.MustCompile(`\s+\+(\w+)\b`)
	ContextsRegex   = regexp.MustCompile(`\s+@(\w+)\b`)
)

func YMD(t time.Time) string {
	return t.Format("2006-01-02")
}

func parseYMD(date string) (time.Time, error) {
	return time.Parse("2006-01-02", date)
}

func parseDayOfWeek(d time.Weekday) string {
	return fmt.Sprintf("%dd", int((7+d-time.Now().Weekday())%7))
}

func parseNewDate(day string, d time.Time) time.Time {
	parsed, err := parseYMD(day)
	if err == nil {
		return parsed.Local()
	}

	day = strings.ToLower(day)
	switch day {
	case "today", "t", "tday", "tod":
		return time.Now()
	case "tomorrow", "tm", "tom":
		day = "d"
	case "monday", "mon":
		day = parseDayOfWeek(time.Monday)
	case "tuesday", "tue":
		day = parseDayOfWeek(time.Tuesday)
	case "wednesday", "wed":
		day = parseDayOfWeek(time.Wednesday)
	case "thursday", "thu":
		day = parseDayOfWeek(time.Thursday)
	case "friday", "fri":
		day = parseDayOfWeek(time.Friday)
	case "saturday", "sat":
		day = parseDayOfWeek(time.Saturday)
	case "sunday", "sun":
		day = parseDayOfWeek(time.Sunday)
	}

	r := newRecurrence(day)
	if r.Period != "" {
		d = r.getNextDate(d)
	}

	return d.Local()
}

func newRecurrence(rec string) Recurrence {
	found := RecurrenceRegex.FindStringSubmatch(" rec:" + rec + " ")
	if len(found) != 4 {
		Log(log.Error, "Invalid format", rec)
		return Recurrence{}
	}
	every, err := strconv.ParseInt((found[2]), 10, 32)
	if err != nil {
		every = 1
	}
	r := Recurrence{
		Period: strings.ToLower(string(found[3])),
		Every:  every,
		Strict: string(found[1]) == "+",
		String: strings.TrimPrefix(strings.TrimSpace(found[0]), "rec:"),
	}
	return r
}

func (r Recurrence) getNextDate(d time.Time) time.Time {
	// Log(log.Debug, r)
	// Log(log.Debug, YMD(d))
	if !r.Strict {
		d = time.Now().Truncate(24 * time.Hour)
	}
	// Log(log.Debug, YMD(d))
	switch r.Period {
	case "d":
		d = d.AddDate(0, 0, int(r.Every))
	case "w":
		d = d.AddDate(0, 0, int(r.Every*7))
	case "m":
		d = d.AddDate(0, int(r.Every), 0)
	case "y":
		d = d.AddDate(int(r.Every), 0, 0)
	default:
		Log(log.Error, "eep")
	}
	// Log(log.Debug, YMD(d))
	return d
}

func parseChanges(line string) (string, string, string, string) {
	fields := strings.Fields(line)
	line = strings.Join(fields, "  ") + " "
	if !strings.HasPrefix(line, "(") {
		line = " " + line
	}
	var priority, threshold, due, recurrence string

	found := PriorityRegex.FindStringSubmatch(line)
	// Logf(log.Infof, "%+v %d", found, len(found))
	if len(found) == 2 {
		priority = strings.TrimSpace(string(found[1]))
	} else {
		priority = ""
	}

	found = ThresholdRegex.FindStringSubmatch(line)
	// Logf(log.Infof, "%+v %d", found, len(found))
	if len(found) == 2 {
		threshold = strings.TrimSpace(string(found[1]))
	} else {
		threshold = ""
	}

	found = DueRegex.FindStringSubmatch(line)
	// Logf(log.Infof, "%+v %d", found, len(found))
	if len(found) == 2 {
		due = strings.TrimSpace(string(found[1]))
	} else {
		due = ""
	}

	found = RecurrenceRegex.FindStringSubmatch(line)
	// Logf(log.Infof, "%+v %d", found, len(found))
	if len(found) == 4 {
		recurrence = strings.TrimSpace(string(found[1]) + string(found[2]) + string(found[3]))
	} else {
		recurrence = ""
	}

	return priority, threshold, due, recurrence
}

func parseTask(line string) (Task, error) {
	parts := strings.Fields(line)
	if len(parts) < 1 {
		return Task{}, errors.New("bad syntax")
	}

	task := Task{
		original: line,
	}

	var err error
	done := DoneRegex.MatchString(line)
	if done {
		task.Done = 1
		task.Completed, err = parseYMD(parts[1])
		if err != nil {
			return Task{}, err
		}
		// skip over the 'x' and completion date
		parts = parts[2:]
		line = strings.Join(parts, " ")
	}

	priority, threshold, due, recurrence := parseChanges(line)
	// Log(log.Debug, priority, threshold, due, recurrence)

	if priority != "" {
		task.Priority = priority
		parts = parts[1:]
		line = strings.Join(parts, " ")
	} else {
		task.Priority = "z"
	}

	found := CreatedRegex.FindStringSubmatch(line)
	// Logf(log.Infof, "%+v %d", found, len(found))
	if len(found) == 2 {
		task.Created, err = parseYMD((found[1]))
		if err != nil {
			return Task{}, err
		}
		parts = parts[1:]
		line = strings.Join(parts, " ")
	}

	if threshold != "" {
		task.Threshold = parseNewDate(threshold, time.Now())
		task.HasThreshold = true
		line = strings.ReplaceAll(line, " t:"+threshold, "")
		task.original = strings.ReplaceAll(task.original,
			" t:"+threshold,
			" t:"+YMD(task.Threshold))
		parts = strings.Fields(line)
		// } else {
		// 	task.Threshold, _ = parseYMD("9999-12-31")
	}

	if due != "" {
		task.Due = parseNewDate(due, time.Now())
		task.HasDue = true
		if task.Due.Before(time.Now()) {
			task.Overdue = true
		}
		line = strings.ReplaceAll(line, " due:"+due, "")
		task.original = strings.ReplaceAll(task.original,
			" due:"+due,
			" due:"+YMD(task.Due))
		parts = strings.Fields(line)
		// } else {
		// 	task.Due, _ = parseYMD("9999-12-31")
	}

	if recurrence != "" {
		task.Recurrence = newRecurrence(recurrence)
		line = strings.ReplaceAll(line, " rec:"+recurrence, "")
		parts = strings.Fields(line)
	}

	matches := ProjectsRegex.FindAllStringSubmatch(line, -1)
	for _, found := range matches {
		if len(found) == 2 {
			task.Projects = append(task.Projects, found[1])
		}
	}

	matches = ContextsRegex.FindAllStringSubmatch(line, -1)
	for _, found := range matches {
		if len(found) == 2 {
			task.Contexts = append(task.Contexts, found[1])
		}
	}

	task.Description = strings.Join(strings.Fields(line), " ")

	return task, nil
}
