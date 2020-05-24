package main

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"go-mod/module"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

type DBConfig struct {
	host string
	name string
	user string
	pass string
	port string
}

func (d DBConfig) GetConnection() *sql.DB {

	mysqlInfo := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		d.user, d.pass, d.host, d.port, d.name)
	db, err := sql.Open("mysql", mysqlInfo)
	if err != nil {
		panic(err)
	}

	return db
}

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

func scanModForDB(inFile string, dbconn *sql.DB) error {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	log.Info("Loading file ", inFile)
	m, err := module.Load(inFile)
	samples := m.Samples()

	filename := path.Base(inFile)
	insert, err := dbconn.Prepare("REPLACE INTO modfile(title, filename) VALUES (?,?)")
	if err != nil {
		return err
	}
	insert.Exec(m.Title(), filename)
	insert.Close()

	// Delete any existing sample records
	delete, err := dbconn.Prepare("DELETE FROM sample WHERE modfile = ?")
	if err != nil {
		return err
	}
	delete.Exec(filename)
	delete.Close()

	// insert samples
	for i, sample := range samples {
		outdata := sample.Data()
		if len(outdata) == 0 {
			continue
		}
		hsh := sha256.New()
		hsh.Write(outdata)
		//checksum := sha256.Sum256(outdata)
		sChecksum := fmt.Sprintf("%x",hsh.Sum(nil))
		sampleInsert, err := dbconn.Prepare("INSERT INTO sample(name, filename, sha256, modfile, pos, len) VALUES (?,?,?,?,?,?)")
		if err != nil {
			return err
		}
		sampleInsert.Exec(sample.Name(), sample.Filename(), sChecksum, filename, int(i+1), len(sample.Data()))
		sampleInsert.Close()
	}

	return nil
}

func createDB(inPath string, dbConfig DBConfig) error {

	if !checkExists(inPath) {
		return errors.New("Input file does not exist")
	}

	log.WithFields(log.Fields{
		"inPath": inPath,
	}).Info("Building sample DB")

	dbconn := dbConfig.GetConnection()
	err := dbconn.Ping()
	if err != nil {
		return err
	}

	filesToScan := make([]string, 0)
	if info, err := os.Stat(inPath); err == nil && info.IsDir() {
		files, err := ioutil.ReadDir(inPath)
		if err != nil {
			return err
		}
		for _, f := range files {
			filesToScan = append(filesToScan, path.Join(inPath, f.Name()))
		}
	} else {
		filesToScan = append(filesToScan, inPath)
	}

	for _, inFile := range filesToScan {
		scanModForDB(inFile, dbconn)
	}

	dbconn.Close()
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
			&cli.Command{
				Name: "create-sample-db",
				Usage: "Create sample database.",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "file", Usage: "File to extract from." },
					&cli.StringFlag{Name: "dir", Usage: "Dir to search." },
					&cli.StringFlag{Name: "db-host", Usage: "Database host." },
					&cli.StringFlag{Name: "db-name", Usage: "Database name." },
					&cli.StringFlag{Name: "db-port", Usage: "Database port." },
					&cli.StringFlag{Name: "db-user", Usage: "Database user." },
					&cli.StringFlag{Name: "db-pass", Usage: "Database password." },
				},
				Action: func(c *cli.Context) error {
					f := c.String("file")
					dir := c.String("dir")
					dbHost := c.String("db-host")
					dbName := c.String("db-name")
					dbPort := c.String("db-port")
					dbUser := c.String("db-user")
					dbPass := c.String("db-pass")
					scanPath := f
					if dir != "" {
						scanPath = dir
					}
					return createDB(scanPath, DBConfig{ dbHost, dbName, dbUser, dbPass, dbPort})
				},
			},

		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
