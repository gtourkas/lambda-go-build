package main

import (
	"fmt"
	"github.com/go-cmd/cmd"
	"github.com/nozzle/throttler"
	"github.com/urfave/cli/v2"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func build(baseSrcPath string, lambdaSrcPath string, baseDestPath string, debugMode bool) {

	lambdaRelDestPath, _ := filepath.Rel(baseSrcPath, lambdaSrcPath)

	output := path.Join(baseDestPath, lambdaRelDestPath, "main")

	args := []string{
		"build",
		"-o",
		output,
	}
	if debugMode {
		args = append(args, []string{"-gcflags", "all=-N -l"}...)
	}

	args = append(args, lambdaSrcPath)

	cmd := cmd.NewCmd("go", args...)
	cmd.Dir = filepath.Dir(lambdaSrcPath)
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")

	<-cmd.Start()

	status := cmd.Status()
	result := "OK"
	if status.Exit != 0 {
		result = "FAILED"
	}
	fmt.Printf("%s %s\n", result, lambdaSrcPath)
	for i := range status.Stdout {
		l := status.Stdout[i]
		fmt.Printf("\t%s\n", l)
	}
	for i := range status.Stderr {
		l := status.Stderr[i]
		fmt.Printf("\t%s\n", l)
	}
	return
}

func main() {

	app := &cli.App{
		Authors:         []*cli.Author{{"George Tourkas", "gtourkas at gmail dot com"}},
		Usage:           "Tool for building  Golang AWS Lambdas",
		HideHelpCommand: true,
		HideHelp:        true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Required: true,
				Name:     "source-path",
				Aliases:  []string{"s"},
				Usage:    "the base path for the lambdas source",
			},
			&cli.StringFlag{
				Required: true,
				Name:     "destination-path",
				Aliases:  []string{"d"},
				Usage:    "the base path for the lambdas built output",
			},
			&cli.IntFlag{
				Name:    "concurrent-builds",
				Aliases: []string{"cb"},
				Value:   2,
				Usage:   "the number of concurrent build queues",
			},
			&cli.BoolFlag{
				Name:    "debug-mode",
				Aliases: []string{"dm"},
				Value:   false,
				Usage:   "whether to build for local step-through debugging",
			},
			&cli.StringFlag{
				Name:        "source-path-pattern",
				Aliases:     []string{"spp"},
				Value:       "",
				Usage:       "source path pattern; use to build only a few lambdas",
				DefaultText: "none = build all lambdas",
			},
		},
		Action: func(c *cli.Context) error {

			var err error
			var baseSrcPath string
			baseSrcPath, err = filepath.Abs(c.String("source-path"))
			if err != nil {
				panic(err)
			}

			var baseDestPath string
			baseDestPath, err = filepath.Abs(c.String("destination-path"))
			if err != nil {
				panic(err)
			}

			concurBuilds := c.Int("concurrent-builds")

			debugMode := c.Bool("debug-mode")

			pattern := c.String("path-pattern")

			// get paths to lambdas
			var found []string
			err = filepath.Walk(baseSrcPath, func(path string, info os.FileInfo, err error) error {
				// stop and propagate any walking errors
				if err != nil {
					return err
				}
				if !info.IsDir() && strings.ToLower(info.Name()) == "main.go" {
					dir := filepath.Dir(path)
					if pattern != "" {
						if strings.Contains(path, pattern) {
							found = append(found, dir)
						}
					} else {
						found = append(found, dir)
					}
				}

				return nil
			})
			if err != nil {
				panic(err)
			}

			// concurrently run the handler builds
			t := throttler.New(concurBuilds, len(found))
			for _, f := range found {
				go func(lambdaSrcPath string) {
					build(baseSrcPath, lambdaSrcPath, baseDestPath, debugMode)
					t.Done(nil)
				}(f)
				t.Throttle()
			}

			return nil
		},
	}

	app.Run(os.Args)
}
