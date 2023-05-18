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

var deleteAliases = []string{"del"}
var deleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: deleteAliases,
	Short:   "Delete task(s) (aliases: " + strings.Join(deleteAliases, ", ") + ")",
	Long: `Delete task(s)

The tasks will be moved to trash.txt.
If the todo file is called something other than todo.txt,
the trash file will be called filename_trash.txt.`,
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
		file.Delete(ids...)
		file.Write()
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
