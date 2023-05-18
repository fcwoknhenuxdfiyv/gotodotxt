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
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/copier"
)

var (
	todoFile = "todo.txt"
	SortFile = false
)

func (tf *TaskFile) findTask(lineNumber int) (int, Task) {
	for i, t := range tf.Tasks {
		if t.LineNumber == lineNumber {
			// Log(log.Debug, t.original)
			return i, t
		}
	}
	return -1, Task{}
}

func (tf *TaskFile) Add(line string) *TaskFile {
	t, err := parseTask(line)
	if err != nil {
		return tf
	}
	t.Created = time.Now()
	today := YMD(t.Created)
	if t.Priority == "z" {
		t.original = today + " " + strings.TrimSpace(t.original)
	} else {
		fields := strings.Fields(t.original)
		line = strings.Join(fields[1:], " ")
		t.original = fmt.Sprintf("(%s) %s %s", t.Priority, today, line)
	}
	tf.Tasks = append(tf.Tasks, t)
	t.LineNumber = len(tf.Tasks)
	return tf
}

func (tf *TaskFile) Archive() *TaskFile {
	done := TaskFile{Path: getRelativeFileName("done")}
	var pending Tasks
	for _, t := range tf.Tasks {
		if t.IsDone() {
			done.Tasks = append(done.Tasks, t)
		} else {
			pending = append(pending, t)
		}
	}
	tf.Tasks = pending
	done.write(true)
	return tf
}

func (tf *TaskFile) Delete(nums ...int) *TaskFile {
	for _, num := range nums {
		i, t := tf.findTask(num)
		t.Deleted = true
		tf.Tasks[i] = t
	}
	trash := TaskFile{Path: getRelativeFileName("trash")}
	var pending Tasks
	for _, t := range tf.Tasks {
		if t.Deleted {
			trash.Tasks = append(trash.Tasks, t)
		} else {
			pending = append(pending, t)
		}
	}
	tf.Tasks = pending
	trash.write(true)
	return tf
}

func (tf *TaskFile) Edit(changes string, force bool, nums ...int) *TaskFile {
	p, t, d, r := parseChanges(changes)
	// Logf(log.Debugf, "%s - pri:%s t:%s due:%s rec:%s", changes, p, t, d, r)
	tf.setPriorities(p, nums...).
		setThresholds(t, force, nums...).
		setDueDates(d, force, nums...).
		setRecurrence(r, nums...)
	return tf
}

func (tf *TaskFile) Replace(replace string, num int) *TaskFile {
	if replace == "" {
		return tf
	}
	i, t := tf.findTask(num)
	if !t.IsDone() {
		t, err := parseTask(replace)
		// Logf(log.Debugf, "%+v", t)
		if err != nil {
			return tf
		}
		// Logf(log.Debugf, "%+v", t)
		tf.Tasks[i] = t
	}
	return tf
}

func (tf *TaskFile) setThresholds(threshold string, force bool, nums ...int) *TaskFile {
	if threshold == "" {
		return tf
	}
	threshold = strings.ToLower(threshold)
	for _, num := range nums {
		i, t := tf.findTask(num)
		if !t.IsDone() {
			if threshold == "x" {
				t.original = strings.TrimSpace(ThresholdRegex.ReplaceAllString(t.original, " "))
				t.HasThreshold = false
			} else if t.HasThreshold {
				// Log(log.Debug, t.Threshold)
				d := parseNewDate(threshold, t.Threshold)
				t.original = strings.ReplaceAll(t.original,
					"t:"+YMD(t.Threshold),
					"t:"+YMD(d))
				t.Threshold = d
				t.HasThreshold = true
			} else if force {
				d := parseNewDate(threshold, time.Now())
				t.original = t.original + " t:" + YMD(d)
				t.Threshold = d
				t.HasThreshold = true
			}
			tf.Tasks[i] = t
		}
	}
	return tf
}

func (tf *TaskFile) setDueDates(due string, force bool, nums ...int) *TaskFile {
	if due == "" {
		return tf
	}
	due = strings.ToLower(due)
	for _, num := range nums {
		i, t := tf.findTask(num)
		if !t.IsDone() {
			if due == "x" {
				t.original = strings.TrimSpace(DueRegex.ReplaceAllString(t.original, " "))
				t.HasDue = false
			} else if t.HasDue {
				d := parseNewDate(due, t.Due)
				t.original = strings.ReplaceAll(t.original,
					"due:"+YMD(t.Due),
					"due:"+YMD(d))
				t.Due = d
				t.HasDue = true
			} else if force {
				d := parseNewDate(due, time.Now())
				t.original = t.original + " due:" + YMD(d)
				t.Due = d
				t.HasDue = true
			}
			tf.Tasks[i] = t
		}
	}
	return tf
}

func (tf *TaskFile) setRecurrence(rec string, nums ...int) *TaskFile {
	if rec == "" {
		return tf
	}
	rec = strings.ToLower(rec)
	for _, num := range nums {
		i, t := tf.findTask(num)
		if !t.IsDone() {
			if rec == "x" {
				t.original = strings.TrimSpace(RecurrenceRegex.ReplaceAllString(t.original, " "))
			} else {
				if !t.HasDue && !t.HasThreshold {
					// There's not much point to adding recurrence without dates
					continue
				}
				r := newRecurrence(rec)
				if t.Recurrence.Period != "" {
					t.original = strings.ReplaceAll(t.original,
						"rec:"+t.Recurrence.String,
						"rec:"+r.String)
				} else {
					t.original += " rec:" + r.String
				}
			}
			tf.Tasks[i] = t
		}
	}
	return tf
}

func (tf *TaskFile) setPriorities(pri string, nums ...int) *TaskFile {
	if pri == "" {
		return tf
	}
	pri = strings.ToUpper(pri)
	for _, num := range nums {
		i, t := tf.findTask(num)
		if !t.IsDone() {
			if pri == "x" {
				t.original = strings.TrimSpace(PriorityRegex.ReplaceAllString(t.original, ""))
				t.Priority = "z"
			} else if t.Priority == "z" {
				t.Priority = pri
				t.original = fmt.Sprintf("(%s) %s", pri, t.original)
			} else {
				t.Priority = pri
				parts := strings.Fields(t.original)
				t.original = fmt.Sprintf("(%s) %s", pri, strings.Join(parts[1:], " "))
			}
			tf.Tasks[i] = t
		}
	}
	return tf
}

func (tf *TaskFile) Toggle(nums ...int) *TaskFile {
	for _, num := range nums {
		i, t := tf.findTask(num)
		if t.IsDone() {
			t.Done = 0
			parts := strings.Fields(t.original)
			t.original = strings.Join(parts[2:], " ")
			tf.Tasks[i] = t
			break
		}
		if t.Recurrence.Period != "" {
			var n Task
			copier.Copy(&n, &t)
			// Log(log.Debug, "1: "+n.Original)
			parts := strings.Fields(n.original)
			today := YMD(time.Now())
			if n.Priority == "z" {
				n.original = today + " " + strings.Join(parts[1:], " ")
				// Log(log.Debug, "2: "+n.Original)
			} else {
				n.original = fmt.Sprintf("(%s) %s %s", t.Priority, today,
					strings.Join(parts[2:], " "))
				// Log(log.Debug, "3: "+n.Original)
			}
			if n.HasDue {
				d := n.Recurrence.getNextDate(n.Due)
				n.original = strings.ReplaceAll(n.original,
					" due:"+YMD(n.Due),
					" due:"+YMD(d))
				n.Due = d
				// Log(log.Debug, "4: "+n.Original)
			}
			if n.HasThreshold {
				d := n.Recurrence.getNextDate(n.Threshold)
				n.original = strings.ReplaceAll(n.original,
					" t:"+YMD(n.Threshold),
					" t:"+YMD(d))
				n.Threshold = d
				// Log(log.Debug, "5: "+n.Original)
			}
			n.Created = time.Now()
			// Log(log.Debug, "6: "+n.Original)
			// Log(log.Warning, len(tasks))
			tf.Tasks = append(tf.Tasks, n)
			n.LineNumber = len(tf.Tasks)
			// Log(log.Warning, len(tasks))
		}
		t.Completed = time.Now()
		t.Done = 1
		t.original = "x " + YMD(t.Completed) + " " + t.original
		tf.Tasks[i] = t
	}
	return tf
}

func (t *Task) IsDone() bool {
	return t.Done != 0
}

func (tf *TaskFile) Original(num int) string {
	_, t := tf.findTask(num)
	return t.original
}
