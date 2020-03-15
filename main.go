package main

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"go-mod/module"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

func info(file string) error {
	log.WithFields(log.Fields{
		"file": file,
	}).Info("Printing Info")

	m, err := module.Load(file)
	if (err == nil) {
		log.WithFields(log.Fields{
			"title": m.Title(),
			"num-patterns": m.NumPatterns(),
		}).Info("Info")
		for idx, sample := range m.Samples() {
			log.WithFields(log.Fields{
				"index":    idx,
				"name":     sample.Name(),
				"filename": sample.Filename(),
			}).Info("Sample")
		}

		if (m.Type() == module.PROTRACKER) {
			pt := m.(*module.ProTracker)
			for idx, patternNum := range pt.SequenceTable() {
				// If we've already moved past the song length in the sequence table, short circuit
				if (idx >= int(pt.SongLength())) {
					break
				}
				pattern, err := pt.GetPattern(patternNum)
				if (err == nil) {
					log.WithFields(log.Fields{
						"seq-no": idx,
						"pattern": patternNum,
						"channel": pattern.NumChannels(),
					}).Info()
				}
			}
		}

	} else {
		return err
	}

	return nil
}

func dumpAll(infile string, dir string) error {

	if !checkExists(infile) {
		return errors.New("Input file does not exist")
	} else if !checkExists(dir) {
		return errors.New("Output dir does not exist")
	}

	log.WithFields(log.Fields{
		"infile": infile,
		"dir": dir,
	}).Info("Exporting all samples")

	m, err := module.Load(infile)
	samples := m.Samples()

	destdir := filepath.Join(dir, filepath.Base(infile))
	os.MkdirAll(destdir, 0755)

	for idx, sample := range samples {
		outdata := sample.Data()
		if len(outdata) == 0 {
			log.Info("Ignoring sample index ",idx)
			continue
		}
		destname := fmt.Sprintf("%d-%s",idx,stripRegex(sample.Filename()))
		outpath := filepath.Join(destdir,destname)
		err = ioutil.WriteFile(outpath, outdata, 0644)
		if err != nil {
			return err
		}
		log.WithFields(log.Fields{
			"num-bytes": len(outdata),
			"out-file": outpath,
		}).Info("Wrote")
	}

	return nil
}

func checkExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func stripRegex(in string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	return reg.ReplaceAllString(in, "")
}

func main() {

	app := &cli.App {
		Name: "go-mod",
		Usage: "Do fun things with mod files.",
		Commands: []*cli.Command {
			&cli.Command {
				Name: "info",
				Usage: "Print info about a file.",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "file", Usage: "File to show" },
				},
				Action: func(c *cli.Context) error {
					f := c.String("file")
					return info(f)
				},
			},
			&cli.Command{
				Name: "dump-samples",
				Usage: "Dump all instrument samples.",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "file", Usage: "File to extract from." },
					&cli.StringFlag{Name: "dir", Usage: "Directory to write to." },
				},
				Action: func(c *cli.Context) error {
					f := c.String("file")
					d := c.String("dir")
					return dumpAll(f, d)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
