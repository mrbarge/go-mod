package main

import
(
	"go-mod/module"
	"log"
	"github.com/alexcesaro/log/stdlog"
)


func main() {

	logger := stdlog.GetFromFlags()
	m, err := module.Load("F:/workspace/go/src/go-mod/test/Origin.mod")
	pt := m.(*module.ProTracker)
	if (err == nil) {
		logger.Infof("%s",pt.Title())
		logger.Infof("song length %d",pt.SongLength())
		logger.Infof("Num patterns: %d",pt.NumPatterns())
	} else {
		log.Fatal(err)
	}
	/*
	m, err = module.Load("F:/workspace/go/src/go-mod/test/Strange.mod")
	fmt.Printf("%s\n",m.Title())
	m, err = module.Load("F:/workspace/go/src/go-mod/test/TestModFive.mod")
	fmt.Printf("%s\n",m.Title())
	m, err = module.Load("F:/workspace/go/src/go-mod/test/PlanetSized.mod")
	fmt.Printf("%s\n",m.Title())
	m, err = module.Load("F:/workspace/go/src/go-mod/test/dreamzon.xm")
	fmt.Printf("%s\n",m.Title())
	m, err = module.Load("F:/workspace/go/src/go-mod/test/realization.s3m")
	fmt.Printf("%s\n",m.Title())
*/
}
