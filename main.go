package main

import
(
	"go-mod/module"
	"log"
	"github.com/alexcesaro/log/stdlog"
	"os"
	"github.com/codegangsta/cli"
	"io/ioutil"
)


func main() {

	app := cli.NewApp()
	logger := stdlog.GetFromFlags()
	app.Name = "go-mod"
	app.Usage = "Do fun things with mod files."
	app.Author = "mrbarge"
	app.Flags = []cli.Flag {
		cli.StringFlag {
			Name: "log",
		},
	}
	app.Commands = []cli.Command {
		{
			Name: "play",
			Usage: "Play a file.",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "file",
					Usage: "File to play",
				},
			},
			Action: func(c *cli.Context) {
				m, err := module.Load(c.String("file"))
				if (err != nil) {
					pt := m.(*module.ProTracker)
					pt.Play()
				} else {
					log.Fatal(err)
				}
			},
		},
		{
			Name: "info",
			Usage: "Print info about a file.",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "file",
					Usage: "File to play",
				},
			},
			Action: func(c *cli.Context) {
				m, err := module.Load(c.String("file"))
				pt := m.(*module.ProTracker)
				if (err == nil) {
					logger.Infof("%s",pt.Title())
					logger.Infof("song length %d",pt.SongLength())
					logger.Infof("Num patterns: %d",pt.NumPatterns())
					for idx, instrument := range pt.Instruments() {
						logger.Infof("Instrument %d: %s",idx,instrument.Name())
					}
					for idx, patternNum := range pt.SequenceTable() {
						// If we've already moved past the song length in the sequence table, short circuit
						if (idx >= int(pt.SongLength())) {
							break
						}
						pattern, err := pt.GetPattern(patternNum)
						if (err == nil) {
							logger.Debugf("Sequence Number %d, Pattern: %d -> %d", idx, patternNum, pattern.NumChannels())
						}
					}
				} else {
					logger.Errorf("%s",err.Error())
				}
			},
		},
		{
			Name: "extract",
			Usage: "Extract an instrument sample.",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "file",
					Usage: "File to extract from.",
				},
				cli.StringFlag{
					Name: "out",
					Usage: "File to write sample to.",
				},
				cli.StringFlag{
					Name: "idx",
					Usage: "Instrument index to extract (starts at 0)",
				},
			},
			Action: func(c *cli.Context) {
				m, err := module.Load(c.String("file"))
				pt := m.(*module.ProTracker)
				if (err != nil) {
					logger.Errorf("%s",err.Error())
					os.Exit(1)
				}
				i,err := pt.GetInstrument(c.Int("idx"))
				if (err != nil) {
					logger.Errorf("%s",err.Error())
					os.Exit(1)
				}
				outdata := i.Data()
				err = ioutil.WriteFile(c.String("out"),outdata,0644)
				if (err != nil) {
					logger.Errorf("%s",err.Error())
					os.Exit(1)
				}
				logger.Infof("Wrote %d bytes to file %s",len(outdata),c.String("out"))
			},
		},
	}

	app.Run(os.Args)

}
