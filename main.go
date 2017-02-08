package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

type Options struct {
	Source      string
	Destination string
	Concurrency int
	Verbose     bool
	DryRun      bool
}

var opts Options

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	app := cli.NewApp()
	app.Usage = "generated pdf files from html files in a directory"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "src",
			Usage:       "dir with html files",
			Value:       "in",
			Destination: &opts.Source,
		},
		cli.StringFlag{
			Name:        "dest",
			Usage:       "dir where pdf files will be placed",
			Value:       "target",
			Destination: &opts.Destination,
		},
		cli.IntFlag{
			Name:        "concurrency",
			Usage:       "number of concurrent workers",
			Value:       25,
			Destination: &opts.Concurrency,
		},
		cli.BoolFlag{
			Name:        "verbose",
			Usage:       "display additional logging",
			Destination: &opts.Verbose,
		},
		cli.BoolFlag{
			Name:        "dryrun",
			Usage:       "do everything but generate the pdfs",
			Destination: &opts.DryRun,
		},
	}
	app.Action = Run
	app.Run(os.Args)
}

func Run(_ *cli.Context) error {
	src, in := WalkFiles(opts.Source)
	errs := make(chan error)

	wg := &sync.WaitGroup{}
	wg.Add(opts.Concurrency)

	target, err := filepath.Abs(opts.Destination)
	if err != nil {
		return errors.Wrapf(err, "unable to determine path for destination, %v", opts.Destination)
	}

	for i := 0; i < opts.Concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			Start(id, src, target, in, errs)
		}(i)
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	for err := range errs {
		check(err)
	}

	return nil
}

func WalkFiles(root string) (string, <-chan string) {
	in := make(chan string)

	dir, err := filepath.Abs(root)
	check(err)

	dirlen := len(dir) + 1

	go func() {
		defer close(in)
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			} else if !strings.HasSuffix(path, ".html") {
				return nil
			} else if strings.HasPrefix(path, ".") {
				return nil
			}

			in <- path[dirlen:]

			return nil
		})
		check(err)
	}()

	return dir, in
}

func Start(id int, src, target string, in <-chan string, errs chan<- error) {
	for path := range in {
		err := RenderPDF(id, src, target, path)
		if err != nil {
			errs <- errors.Wrapf(err, "[%v] unable to render file, %v", id, path)
			return
		}
	}

	errs <- nil
}

func RenderPDF(id int, src, target, path string) error {
	pdf := strings.Replace(path, ".html", ".pdf", -1)

	args := []string{
		"run",
		"--rm",
		"-v",
		fmt.Sprintf("%v:/work", filepath.Dir(filepath.Join(src, path))),
		"-v",
		fmt.Sprintf("%v:/dest", filepath.Dir(filepath.Join(target, pdf))),
		"savaki/genpdf:latest",
		"html-pdf.js",
		filepath.Base(path),
		fmt.Sprintf("/dest/%v", filepath.Base(pdf)),
	}
	if opts.Verbose {
		fmt.Printf("[%2d] rendering %v\n", id, path)
	}
	if opts.DryRun {
		return nil
	}

	os.MkdirAll(filepath.Dir(filepath.Join(target, pdf)), 0755)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = ioutil.Discard
	return cmd.Run()
}
