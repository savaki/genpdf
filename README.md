# genpdf
docker utility to generate pdf from an html page

# Usage

```
NAME:
   main - generated pdf files from html files in a directory

USAGE:
   main [global options] command [command options] [arguments...]

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --src value          dir with html files (default: "in")
   --dest value         dir where pdf files will be placed (default: "target")
   --concurrency value  number of concurrent workers (default: 25)
   --verbose            display additional logging
   --dryrun             do everything but generate the pdfs
   --help, -h           show help
   --version, -v        print the version
```

## Example

```
$ go run main.go --src . --verbose
[ 0] rendering /Users/matt/src/github.com/savaki/genpdf/sample.html
[ 9] rendering /Users/matt/src/github.com/savaki/genpdf/statement.html
```