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
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/op/go-logging"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gotodotxt/tdt"
)

var (
	cfgFile string
	cfgDir  string
	tempDir string
	homeDir string

	// todoFile    = "todo.txt"
	sortOrder   = "done,priority,due-,threshold-"
	showFuture  = false
	sortFile    = false
	printOnExit = true

	davUrl      = ""
	davUser     = ""
	davPassword = ""
	davMode     = false

	file *tdt.TaskFile
	ids  []int
	log  *logging.Logger
)

var rootCmd = &cobra.Command{
	Use:   "gotodotxt",
	Short: "A small utility to manipulate todo.txt files",
	Long: `A small utility to manipulate todo.txt files

The default tasks file is "todo.txt" in the current
directory. To use a different location, use the TODO_FILE
environment variable.

Example:
TODO_FILE=/some/directory/blah.txt gotodotxt

If all three webdav related parameters are supplied,
the program will switch to WebDAV mode. In this mode,
file update checks are done by polling every 15 seconds.`,
	Run: func(cmd *cobra.Command, args []string) {
		davMode = checkDavMode()
		// fmt.Println(viper.AllKeys())
		// fmt.Println(">>> ", viper.GetString("file"))
		// fmt.Println(">>> ", viper.GetString("dav-user"))
		// fmt.Println(davMode, davUrl, davUser, davPassword)
		opts := tdt.Opts{
			ShowFuture: showFuture,
			SortOrder:  sortOrder,
		}
		file = tdt.Read(viper.GetString("file"), opts)
	},
}

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	if printOnExit && file != nil && len(file.Tasks) > 0 {
		printTasks(file.Sort().Filter())
	}
}

func checkIds() {
	if len(ids) == 0 {
		fmt.Println("No id (--ids/-i) supplied")
		os.Exit(1)
	}
}

func checkDavMode() bool {
	return tdt.SetWebdavCredentials(
		viper.GetString("dav-url"),
		viper.GetString("dav-user"),
		viper.GetString("dav-password"),
		viper.GetString("temp-dir"),
	)
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", cfgFile, "config file")
	rootCmd.PersistentFlags().BoolVarP(&showFuture, "future", "f", false, "show future tasks (default is false)")
	rootCmd.PersistentFlags().StringVarP(&sortOrder, "sort", "s", sortOrder, "sort order")
	rootCmd.PersistentFlags().IntSliceVarP(&ids, "ids", "i", nil, "List of task ids")
	rootCmd.PersistentFlags().StringVar(&davUrl, "dav-url", davUrl, "webdav base url")
	rootCmd.PersistentFlags().StringVar(&davUser, "dav-user", davUser, "webdav user")
	rootCmd.PersistentFlags().StringVar(&davPassword, "dav-pass", davPassword, "webdav password")
	rootCmd.PersistentFlags().StringVar(&tempDir, "temp-dir", tempDir, "non-standard temp directory")
}

func initConfig() {
	usingDefaultConfigFile := false
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		cfgDir = path.Dir(cfgFile)
	} else {
		var err error
		homeDir, err = os.UserHomeDir()
		cobra.CheckErr(err)
		cfgFile = path.Join(homeDir, ".config/gotodotxt/config.yaml")
		_, err = os.Stat(cfgFile)
		if !errors.Is(err, os.ErrNotExist) {
			cfgDir = path.Dir(cfgFile)
			viper.AddConfigPath(cfgDir)
			viper.SetConfigType("yaml")
			viper.SetConfigName("config.yaml")
			usingDefaultConfigFile = true
		}
	}

	if err := viper.ReadInConfig(); err == nil && !usingDefaultConfigFile {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	viper.SetDefault("file", "todo.txt")
	viper.SetDefault("temp-dir", os.TempDir())
	viper.SetEnvPrefix("todo")
	viper.BindEnv("file")
	viper.AutomaticEnv()

	viperSetWithFlags(rootCmd)

	if viper.GetInt("debug") > 0 {
		log = GetLogger(viper.GetInt("debug")-1, false)
		tdt.SetLogger(log)
	}

}

func viperSetWithFlags(cmd *cobra.Command) {
	viper.BindPFlags(cmd.Flags())
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !viper.IsSet(f.Name) && f.Value != nil {
			viper.Set(f.Name, f.Value)
		}
	})
	if cmd.HasSubCommands() {
		for _, c := range cmd.Commands() {
			viperSetWithFlags(c)
		}
	}
}
