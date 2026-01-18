# Module JSON Format Specification

This document describes the JSON format used for importing and exporting tracker module files. This format provides a human-readable and programmatically accessible representation of module data.

Supported formats:
- **ProTracker MOD** - Classic 4-channel Amiga tracker format
- **FastTracker XM** - Extended multi-channel format with advanced features

## Overview

The JSON format represents a complete tracker module including:
- Module metadata (title, song structure, format-specific properties)
- Sample/Instrument data (audio data and playback parameters)
- Pattern data (musical notation and effects)

The format uses optional fields to support both ProTracker MOD and FastTracker XM while maintaining backward compatibility.

## Root Structure

### ProTracker MOD Format
```json
{
  "format": "protracker",
  "title": "string",
  "song_length": number,
  "restart_position": number,
  "num_channels": number,
  "pattern_order": [array of numbers],
  "samples": [array of SampleExport objects],
  "patterns": [array of PatternExport objects]
}
```

### FastTracker XM Format
```json
{
  "format": "fasttracker",
  "title": "string",
  "song_length": number,
  "restart_position": number,
  "num_channels": number,
  "pattern_order": [array of numbers],
  "samples": [array of SampleExport objects],
  "patterns": [array of PatternExport objects],
  "author": "string",
  "version": number,
  "flags": number,
  "tempo": number,
  "bpm": number,
  "instruments": [array of InstrumentExport objects]
}
```

### Root Fields

#### Common Fields (Both Formats)

| Field | Type | Description |
|-------|------|-------------|
| `format` | string | Module format: "protracker" or "fasttracker" |
| `title` | string | Module title (max 20 characters for MOD, 20 for XM) |
| `song_length` | number | Number of positions in the pattern order table (1-128 for MOD, 1-256 for XM) |
| `restart_position` | number | Position to restart playback |
| `num_channels` | number | Number of channels (4 for standard MOD, 1-32 for XM) |
| `pattern_order` | number[] | Array of pattern indices defining song structure |
| `samples` | SampleExport[] | Array of sample definitions (max 31 for MOD, flattened from instruments for XM) |
| `patterns` | PatternExport[] | Array of pattern definitions |

#### XM-Specific Fields (Optional)

These fields only appear when `format` is "fasttracker":

| Field | Type | Description |
|-------|------|-------------|
| `author` | string | Module author name (max 20 characters) |
| `version` | number | XM format version number (e.g., 0x0104 for v1.04) |
| `flags` | number | Module flags (bit 0: Amiga frequency table) |
| `tempo` | number | Default tempo (ticks per row, typically 6) |
| `bpm` | number | Default BPM (beats per minute, typically 125) |
| `instruments` | InstrumentExport[] | Hierarchical instrument structures (XM-specific) |

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

- **ProTracker MOD**: Sample data should be 8-bit signed mono PCM audio
- **ProTracker MOD**: Sample lengths are always even (stored in 2-byte words)
- **ProTracker MOD**: If `repeat_length` is 1, the sample doesn't loop
- **ProTracker MOD**: If `repeat_length` > 1, the sample loops from `repeat_offset` for `repeat_length` words
- **FastTracker XM**: Samples are exported in a flattened format for backward compatibility
- When importing, the actual decoded base64 data length is used, not the `length` field

## XM Instrument Structure

For FastTracker XM files, the `instruments` array contains hierarchical instrument data. Each instrument can contain multiple samples with extended parameters.

```json
{
  "number": number,
  "name": "string",
  "samples": [array of XMSampleExport objects]
}
```

### Instrument Fields

| Field | Type | Description |
|-------|------|-------------|
| `number` | number | Instrument number (1-based index, max 128) |
| `name` | string | Instrument name (max 22 characters) |
| `samples` | XMSampleExport[] | Array of samples within this instrument |

## XM Sample Structure

XM samples have extended properties compared to MOD samples.

```json
{
  "number": number,
  "name": "string",
  "length": number,
  "loop_start": number,
  "loop_end": number,
  "volume": number,
  "finetune": number,
  "sample_type": number,
  "panning": number,
  "relative_note": number,
  "data_type": number,
  "data": "base64-encoded-string"
}
```

### XM Sample Fields

| Field | Type | Range | Description |
|-------|------|-------|-------------|
| `number` | number | 1-16 | Sample number within instrument (1-based) |
| `name` | string | max 22 chars | Sample name |
| `length` | number | 0-4GB | Sample length in bytes |
| `loop_start` | number | 0-length | Loop start position in bytes |
| `loop_end` | number | 0-length | Loop end position in bytes |
| `volume` | number | 0-64 | Default playback volume |
| `finetune` | number | -128 to 127 | Fine-tuning value (signed byte) |
| `sample_type` | number | 0-3 | Bit 0: loop on/off, Bit 1: ping-pong loop |
| `panning` | number | 0-255 | Default panning (0=left, 128=center, 255=right) |
| `relative_note` | number | -96 to 95 | Relative note (transpose, signed) |
| `data_type` | number | 0-1 | 0=8-bit, 1=16-bit sample data |
| `data` | string | base64 | Sample audio data encoded as base64 |

### XM Sample Notes

- XM samples can be 8-bit or 16-bit (check `data_type`)
- Loop positions are in bytes, not words like MOD
- `sample_type` bit flags: `0x01` = forward loop, `0x02` = ping-pong loop
- XM samples are stored as delta values in the file but are decoded to absolute values in the JSON
- Samples support advanced features like ping-pong loops and per-sample panning

## Pattern Structure

Each pattern represents musical data across all channels. ProTracker MOD patterns always have 64 rows, while FastTracker XM patterns can have 1-256 rows.

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

- **ProTracker MOD**: Patterns MUST have exactly 64 rows
- **FastTracker XM**: Patterns can have 1-256 rows (variable per pattern)
- **Note**: XM pattern parsing is not yet fully implemented in the current version
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

### Export to JSON
```bash
# Export ProTracker MOD to JSON
./go-mod dump-patterns input.mod -o output.json

# Export FastTracker XM to JSON
./go-mod dump-patterns input.xm -o output.json
```

### Import from JSON
```bash
# Import JSON to ProTracker MOD
./go-mod import-patterns input.json output.mod

# Import JSON to FastTracker XM (not yet implemented)
# ./go-mod import-patterns input.json output.xm
```

**Note**: XM import functionality is not yet implemented in the current version.

## Technical Notes

### Binary Encoding

#### ProTracker MOD
- Uses big-endian byte order
- Sample lengths and repeat values are stored in words (2-byte units)
- Note data is packed into 4 bytes per note

#### FastTracker XM
- Uses little-endian byte order
- Sample data stored as delta-encoded values (decoded during load)
- Note data is 5 bytes per note with pattern packing
- Supports 8-bit and 16-bit sample data

### Round-Trip Compatibility

#### ProTracker MOD
- Export and import operations are fully lossless
- Round-trip conversion (MOD → JSON → MOD) produces byte-identical files
- Sample data with odd byte lengths will be truncated to even lengths during import

#### FastTracker XM
- Export is functional but patterns are not yet parsed
- Import functionality not yet implemented
- XM samples are decoded from delta format to absolute values in JSON

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
- **ProTracker MOD**:
  - ProTracker 2.3 MOD files (4-channel)
  - ProTracker M.K. format
  - Extended formats (6CHN, 8CHN, FLT4, FLT8)
- **FastTracker XM**:
  - FastTracker II XM files (v1.04)
  - Multi-channel modules (1-32 channels)
  - Variable pattern lengths (1-256 rows)

Generated by: go-mod tool
Format Version: 1.1 (added XM support)
Last Updated: 2026-01-18

## Implementation Status

### Completed
- ✅ ProTracker MOD export to JSON (fully functional)
- ✅ ProTracker MOD import from JSON (fully functional)
- ✅ FastTracker XM metadata export (title, author, version, tempo, BPM)
- ✅ FastTracker XM instrument/sample export (hierarchical and flattened)
- ✅ Backward compatibility (MOD files unchanged by XM additions)

### In Progress / TODO
- ⏳ FastTracker XM pattern parsing (currently patterns export as empty array)
- ⏳ FastTracker XM import from JSON
- ⏳ XM envelope data export/import
- ⏳ XM note encoding/decoding (5-byte compressed format)
