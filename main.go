package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
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

// JSON export structures for pattern data
type PatternExportNote struct {
	Note       string `json:"note"`
	Period     int    `json:"period"`
	Instrument int    `json:"instrument"`
	Effect     int    `json:"effect"`
	Parameter  int    `json:"parameter"`
}

type PatternExportRow struct {
	RowNumber int                  `json:"row"`
	Channels  []PatternExportNote  `json:"channels"`
}

type PatternExport struct {
	PatternNumber int                 `json:"pattern_number"`
	NumChannels   int                 `json:"num_channels"`
	NumRows       int                 `json:"num_rows"`
	Rows          []PatternExportRow  `json:"rows"`
}

type SampleExport struct {
	Number       int    `json:"number"`
	Name         string `json:"name"`
	Length       int    `json:"length"`
	Finetune     int    `json:"finetune"`
	Volume       int    `json:"volume"`
	RepeatOffset int    `json:"repeat_offset"`
	RepeatLength int    `json:"repeat_length"`
	Data         string `json:"data"` // Base64 encoded
}

// XM-specific structures
type XMSampleExport struct {
	Number       int    `json:"number"`
	Name         string `json:"name"`
	Length       uint32 `json:"length"`
	LoopStart    uint32 `json:"loop_start"`
	LoopEnd      uint32 `json:"loop_end"`
	Volume       uint8  `json:"volume"`
	Finetune     uint8  `json:"finetune"`
	SampleType   uint8  `json:"sample_type"`
	Panning      uint8  `json:"panning"`
	RelativeNote uint8  `json:"relative_note"`
	DataType     uint8  `json:"data_type"`
	Data         string `json:"data"` // Base64 encoded
}

type InstrumentExport struct {
	Number  int              `json:"number"`
	Name    string           `json:"name"`
	Samples []XMSampleExport `json:"samples"`
}

type ModulePatternExport struct {
	Format          string             `json:"format,omitempty"`       // "protracker" or "fasttracker"
	Title           string             `json:"title"`
	SongLength      int                `json:"song_length"`
	RestartPosition int                `json:"restart_position"`
	NumChannels     int                `json:"num_channels"`
	PatternOrder    []int              `json:"pattern_order"`
	Samples         []SampleExport     `json:"samples"`
	Patterns        []PatternExport    `json:"patterns"`
	// XM-specific fields (omitted for MOD files)
	Author      string              `json:"author,omitempty"`
	Version     uint16              `json:"version,omitempty"`
	Flags       uint16              `json:"flags,omitempty"`
	Tempo       uint16              `json:"tempo,omitempty"`
	BPM         uint16              `json:"bpm,omitempty"`
	Instruments []InstrumentExport  `json:"instruments,omitempty"`
}

func dumpPatterns(infile string, output string) error {
	if !checkExists(infile) {
		return fmt.Errorf("input file does not exist: %s", infile)
	}

	slog.Info("Loading module", "file", infile)
	m, err := module.Load(infile)
	if err != nil {
		return fmt.Errorf("failed to load module: %w", err)
	}

	var export ModulePatternExport

	// Handle different module formats
	switch m.Type() {
	case module.PROTRACKER:
		pt := m.(*module.ProTracker)

		// Build pattern order from sequence table
		patternOrder := make([]int, pt.SongLength())
		for i := 0; i < int(pt.SongLength()); i++ {
			patternOrder[i] = int(pt.SequenceTable()[i])
		}

		// Export samples
		samples := make([]SampleExport, 0)
		for idx, sample := range pt.Samples() {
			ptSample := sample.(module.PTSample)
			sampleExport := SampleExport{
				Number:       idx + 1,
				Name:         ptSample.Name(),
				Length:       int(ptSample.Length()),
				Finetune:     int(ptSample.Finetune()),
				Volume:       int(ptSample.Volume()),
				RepeatOffset: int(ptSample.RepeatOffset()),
				RepeatLength: int(ptSample.RepeatLength()),
				Data:         base64.StdEncoding.EncodeToString(ptSample.Data()),
			}
			samples = append(samples, sampleExport)
		}

		// Export all patterns
		patterns := make([]PatternExport, 0)
		for patNum := 0; patNum < pt.NumPatterns(); patNum++ {
			pattern, err := pt.GetPattern(int8(patNum))
			if err != nil {
				slog.Warn("Failed to get pattern", "pattern", patNum, "error", err)
				continue
			}

			patternExport := PatternExport{
				PatternNumber: patNum,
				NumChannels:   pattern.NumChannels(),
				NumRows:       pattern.NumRows(),
				Rows:          make([]PatternExportRow, 0),
			}

			for rowIdx := 0; rowIdx < pattern.NumRows(); rowIdx++ {
				row, err := pattern.GetRow(rowIdx)
				if err != nil {
					continue
				}

				notes := row.Notes()
				channels := make([]PatternExportNote, pattern.NumChannels())
				for chanIdx := 0; chanIdx < pattern.NumChannels(); chanIdx++ {
					note := notes[chanIdx]
					noteStr := "---"
					if note.Period() > 0 {
						if str, err := note.ToString(); err == nil {
							noteStr = str
						}
					}

					channels[chanIdx] = PatternExportNote{
						Note:       noteStr,
						Period:     note.Period(),
						Instrument: note.Instrument(),
						Effect:     note.Effect(),
						Parameter:  note.Parameter(),
					}
				}

				patternExport.Rows = append(patternExport.Rows, PatternExportRow{
					RowNumber: rowIdx,
					Channels:  channels,
				})
			}

			patterns = append(patterns, patternExport)
		}

		export = ModulePatternExport{
			Format:          "protracker",
			Title:           pt.Title(),
			SongLength:      int(pt.SongLength()),
			RestartPosition: pt.RestartPos(),
			NumChannels:     pt.NumChannels(),
			PatternOrder:    patternOrder,
			Samples:         samples,
			Patterns:        patterns,
		}

	case module.FASTTRACKER:
		ft := m.(*module.FastTracker)

		// Build pattern order from order table
		patternOrder := make([]int, int(ft.PatternSize()))
		for i := 0; i < int(ft.PatternSize()); i++ {
			patternOrder[i] = int(ft.OrderTable()[i])
		}

		// Export samples (flattened for backward compatibility)
		samples := make([]SampleExport, 0)
		for _, sample := range ft.Samples() {
			ftSample := sample.(module.FTSample)
			// Convert XM sample to MOD-style sample for compatibility
			sampleExport := SampleExport{
				Number:       len(samples) + 1,
				Name:         ftSample.Name(),
				Length:       int(ftSample.Length()),
				Finetune:     int(ftSample.Finetune()),
				Volume:       int(ftSample.Volume()),
				RepeatOffset: int(ftSample.LoopStart() / 2), // Convert bytes to words
				RepeatLength: int((ftSample.LoopEnd() - ftSample.LoopStart()) / 2),
				Data:         base64.StdEncoding.EncodeToString(ftSample.Data()),
			}
			samples = append(samples, sampleExport)
		}

		// Export instruments (XM hierarchical structure)
		instruments := make([]InstrumentExport, 0)
		for idx, inst := range ft.FTInstruments() {
			instSamples := make([]XMSampleExport, 0)
			for sIdx, sample := range inst.Samples() {
				xmSample := XMSampleExport{
					Number:       sIdx + 1,
					Name:         sample.Name(),
					Length:       sample.Length(),
					LoopStart:    sample.LoopStart(),
					LoopEnd:      sample.LoopEnd(),
					Volume:       sample.Volume(),
					Finetune:     sample.Finetune(),
					SampleType:   sample.SampleType(),
					Panning:      sample.Panning(),
					RelativeNote: sample.RelativeNote(),
					DataType:     sample.DataType(),
					Data:         base64.StdEncoding.EncodeToString(sample.Data()),
				}
				instSamples = append(instSamples, xmSample)
			}

			instExport := InstrumentExport{
				Number:  idx + 1,
				Name:    inst.Name(),
				Samples: instSamples,
			}
			instruments = append(instruments, instExport)
		}

		// Note: XM patterns aren't currently parsed, so patterns will be empty
		patterns := make([]PatternExport, 0)

		export = ModulePatternExport{
			Format:          "fasttracker",
			Title:           ft.Title(),
			SongLength:      int(ft.PatternSize()),
			RestartPosition: int(ft.RestartPosition()),
			NumChannels:     0, // TODO: Parse from XM header
			PatternOrder:    patternOrder,
			Samples:         samples,
			Patterns:        patterns,
			Author:          ft.Author(),
			Version:         ft.Version(),
			Flags:           ft.Flags(),
			Tempo:           ft.Tempo(),
			BPM:             ft.BPM(),
			Instruments:     instruments,
		}

	default:
		return fmt.Errorf("unsupported module format: %v", m.Type())
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file or stdout
	if output == "-" || output == "" {
		fmt.Println(string(jsonData))
	} else {
		if err := os.WriteFile(output, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		slog.Info("Pattern data exported", "output", output)
	}

	return nil
}

func importPatterns(jsonFile string, output string) error {
	if !checkExists(jsonFile) {
		return fmt.Errorf("input JSON file does not exist: %s", jsonFile)
	}

	slog.Info("Loading JSON file", "file", jsonFile)
	jsonData, err := os.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	var export ModulePatternExport
	if err := json.Unmarshal(jsonData, &export); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	slog.Info("Building MOD file", "title", export.Title)

	// Build the MOD file binary
	var modData []byte

	// First, decode all sample data to get actual lengths
	decodedSamples := make([][]byte, len(export.Samples))
	for i, sample := range export.Samples {
		if sample.Data != "" {
			sampleData, err := base64.StdEncoding.DecodeString(sample.Data)
			if err != nil {
				return fmt.Errorf("failed to decode sample %d data: %w", sample.Number, err)
			}
			decodedSamples[i] = sampleData
		}
	}

	// 1. Write title (20 bytes, padded with nulls)
	title := make([]byte, 20)
	copy(title, []byte(export.Title))
	modData = append(modData, title...)

	// 2. Write sample metadata (31 samples × 30 bytes each)
	for i := 0; i < 31; i++ {
		sampleMeta := make([]byte, 30)
		if i < len(export.Samples) {
			sample := export.Samples[i]
			// Name (22 bytes)
			copy(sampleMeta[0:22], []byte(sample.Name))
			// Length in words (2 bytes, big endian) - use actual decoded length
			actualLength := len(decodedSamples[i])
			lengthWords := uint16(actualLength / 2)
			sampleMeta[22] = byte(lengthWords >> 8)
			sampleMeta[23] = byte(lengthWords & 0xFF)
			// Finetune (1 byte)
			sampleMeta[24] = byte(sample.Finetune)
			// Volume (1 byte)
			sampleMeta[25] = byte(sample.Volume)
			// Repeat offset in words (2 bytes, big endian)
			sampleMeta[26] = byte(sample.RepeatOffset >> 8)
			sampleMeta[27] = byte(sample.RepeatOffset & 0xFF)
			// Repeat length in words (2 bytes, big endian)
			sampleMeta[28] = byte(sample.RepeatLength >> 8)
			sampleMeta[29] = byte(sample.RepeatLength & 0xFF)
		}
		modData = append(modData, sampleMeta...)
	}

	// 3. Write song length (1 byte)
	modData = append(modData, byte(export.SongLength))

	// 4. Write restart position (1 byte) - legacy, usually 127
	modData = append(modData, byte(export.RestartPosition))

	// 5. Write pattern order table (128 bytes)
	patternOrder := make([]byte, 128)
	for i := 0; i < len(export.PatternOrder) && i < 128; i++ {
		patternOrder[i] = byte(export.PatternOrder[i])
	}
	modData = append(modData, patternOrder...)

	// 6. Write magic number "M.K." for 4-channel MOD
	modData = append(modData, []byte("M.K.")...)

	// 7. Write pattern data
	for _, pattern := range export.Patterns {
		// Each pattern MUST be exactly 64 rows × numChannels × 4 bytes
		rowMap := make(map[int]PatternExportRow)
		for _, row := range pattern.Rows {
			rowMap[row.RowNumber] = row
		}

		for rowIdx := 0; rowIdx < 64; rowIdx++ {
			row, exists := rowMap[rowIdx]

			for chanIdx := 0; chanIdx < export.NumChannels; chanIdx++ {
				// Encode note as 4 bytes
				noteBytes := make([]byte, 4)

				if exists && chanIdx < len(row.Channels) {
					channel := row.Channels[chanIdx]

					// Byte 0: upper 4 bits of sample + upper 4 bits of period
					// Byte 1: lower 8 bits of period
					// Byte 2: lower 4 bits of sample + effect
					// Byte 3: effect parameter

					period := uint16(channel.Period)
					instrument := byte(channel.Instrument)
					effect := byte(channel.Effect)
					parameter := byte(channel.Parameter)

					noteBytes[0] = (instrument & 0xF0) | byte((period>>8)&0x0F)
					noteBytes[1] = byte(period & 0xFF)
					noteBytes[2] = ((instrument & 0x0F) << 4) | (effect & 0x0F)
					noteBytes[3] = parameter
				}
				// else: noteBytes remains all zeros (empty note)

				modData = append(modData, noteBytes...)
			}
		}
	}

	// 8. Write sample data
	for _, sampleData := range decodedSamples {
		if len(sampleData) > 0 {
			modData = append(modData, sampleData...)
		}
	}

	// Write to output file
	if err := os.WriteFile(output, modData, 0644); err != nil {
		return fmt.Errorf("failed to write MOD file: %w", err)
	}

	slog.Info("MOD file created", "output", output, "size", len(modData))
	return nil
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

	// Dump patterns command
	var dumpPatternsCmd = &cobra.Command{
		Use:   "dump-patterns [file]",
		Short: "Export pattern data to JSON format",
		Long:  "Export all pattern data including pattern order, notes, instruments, and effects to JSON format.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			output, _ := cmd.Flags().GetString("output")
			return dumpPatterns(args[0], output)
		},
	}
	dumpPatternsCmd.Flags().StringP("output", "o", "-", "Output file (use '-' for stdout)")

	// Import patterns command
	var importPatternsCmd = &cobra.Command{
		Use:   "import-patterns [json-file] [output-mod]",
		Short: "Recreate a MOD file from JSON format",
		Long:  "Import pattern and sample data from JSON and recreate the original MOD file.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return importPatterns(args[0], args[1])
		},
	}

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

	rootCmd.AddCommand(infoCmd, dumpCmd, dumpPatternsCmd, importPatternsCmd, dbCmd)

	if err := rootCmd.Execute(); err != nil {
		slog.Error("Command failed", "error", err)
		os.Exit(1)
	}
}
