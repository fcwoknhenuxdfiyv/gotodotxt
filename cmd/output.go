package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/gookit/color"
	"gotodotxt/tdt"
)

var (
	dateFormat = "2006-01-02"
)

type Row struct {
	LineNumber int
	Line1      string
	Line2      string
	Lines      int
}

type Rows []Row

func printTask(t tdt.Task, drawLineNumber bool) (string, string) {
	baseCol := color.White.Render
	priCol := baseCol
	taskCol := baseCol
	contextCol := color.Green.Render
	projectCol := color.Magenta.Render
	dateOkay := color.Green.Render
	dateLate := color.Red.Render

	if t.IsDone() {
		baseCol = color.Gray.Render
		priCol = baseCol
		taskCol = func(a ...interface{}) string {
			return baseCol(color.OpStrikethrough.Render(a...))
		}
		contextCol = baseCol
		projectCol = baseCol
		dateOkay = baseCol
		dateLate = baseCol
	} else {
		matches := tdt.ProjectsRegex.FindAllStringSubmatch(t.Description, -1)
		for _, found := range matches {
			if len(found) == 2 {
				t.Projects = append(t.Projects, found[1])
				t.Description = strings.ReplaceAll(t.Description,
					string(found[0]), projectCol(string(found[0])))
			}
		}
		matches = tdt.ContextsRegex.FindAllStringSubmatch(t.Description, -1)
		for _, found := range matches {
			if len(found) == 2 {
				t.Contexts = append(t.Contexts, found[1])
				t.Description = strings.ReplaceAll(t.Description,
					string(found[0]), contextCol(string(found[0])))
			}
		}
	}

	// Line 1
	line1 := ""
	if drawLineNumber {
		line1 += color.Gray.Render(fmt.Sprintf("%4d ", t.LineNumber))
	}

	if t.Priority == "z" {
		line1 += "    " + taskCol(t.Description)
	} else {
		line1 += priCol("("+t.Priority+")") + " " + taskCol(t.Description)
	}

	if !t.HasDue && !t.HasThreshold && t.Recurrence.Period == "" {
		return line1, ""
	}

	// Line 2
	line2 := ""
	if t.HasDue {
		col := dateOkay
		if t.Overdue {
			col = dateLate
		}
		line2 += col("due:" + t.Due.Format(dateFormat))
		// line2 += col("due:" + prettytime.Format(t.Due))
	}
	if t.HasThreshold {
		if line2 != "" {
			line2 += " "
		}
		col := dateOkay
		if t.Threshold.Before(time.Now()) {
			col = dateLate
		}
		col = color.Gray.Render
		line2 += col("t:" + t.Threshold.Format(dateFormat))
		// line2 += col("t:" + prettytime.Format(t.Threshold))
	}
	if t.Recurrence.Period != "" {
		if line2 != "" {
			line2 += " "
		}
		line2 += color.Gray.Render("rec:" + t.Recurrence.String)
	}
	line2 = "    " + line2
	if drawLineNumber {
		line2 = "     " + line2
	}

	return line1, line2
}

func printTasks(tf *tdt.TaskFile) {
	for _, t := range tf.Tasks {
		if t.FilteredOut {
			continue
		}
		line1, line2 := printTask(t, true)
		fmt.Println(line1)
		if line2 != "" {
			fmt.Println(line2)
		}
	}
	fmt.Println()
	fmt.Println(color.Gray.Render("         sort: " + sortOrder))
}

func renderTasks(tf *tdt.TaskFile, drawLineNumber bool) Rows {
	var rows []Row
	for _, t := range tf.Tasks {
		if t.FilteredOut {
			continue
		}
		line1, line2 := printTask(t, drawLineNumber)
		r := Row{
			LineNumber: t.LineNumber,
			Line1:      line1,
			Line2:      line2,
			Lines:      1,
		}
		if line2 != "" {
			r.Lines = 2
		}
		rows = append(rows, r)
	}
	return rows
}
