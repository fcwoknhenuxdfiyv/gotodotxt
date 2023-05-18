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

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gotodotxt/tdt"
)

var (
	follow = false
)

func printJson(tf *tdt.TaskFile) {
	type output struct {
		SortOrder  string    `json:"sort_order"`
		ShowFuture bool      `json:"show_future"`
		TaskCount  int       `json:"task_count"`
		Tasks      tdt.Tasks `json:"tasks"`
	}
	var filtered tdt.Tasks
	for _, t := range tf.Tasks {
		if !t.FilteredOut {
			filtered = append(filtered, t)
		}
	}
	data := output{
		SortOrder:  tf.Opts.SortOrder,
		ShowFuture: tf.Opts.ShowFuture,
		TaskCount:  len(filtered),
		Tasks:      filtered,
	}
	js, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(js))
}

var jsonAliases = []string{"spoonfeed"}
var jsonCmd = &cobra.Command{
	Use:     "json",
	Aliases: jsonAliases,
	Short:   "Output filtered tasks as JSON",
	Long:    `Output filtered tasks as JSON`,
	Run: func(cmd *cobra.Command, args []string) {
		printOnExit = false
		davMode = checkDavMode()
		todoFile := viper.GetString("file")
		opts := tdt.Opts{
			SortOrder:  sortOrder,
			ShowFuture: showFuture,
		}
		if !follow {
			printJson(tdt.Read(todoFile, opts).Sort().Filter())
			return
		}
		file = tdt.Watch(todoFile, opts)
		printJson(file.Sort().Filter())
		for {
			select {
			case <-file.Events:
				file = tdt.Watch(todoFile, opts)
				printJson(file.Sort().Filter())
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(jsonCmd)
	jsonCmd.Flags().BoolVarP(&follow, "follow", "F", false, "Loop indefinitely")
}
