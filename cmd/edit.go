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
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gotodotxt/tdt"
)

var (
	due,
	priority,
	recurrence,
	threshold,
	replace string
	force bool
)

var editAliases = []string{"e", "set"}
var editCmd = &cobra.Command{
	Use:     "edit",
	Aliases: editAliases,
	Short:   "Edit tasks (aliases: " + strings.Join(editAliases, ", ") + ")",
	Long: `Edit tasks

Due and threshold dates can me specified as a date in the
form of YYYY-MM-DD. "today", "tomorrow", "monday" (etc.) 
can also be used.

For recurrence, a pattern like 1w can be used to set the
recurrence to one week after the completion date. To
specify a strict recurrence, prefix the string with a "+"
(a plus character).

Priority is just a letter a-z (It will be made uppercase).`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},
	Run: func(cmd *cobra.Command, args []string) {
		checkIds()
		davMode = checkDavMode()
		opts := tdt.Opts{
			SortOrder:  sortOrder,
			ShowFuture: showFuture,
		}
		file = tdt.Read(viper.GetString("file"), opts)
		if replace != "" {
			if len(ids) != 1 {
				return
			}
			file.Replace(replace, ids[0])
		} else {
			file.Edit(fmt.Sprintf("%s %s %s %s", priority, threshold, due, recurrence), force, ids...)
		}
		file.Write()
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringVarP(&due, "due", "d", "", "Set due")
	editCmd.Flags().StringVarP(&priority, "pri", "P", "", "Set priority")
	editCmd.Flags().StringVarP(&recurrence, "rec", "r", "", "Recurrence pattern")
	editCmd.Flags().StringVarP(&threshold, "threshold", "t", "", "Set threshold")
	editCmd.Flags().StringVarP(&replace, "replace", "R", "", "Replace whole task")
	editCmd.Flags().BoolVarP(&force, "no-force", "F", true, "Don't force adding due/threshold")
}
