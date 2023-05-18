package tdt

import "encoding/json"

type TaskAlias Task

type TaskJSON struct {
	*TaskAlias
	Original string `json:"original,omitempty"`
}

func (t *Task) MarshalJSON() ([]byte, error) {
	return json.Marshal(&TaskJSON{
		TaskAlias: (*TaskAlias)(t),
		// Unexported or custom-formatted fields are listed here:
		Original: t.original,
	})
}
