package tdt

import "time"

type TaskFile struct {
	Path       string
	Opts       Opts
	Tasks      Tasks
	SortOrder  string
	LastUpdate time.Time
	Events     chan FileChangedEvent
}

type FileChangedEvent struct {
	EventName string
}

type Opts struct {
	ShowFuture bool
	SortOrder  string
}

type Task struct {
	Done         int        `json:"done,omitempty"`
	Priority     string     `json:"priority,omitempty"`
	Completed    time.Time  `json:"completed,omitempty"`
	Created      time.Time  `json:"created,omitempty"`
	Description  string     `json:"description,omitempty"`
	Projects     []string   `json:"projects,omitempty"`
	Contexts     []string   `json:"contexts,omitempty"`
	Due          time.Time  `json:"due,omitempty"`
	HasDue       bool       `json:"has_due,omitempty"`
	Overdue      bool       `json:"overdue,omitempty"`
	Threshold    time.Time  `json:"threshold,omitempty"`
	HasThreshold bool       `json:"has_threshold,omitempty"`
	Recurrence   Recurrence `json:"recurrence,omitempty"`
	original     string     // no tag, see TaskJSON
	LineNumber   int        `json:"line_number,omitempty"`
	Deleted      bool       `json:"deleted,omitempty"`
	FilteredOut  bool       `json:"filtered_out,omitempty"`
}

type Tasks []Task

type Recurrence struct {
	Period string `json:"period"`
	Every  int64  `json:"every,omitempty"`
	Strict bool   `json:"strict,omitempty"`
	String string `json:"string,omitempty"`
}
