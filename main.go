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
	in := WalkFiles(opts.Source)
	errs := make(chan error)

	wg := &sync.WaitGroup{}
	wg.Add(opts.Concurrency)

	for i := 0; i < opts.Concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			Start(id, in, errs)
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

func WalkFiles(root string) <-chan string {
	in := make(chan string)

	dir, err := filepath.Abs(root)
	check(err)

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

			in <- path

			return nil
		})
		check(err)
	}()

	return in
}

func Start(id int, in <-chan string, errs chan<- error) {
	for path := range in {
		err := RenderPDF(id, path)
		if err != nil {
			errs <- errors.Wrapf(err, "[%v] unable to render file, %v", id, path)
			return
		}
	}

	errs <- nil
}

func RenderPDF(id int, path string) error {

	pdf := filepath.Base(strings.Replace(path, ".html", ".pdf", -1))

	args := []string{
		"run",
		"--rm",
		"-v",
		fmt.Sprintf("%v:/work", filepath.Dir(path)),
		"savaki/genpdf:latest",
		"html-pdf.js",
		filepath.Base(path),
		pdf,
	}
	if opts.Verbose {
		fmt.Printf("[%2d] rendering %v\n", id, path)
	}
	if opts.DryRun {
		return nil
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdout = ioutil.Discard
	cmd.Stderr = ioutil.Discard
	return cmd.Run()
}
