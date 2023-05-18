# gotodotxt

This is a small TUI program to manage `todo.txt` lists. Multiple lists can be added to the configuration and easily switched.

The default tasks file is "todo.txt" in the current directory. To use a different location, use the TODO_FILE environment variable.

## Example:

```
TODO_FILE=/some/directory/blah.txt gotodotxt
```

If all three webdav related parameters are supplied, the program will switch to WebDAV mode. In this mode, file update checks are done by polling every 15 seconds.

## Usage:

```
gotodotxt [flags]
gotodotxt [command]
```

## Available Commands:

```
archive     Archive completed tasks (aliases: a)
delete      Delete task(s) (aliases: del)
edit        Edit tasks (aliases: e, set)
help        Help about any command
json        Output filtered tasks as JSON
new         Create a new task (aliases: n, create, add)
toggle      Toggle task state (aliases: x, mark)
tui         Run in interactive mode
```

## Flags:

```
    --config string     config file
    --dav-pass string   webdav password
    --dav-url string    webdav base url
    --dav-user string   webdav user
-f, --future            show future tasks (default is false)
-h, --help              help for gotodotxt
-i, --ids ints          List of task ids
-s, --sort string       sort order (default "done,priority,due-,threshold-")
    --temp-dir string   non-standard temp directory
```

Use `gotodotxt [command] --help` for more information about a command.
