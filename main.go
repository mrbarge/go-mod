package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"log/slog"

	"go-mod/module"
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
	slog.Info("Printing Info", "file", file)

	m, err := module.Load(file)
	if (err == nil) {
		slog.Info("Info", "title", m.Title(), "num-patterns", m.NumPatterns())
		for idx, sample := range m.Samples() {
			slog.Info("Sample",
				"index", idx,
				"name", sample.Name(),
				"filename", sample.Filename())
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
					slog.Info("Pattern",
						"seq-no", idx,
						"pattern", patternNum,
						"channel", pattern.NumChannels())
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
		return fmt.Errorf("input file does not exist: %s", infile)
	}
	if !checkExists(dir) {
		return fmt.Errorf("output directory does not exist: %s", dir)
	}

	slog.Info("Exporting all samples", "infile", infile, "dir", dir)

	m, err := module.Load(infile)
	if err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}
	samples := m.Samples()

	destdir := filepath.Join(dir, filepath.Base(infile))
	if err := os.MkdirAll(destdir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for idx, sample := range samples {
		outdata := sample.Data()
		if len(outdata) == 0 {
			slog.Info("Ignoring sample index", "index", idx)
			continue
		}
		destname := fmt.Sprintf("%d-%s", idx, stripRegex(sample.Filename()))
		outpath := filepath.Join(destdir, destname)

		if err := os.WriteFile(outpath, outdata, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", outpath, err)
		}
		slog.Info("Wrote", "num-bytes", len(outdata), "out-file", outpath)
	}

	return nil
}

func scanModForDB(inFile string, dbconn *sql.DB) error {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	slog.Info("Loading file", "file", inFile)
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
		return fmt.Errorf("input path does not exist: %s", inPath)
	}

	slog.Info("Building sample DB", "inPath", inPath)

	dbconn := dbConfig.GetConnection()
	defer dbconn.Close()

	if err := dbconn.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	var filesToScan []string
	if info, err := os.Stat(inPath); err == nil && info.IsDir() {
		entries, err := os.ReadDir(inPath)
		if err != nil {
			return fmt.Errorf("failed to read directory: %w", err)
		}
		for _, entry := range entries {
			filesToScan = append(filesToScan, path.Join(inPath, entry.Name()))
		}
	} else {
		filesToScan = append(filesToScan, inPath)
	}

	for _, inFile := range filesToScan {
		if err := scanModForDB(inFile, dbconn); err != nil {
			slog.Warn("Failed to scan file", "file", inFile, "error", err)
		}
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
	var rootCmd = &cobra.Command{
		Use:   "go-mod",
		Short: "A tool for working with MOD music files",
		Long:  "go-mod provides utilities for analyzing and extracting data from ProTracker, FastTracker, and other MOD format music files.",
	}

	// Info command
	var infoCmd = &cobra.Command{
		Use:   "info [file]",
		Short: "Print info about a MOD file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return info(args[0])
		},
	}

	// Dump samples command
	var dumpCmd = &cobra.Command{
		Use:   "dump-samples [file]",
		Short: "Dump all instrument samples from a MOD file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			if dir == "" {
				return fmt.Errorf("--dir flag is required")
			}
			return dumpAll(args[0], dir)
		},
	}
	dumpCmd.Flags().StringP("dir", "d", "", "Directory to write samples to (required)")
	dumpCmd.MarkFlagRequired("dir")

	// Create sample database command
	var dbCmd = &cobra.Command{
		Use:   "create-sample-db [path]",
		Short: "Create sample database from MOD file(s)",
		Long:  "Scan a MOD file or directory of MOD files and store sample information in a database.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbHost, _ := cmd.Flags().GetString("db-host")
			dbName, _ := cmd.Flags().GetString("db-name")
			dbPort, _ := cmd.Flags().GetString("db-port")
			dbUser, _ := cmd.Flags().GetString("db-user")
			dbPass, _ := cmd.Flags().GetString("db-pass")

			config := DBConfig{
				host: dbHost,
				name: dbName,
				user: dbUser,
				pass: dbPass,
				port: dbPort,
			}
			return createDB(args[0], config)
		},
	}
	dbCmd.Flags().String("db-host", "localhost", "Database host")
	dbCmd.Flags().String("db-name", "", "Database name (required)")
	dbCmd.Flags().String("db-port", "3306", "Database port")
	dbCmd.Flags().String("db-user", "", "Database user (required)")
	dbCmd.Flags().String("db-pass", "", "Database password")
	dbCmd.MarkFlagRequired("db-name")
	dbCmd.MarkFlagRequired("db-user")

	rootCmd.AddCommand(infoCmd, dumpCmd, dbCmd)

	if err := rootCmd.Execute(); err != nil {
		slog.Error("Command failed", "error", err)
		os.Exit(1)
	}
}
