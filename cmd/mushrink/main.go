package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
	"hack.moontide.ink/lukas/binfo"
	"hack.moontide.ink/lukas/mushrink/internal/pathutil"
	"hack.moontide.ink/lukas/mushrink/internal/planning"
)

var bi = binfo.MustGet()

func main() {
	var (
		from, to             string
		pattern, repl        string
		workers              int
		verbose, interactive bool
	)

	cli.VersionPrinter = func(cmd *cli.Command) {
		_, _ = fmt.Fprintf(
			cmd.Root().Writer,
			"%s\n",
			bi.Summarize(
				cmd.Name,
				cmd.Version,
				binfo.Multiline|binfo.Build|binfo.VCS|binfo.Module|binfo.CGO,
			),
		)
	}

	app := &cli.Command{
		Name:    "mushrink",
		Usage:   "compress music",
		Version: bi.Module.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "match",
				Usage:       "pattern for source files",
				Destination: &pattern,
				Aliases:     []string{"m"},
			},
			&cli.StringFlag{
				Name:        "replace",
				Usage:       "replacement for destination files",
				Destination: &repl,
				Aliases:     []string{"r"},
			},
			&cli.IntFlag{
				Name:        "workers",
				Usage:       "number of worker processes",
				Destination: &workers,
				Value:       runtime.NumCPU(),
				Aliases:     []string{"w"},
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Usage:       "verbose output",
				Destination: &verbose,
				Value:       false,
				Aliases:     []string{"v"},
			},
			&cli.BoolFlag{
				Name:        "interactive",
				Usage:       "confirm before executing plan",
				Destination: &interactive,
				Value:       false,
				Aliases:     []string{"i"},
			},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:        "from",
				UsageText:   "from directory",
				Destination: &from,
			},
			&cli.StringArg{
				Name:        "to",
				UsageText:   "to directory",
				Destination: &to,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) (err error) {
			if pattern == "" && repl == "" {
				pattern = `\.flac$`
				repl = `.mp3`
			}

			if pattern == "" || repl == "" {
				err = errors.New("both the pattern and replacement have to be specified")
				return
			}

			if from == "" {
				err = errors.New("please specify the source directory")
				return
			}

			if to == "" {
				err = errors.New("please specify the destination directory")
				return
			}

			if cmd.Args().Len() > 0 {
				err = errors.New("too many arguments")
				return
			}

			err = os.MkdirAll(to, 0777)
			if err != nil {
				return
			}

			from, err = pathutil.Resolve(from)
			if err != nil {
				return
			}

			to, err = pathutil.Resolve(to)
			if err != nil {
				return
			}

			if from == to {
				err = errors.New("source and destination paths point to the same location, this was probably not intended")
				return
			}

			var fi os.FileInfo
			fi, err = os.Stat(from)
			if err != nil {
				return
			}
			if !fi.IsDir() {
				return fmt.Errorf("%s is not a directory", from)
			}

			re, err := regexp.Compile(pattern)
			if err != nil {
				err = fmt.Errorf("pattern: %w", err)
				return
			}

			var plan []planning.Task
			if verbose {
				fmt.Fprintln(cmd.ErrWriter, "Planning...")
			}
			plan, err = planning.Plan(from, to, re, repl)
			if err != nil {
				return
			}
			if len(plan) == 0 {
				if verbose {
					fmt.Fprintln(cmd.Writer, "Nothing to do")
				}
				return nil
			}
			if verbose {
				for _, task := range plan {
					fmt.Fprintln(cmd.ErrWriter, task)
				}
			}

			if interactive {
				r := bufio.NewReader(os.Stdin)
				fmt.Fprint(cmd.ErrWriter, "Do you want to continue? [Y/n] ")
				var line []byte
				line, _, err = r.ReadLine()
				if err != nil {
					return
				}
				switch strings.ToLower(string(line)) {
				case "yes":
				case "y":
				case "":
				default:
					return
				}
			}

			g := new(errgroup.Group)
			g.SetLimit(workers)
			for _, task := range plan {
				g.Go(func() error {
					err := task.Run()
					if err != nil {
						return err
					}
					fmt.Fprintf(cmd.Writer, "finished %s\n", task)
					return nil
				})
			}
			if err := g.Wait(); err != nil {
				return err
			}

			return nil
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %s\n", err)
		os.Exit(1)
	}
}
