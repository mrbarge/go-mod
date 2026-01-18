# ProTracker MOD JSON Format Specification

This document describes the JSON format used for importing and exporting ProTracker MOD files. This format provides a human-readable and programmatically accessible representation of MOD file data.

## Overview

The JSON format represents a complete ProTracker module including:
- Module metadata (title, song structure)
- Sample data (instruments with audio data)
- Pattern data (musical notation and effects)

## Root Structure

```json
{
  "title": "string",
  "song_length": number,
  "restart_position": number,
  "num_channels": number,
  "pattern_order": [array of numbers],
  "samples": [array of SampleExport objects],
  "patterns": [array of PatternExport objects]
}
```

### Root Fields

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Module title (max 20 characters, will be truncated/padded) |
| `song_length` | number | Number of positions in the pattern order table (1-128) |
| `restart_position` | number | Position to restart playback (legacy field, typically 0) |
| `num_channels` | number | Number of channels (typically 4 for standard ProTracker) |
| `pattern_order` | number[] | Array of pattern indices defining song structure (max 128 entries) |
| `samples` | SampleExport[] | Array of sample/instrument definitions (max 31) |
| `patterns` | PatternExport[] | Array of pattern definitions |

## Sample Structure

Each sample represents an instrument with its audio data and playback parameters.

```json
{
  "number": number,
  "name": "string",
  "length": number,
  "finetune": number,
  "volume": number,
  "repeat_offset": number,
  "repeat_length": number,
  "data": "base64-encoded-string"
}
```

### Sample Fields

| Field | Type | Range | Description |
|-------|------|-------|-------------|
| `number` | number | 1-31 | Sample number (1-based index) |
| `name` | string | max 22 chars | Sample/instrument name |
| `length` | number | 0-131070 | Sample length in bytes (stored as words in MOD) |
| `finetune` | number | -8 to 7 | Fine-tuning value for pitch adjustment |
| `volume` | number | 0-64 | Default playback volume |
| `repeat_offset` | number | 0-65535 | Loop start offset in words (multiply by 2 for bytes) |
| `repeat_length` | number | 1-65535 | Loop length in words (1 = no loop, >1 = looping sample) |
| `data` | string | base64 | Sample audio data encoded as base64 |

### Sample Notes

- Sample data should be 8-bit signed mono PCM audio
- Sample lengths are always even (stored in 2-byte words)
- If `repeat_length` is 1, the sample doesn't loop
- If `repeat_length` > 1, the sample loops from `repeat_offset` for `repeat_length` words
- When importing, the actual decoded base64 data length is used, not the `length` field

## Pattern Structure

Each pattern represents 64 rows of musical data across all channels.

```json
{
  "pattern_number": number,
  "num_channels": number,
  "num_rows": number,
  "rows": [array of PatternExportRow objects]
}
```

### Pattern Fields

| Field | Type | Description |
|-------|------|-------------|
| `pattern_number` | number | Pattern index (0-based) |
| `num_channels` | number | Number of channels in this pattern (typically 4) |
| `num_rows` | number | Number of rows (always 64 for ProTracker) |
| `rows` | PatternExportRow[] | Array of row data |

### Pattern Notes

- ProTracker patterns MUST have exactly 64 rows
- When importing, missing rows are automatically filled with empty notes
- Rows can be sparse in the JSON (only non-empty rows need to be specified)

## Row Structure

Each row represents one step in the pattern across all channels.

```json
{
  "row": number,
  "channels": [array of PatternExportNote objects]
}
```

### Row Fields

| Field | Type | Description |
|-------|------|-------------|
| `row` | number | Row number (0-63) |
| `channels` | PatternExportNote[] | Array of note data for each channel |

## Note Structure

Each note represents the musical data for one channel at one row.

```json
{
  "note": "string",
  "period": number,
  "instrument": number,
  "effect": number,
  "parameter": number
}
```

### Note Fields

| Field | Type | Range | Description |
|-------|------|-------|-------------|
| `note` | string | - | Human-readable note (e.g., "C-2", "A#3", "---" for empty) |
| `period` | number | 0-4095 | Amiga period value (0 = no note) |
| `instrument` | number | 0-31 | Instrument/sample number (0 = no instrument change) |
| `effect` | number | 0-15 | Effect type (0x0-0xF in hex) |
| `parameter` | number | 0-255 | Effect parameter value |

### Note Period Values

Common note periods for standard ProTracker tuning (finetune = 0):

| Note | Oct 0 | Oct 1 | Oct 2 | Oct 3 | Oct 4 |
|------|-------|-------|-------|-------|-------|
| C    | 1712  | 856   | 428   | 214   | 107   |
| C#   | 1616  | 808   | 404   | 202   | 101   |
| D    | 1524  | 762   | 381   | 190   | 95    |
| D#   | 1440  | 720   | 360   | 180   | 90    |
| E    | 1356  | 678   | 339   | 170   | 85    |
| F    | 1280  | 640   | 320   | 160   | 80    |
| F#   | 1208  | 604   | 302   | 151   | 75    |
| G    | 1140  | 570   | 285   | 143   | 71    |
| G#   | 1076  | 538   | 269   | 135   | 67    |
| A    | 1016  | 508   | 254   | 127   | 63    |
| A#   | 960   | 480   | 240   | 120   | 60    |
| B    | 907   | 453   | 226   | 113   | 56    |

### Common Effects

| Effect | Name | Description |
|--------|------|-------------|
| 0 | Arpeggio | Rapid note cycling |
| 1 | Slide Up | Increase pitch |
| 2 | Slide Down | Decrease pitch |
| 3 | Tone Portamento | Slide to target note |
| 4 | Vibrato | Pitch oscillation |
| 5 | Continue Portamento + Volume Slide | Combined effect |
| 6 | Continue Vibrato + Volume Slide | Combined effect |
| 7 | Tremolo | Volume oscillation |
| 8 | Set Panning | Set stereo position |
| 9 | Set Sample Offset | Start sample at offset |
| A | Volume Slide | Increase/decrease volume |
| B | Position Jump | Jump to pattern |
| C | Set Volume | Set channel volume |
| D | Pattern Break | Jump to next pattern |
| E | Extended Effects | Extended command set |
| F | Set Speed/Tempo | Set playback speed |

## Example JSON

```json
{
  "title": "Example Song",
  "song_length": 2,
  "restart_position": 0,
  "num_channels": 4,
  "pattern_order": [0, 1],
  "samples": [
    {
      "number": 1,
      "name": "Kick",
      "length": 1000,
      "finetune": 0,
      "volume": 64,
      "repeat_offset": 0,
      "repeat_length": 1,
      "data": "AP8A/wD/AP8A..."
    }
  ],
  "patterns": [
    {
      "pattern_number": 0,
      "num_channels": 4,
      "num_rows": 64,
      "rows": [
        {
          "row": 0,
          "channels": [
            {
              "note": "C-2",
              "period": 428,
              "instrument": 1,
              "effect": 0,
              "parameter": 0
            },
            {
              "note": "---",
              "period": 0,
              "instrument": 0,
              "effect": 0,
              "parameter": 0
            },
            {
              "note": "---",
              "period": 0,
              "instrument": 0,
              "effect": 0,
              "parameter": 0
            },
            {
              "note": "---",
              "period": 0,
              "instrument": 0,
              "effect": 0,
              "parameter": 0
            }
          ]
        }
      ]
    }
  ]
}
```

## Import/Export Commands

### Export MOD to JSON
```bash
./go-mod dump-patterns input.mod -o output.json
```

### Import JSON to MOD
```bash
./go-mod import-patterns input.json output.mod
```

## Technical Notes

### Binary Encoding
- ProTracker MOD files use big-endian byte order
- Sample lengths and repeat values are stored in words (2-byte units)
- Note data is packed into 4 bytes per note

### Round-Trip Compatibility
- Export and import operations are designed to be lossless
- Round-trip conversion (MOD → JSON → MOD) produces byte-identical files
- Sample data with odd byte lengths will be truncated to even lengths during import

### Sparse Row Data
When exporting, all 64 rows are included. When importing:
- Only specified rows need to be in the JSON
- Missing rows are filled with empty notes (all zeros)
- This allows for compact representation of sparse patterns

### Sample Data Format
- Audio data must be base64-encoded in the JSON
- Decoded data should be 8-bit signed PCM
- Sample rate is typically 8363 Hz (middle C period)
- Mono audio only

## Version Information

This format is compatible with:
- ProTracker 2.3 MOD files (4-channel)
- ProTracker M.K. format
- Extended formats (6CHN, 8CHN, FLT4, FLT8)

Generated by: go-mod tool
Format Version: 1.0
Last Updated: 2026-01-18
