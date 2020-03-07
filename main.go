package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"go-mod/module"
	"io/ioutil"
	"os"
)

func info(file string) error {
	log.WithFields(log.Fields{
		"file": file,
	}).Info("Printing Info")

	m, err := module.Load(file)
	pt := m.(*module.ProTracker)
	if (err == nil) {
		log.WithFields(log.Fields{
			"title": pt.Title(),
			"song-length": pt.SongLength(),
			"num-patterns": pt.NumPatterns(),
		}).Info("Info")
		for idx, instrument := range pt.Instruments() {
			log.WithFields(log.Fields{
				"index": idx,
				"name": instrument.Name(),
			}).Info("Instrument")
		}
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

	} else {
		return err
	}

	return nil
}

func export(infile string, index int, outfile string) error {
	log.WithFields(log.Fields{
		"infile": infile,
		"outfile": outfile,
		"index": index,
	}).Info("Exporting sample")

	m, err := module.Load(infile)
	pt := m.(*module.ProTracker)
	if (err != nil) {
		return err
	}
	i,err := pt.GetInstrument(index)
	if (err != nil) {
		return err
	}
	outdata := i.Data()
	err = ioutil.WriteFile(outfile, outdata,0644)
	if (err != nil) {
		return err
	}

	log.WithFields(log.Fields{
		"num-bytes": len(outdata),
		"out-file": outfile,
	}).Info("Wrote")

	return nil
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
				Name: "extract",
				Usage: "Extract an instrument sample.",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "file", Usage: "File to extract from." },
					&cli.StringFlag{Name: "out", Usage: "File to write sample to." },
					&cli.StringFlag{Name: "idx", Usage: "Instrument index to extract (starts at 0)" },
				},
				Action: func(c *cli.Context) error {
					f := c.String("file")
					idx := c.Int("idx")
					o := c.String("out")
					return export(f, idx, o)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
