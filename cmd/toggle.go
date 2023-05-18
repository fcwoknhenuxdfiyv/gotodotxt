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
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gotodotxt/tdt"
)

var toggleAliases = []string{"x", "mark"}
var toggleCmd = &cobra.Command{
	Use:     "toggle",
	Aliases: toggleAliases,
	Short:   "Toggle task state (aliases: " + strings.Join(toggleAliases, ", ") + ")",
	Long: `Toggle the state of one or more tasks.

N.B. Marking recurring tasks as not done will not remove 
the recurring instance generated. You can use the delete
command to delete the unwanted task.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},
	Run: func(cmd *cobra.Command, args []string) {
		checkIds()
		davMode = checkDavMode()
		opts := tdt.Opts{
			ShowFuture: showFuture,
			SortOrder:  sortOrder,
		}
		file = tdt.Read(viper.GetString("file"), opts)
		file.Toggle(ids...)
		file.Write()
	},
}

func init() {
	rootCmd.AddCommand(toggleCmd)
}
