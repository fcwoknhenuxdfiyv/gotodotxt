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
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/go-homedir"
	"github.com/studio-b12/gowebdav"
)

var (
	davUrl       = ""
	davUser      = ""
	davPassword  = ""
	davTmpDir    = ""
	davMode      = false
	stopWatching chan struct{}
	writeLock    = sync.RWMutex{}
)

func SetWebdavCredentials(url, user, password, tmpDir string) bool {
	if url == "" || user == "" || password == "" {
		return false
	}
	davMode = true
	davUrl = url
	davUser = user
	davPassword = password
	davTmpDir = tmpDir
	return true
}

func getRelativeFileName(fn string) string {
	Log(log.Debug, todoFile)
	relFn := path.Join(path.Dir(todoFile), fn+".txt")
	if path.Base(todoFile) != "todo.txt" {
		name := strings.TrimSuffix(path.Base(todoFile), ".txt")
		relFn = path.Join(path.Dir(todoFile), name+"_"+fn+".txt")
	}
	return relFn
}

func Read(fn string, opts Opts) *TaskFile {
	todoFile = fn
	var err error
	fn, err = homedir.Expand(fn)
	if err != nil {
		panic(err)
	}
	_, err = os.Stat(fn)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		_, err := os.Create(fn)
		if err != nil {
			panic(err)
		}
	}
	if err != nil {
		panic(err)
	}
	taskFile := TaskFile{Path: fn, Opts: opts}
	if davMode {
		fn, taskFile.LastUpdate = downloadWebdavFile(fn)
		taskFile.Tasks, _ = readTasksFile(fn)
	} else {
		taskFile.Tasks, taskFile.LastUpdate = readTasksFile(fn)
	}
	return &taskFile
}

func readTasksFile(fn string) (Tasks, time.Time) {
	file, err := os.Open(fn)
	if err != nil {
		log.Fatalf("readTasksFile %v", err)
	}
	defer file.Close()

	var tasks []Task
	scanner := bufio.NewScanner(file)
	var lineNumber int = 0
	for scanner.Scan() {
		line := scanner.Text()
		t, err := parseTask(line)
		if err != nil {
			continue
		}
		t.LineNumber = lineNumber
		lineNumber++
		tasks = append(tasks, t)
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("readTasksFile %v", err)
	}

	info, err := os.Stat(fn)
	if err != nil {
		panic(err)
	}

	return tasks, info.ModTime()
}

func Watch(fn string, opts Opts) *TaskFile {
	var err error
	fn, err = homedir.Expand(fn)
	if err != nil {
		panic(err)
	}
	taskFile := Read(fn, opts)
	taskFile.Events = make(chan FileChangedEvent)
	if davMode {
		go watchWebdavFile(fn, taskFile.Events)
	} else {
		go watchLocalFile(fn, taskFile.Events)
	}
	return taskFile
}

func watchWebdavFile(fn string, changed chan FileChangedEvent) {
	if stopWatching != nil {
		stopWatching <- struct{}{}
		time.Sleep(1 * time.Second)
	}
	stopWatching = make(chan struct{})

	c := gowebdav.NewClient(davUrl, davUser, davPassword)
	info, err := c.Stat(fn)
	if err != nil {
		panic(err)
	}
	updatedAt := info.ModTime()
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-stopWatching:
			Log(log.Warning, "stop watching "+fn)
			return
		case <-ticker.C:
			Log(log.Warning, "checking "+fn)
			info, err := c.Stat(fn)
			if err != nil {
				panic(err)
			}
			if info.ModTime().After(updatedAt) {
				Log(log.Warning, "modified "+fn)
				updatedAt = info.ModTime()
				changed <- FileChangedEvent{"WebDAV"}
			}
		}
	}
}

func watchLocalFile(fn string, changed chan FileChangedEvent) {
	if stopWatching != nil {
		stopWatching <- struct{}{}
		time.Sleep(1 * time.Second)
	}
	stopWatching = make(chan struct{})

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	err = watcher.Add(fn)
	if err != nil {
		panic(err)
	}
	changing := false
	for {
		select {
		case <-stopWatching:
			Log(log.Warning, "stop watching "+fn)
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// This works around multiple FSEvents on macOS
			if !changing {
				Log(log.Warning, "modified "+fn)
				changing = true
				go func() {
					time.Sleep(time.Second * 4)
					changed <- FileChangedEvent{event.Op.String()}
					changing = false
				}()
			}
		}
	}
}

func (tf *TaskFile) Write() {
	tf.write(false)
}

func (tf *TaskFile) write(addToEnd bool) {
	writeLock.Lock()
	if davMode {
		tmp, _ := downloadWebdavFile(tf.Path)
		tf.writeLocalFile(tmp, addToEnd)
		uploadWebdavFile(tmp, tf.Path)
	} else {
		tf.writeLocalFile(tf.Path, addToEnd)
	}
	writeLock.Unlock()
}

func (tf *TaskFile) writeLocalFile(fn string, addToEnd bool) {
	var newContents []string
	for _, t := range tf.sort("").Tasks {
		// Log(log.Debug, t.Original)
		newContents = append(newContents, t.original)
	}
	if SortFile {
		sort.Strings(newContents)
	}

	var f *os.File
	var err error
	if addToEnd {
		opts := os.O_APPEND | os.O_CREATE | os.O_WRONLY
		f, err = os.OpenFile(fn, opts, 0644)
	} else {
		f, err = os.Create(fn)
	}
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for _, t := range newContents {
		_, err = f.WriteString(t + "\n")
		if err != nil {
			panic(err)
		}
	}
}

func downloadWebdavFile(fn string) (string, time.Time) {
	tmp := path.Join(davTmpDir, "__"+path.Base(fn))

	file, err := os.Create(tmp)
	if err != nil {
		log.Fatalf("downloadWebdavFile %v", err)
	}
	defer file.Close()

	c := gowebdav.NewClient(davUrl, davUser, davPassword)
	reader, err := c.ReadStream(fn)
	if e, ok := err.(*os.PathError); ok && e.Err.Error() == "404" {
		return tmp, time.Now()
	} else if err != nil {
		log.Fatalf("downloadWebdavFile %v", err)
	}

	info, err := c.Stat(fn)
	if err != nil {
		panic(err)
	}

	io.Copy(file, reader)

	return tmp, info.ModTime()
}

func uploadWebdavFile(tmp, fn string) {
	file, _ := os.Open(tmp)
	defer file.Close()

	c := gowebdav.NewClient(davUrl, davUser, davPassword)
	c.WriteStream(fn, file, 0644)
}
