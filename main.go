package main

import
(
	"go-mod/module"
	"log"
	"github.com/alexcesaro/log/stdlog"
	"flag"
	"os"
)


func main() {
	modfilepath := flag.String("load","","Path to file to load")
	logger := stdlog.GetFromFlags()
	flag.Parse()
	if (*modfilepath == "") {
		logger.Error("Usage:")
		logger.Error("   main.go -load <modfile>")
		os.Exit(1)
	}

	m, err := module.Load(*modfilepath)
	pt := m.(*module.ProTracker)
	if (err == nil) {
		logger.Infof("%s",pt.Title())
		logger.Infof("song length %d",pt.SongLength())
		logger.Infof("Num patterns: %d",pt.NumPatterns())
	} else {
		log.Fatal(err)
	}
}
