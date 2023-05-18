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
	"sort"
	"strings"
	"time"
)

func (tf *TaskFile) Filter() *TaskFile {
	for i, t := range tf.Tasks {
		t.FilteredOut = false
		if !(tf.Opts.ShowFuture || (!t.HasThreshold || t.Threshold.Before(time.Now()))) {
			t.FilteredOut = true
		}
		tf.Tasks[i] = t
	}
	return tf
}

func (tf *TaskFile) Sort() *TaskFile {
	tf.sort(tf.Opts.SortOrder)
	return tf
}

func (tf *TaskFile) sort(sortOrder string) *TaskFile {
	sortFields := strings.Fields(
		strings.ReplaceAll(strings.ToLower(sortOrder), ",", " "))
	sort.Slice(tf.Tasks, func(i, j int) bool {
		for _, f := range sortFields {
			switch f {
			case "due", "due+", "d", "d+":
				if tf.Tasks[j].Due.IsZero() {
					return false
				}
				if tf.Tasks[i].Due != tf.Tasks[j].Due {
					return tf.Tasks[i].Due.After(tf.Tasks[j].Due)
				}
			case "due-", "d-":
				if tf.Tasks[j].Due.IsZero() {
					return true
				}
				if tf.Tasks[i].Due != tf.Tasks[j].Due {
					return tf.Tasks[i].Due.Before(tf.Tasks[j].Due)
				}
			case "threshold", "threshold+", "t", "t+":
				if tf.Tasks[j].Threshold.IsZero() {
					return false
				}
				if tf.Tasks[i].Threshold != tf.Tasks[j].Threshold {
					return tf.Tasks[i].Threshold.After(tf.Tasks[j].Threshold)
				}
			case "threshold-", "t-":
				if tf.Tasks[j].Threshold.IsZero() {
					return true
				}
				if tf.Tasks[i].Threshold != tf.Tasks[j].Threshold {
					return tf.Tasks[i].Threshold.Before(tf.Tasks[j].Threshold)
				}
			case "done", "done+", "x", "x+":
				if tf.Tasks[i].Done != tf.Tasks[j].Done {
					return tf.Tasks[i].Done < tf.Tasks[j].Done
				}
			case "done-", "x-":
				if tf.Tasks[i].Done != tf.Tasks[j].Done {
					return tf.Tasks[i].Done > tf.Tasks[j].Done
				}
			case "priority", "priority+", "pri", "pri+", "P", "P+":
				if tf.Tasks[i].Priority == "" {
					return false
				}
				if tf.Tasks[i].Priority != tf.Tasks[j].Priority {
					return tf.Tasks[i].Priority < tf.Tasks[j].Priority
				}
			case "priority-", "pri-", "P-":
				if tf.Tasks[i].Priority == "" {
					return false
				}
				if tf.Tasks[i].Priority != tf.Tasks[j].Priority {
					return tf.Tasks[i].Priority > tf.Tasks[j].Priority
				}
			}
		}
		return tf.Tasks[i].LineNumber < tf.Tasks[j].LineNumber
	})
	return tf
}
