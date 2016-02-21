package main

import
(
	"go-mod/module"
	"log"
	"github.com/alexcesaro/log/stdlog"
	"os"
	"github.com/codegangsta/cli"
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
				} else {
					logger.Errorf("%s",err.Error())
				}
			},
		},
	}

	app.Run(os.Args)

}
