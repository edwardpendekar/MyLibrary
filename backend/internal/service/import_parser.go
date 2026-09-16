package service

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// ImportRow is one parsed line of the Book/Chapter/Verse/text_en/text_id/title_en/title_id
// spreadsheet format the product spec defines. RowNumber is 1-based and counts the
// header row as row 1, matching what a spreadsheet user actually sees.
type ImportRow struct {
	RowNumber int
	Book      string
	Chapter   int
	Verse     int
	TextEN    string
	TextID    string
	TitleEN   string
	TitleID   string
}

type RowError struct {
	RowNumber int
	Message   string
}

// expectedHeaders maps the spec's required column names (case-insensitive) to
// their canonical field. Column order in the file does not matter.
var expectedHeaders = []string{"book", "chapter", "verse", "text_en", "text_id", "title_en", "title_id"}

// RowReader streams rows one at a time so a 1M+ row file is never fully
// materialized in memory. EstimateTotal is a cheap upper-bound row count (used
// for the import progress bar) computed without a full parse pass.
type RowReader interface {
	EstimateTotal() (int, error)
	Next() (*ImportRow, *RowError, bool, error) // (row, rowErr, hasMore, fatalErr)
	Close() error
}

func NewRowReader(ext string, r io.ReadSeeker) (RowReader, error) {
	switch strings.ToLower(ext) {
	case ".xlsx", ".xls":
		return newExcelRowReader(r)
	case ".csv":
		return newCSVRowReader(r)
	default:
		return nil, fmt.Errorf("unsupported file type %q: must be .xlsx, .xls, or .csv", ext)
	}
}

// --- Excel (xlsx/xls) ---

type excelRowReader struct {
	file      *excelize.File
	sheet     string
	rows      *excelize.Rows
	header    map[string]int
	rowNumber int
}

func newExcelRowReader(r io.ReadSeeker) (*excelRowReader, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("open spreadsheet: %w", err)
	}
	sheet := f.GetSheetName(0)
	if sheet == "" {
		return nil, fmt.Errorf("spreadsheet has no sheets")
	}

	// excelize.Rows is the streaming row iterator: it does not load the whole
	// sheet into memory, which is what makes a 1M+ row import feasible at all.
	rows, err := f.Rows(sheet)
	if err != nil {
		return nil, fmt.Errorf("open sheet rows: %w", err)
	}
	if !rows.Next() {
		return nil, fmt.Errorf("spreadsheet is empty")
	}
	headerCells, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("read header row: %w", err)
	}
	header, err := indexHeader(headerCells)
	if err != nil {
		return nil, err
	}

	return &excelRowReader{file: f, sheet: sheet, rows: rows, header: header, rowNumber: 1}, nil
}

func (e *excelRowReader) EstimateTotal() (int, error) {
	dim, err := e.file.GetSheetDimension(e.sheet)
	if err != nil || dim == "" {
		return 0, nil // unknown; progress bar falls back to a spinner
	}
	// dim looks like "A1:G1000001" — the trailing row number is a cheap upper bound.
	parts := strings.Split(dim, ":")
	if len(parts) != 2 {
		return 0, nil
	}
	rowDigits := strings.TrimLeft(parts[1], "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	total, err := strconv.Atoi(rowDigits)
	if err != nil {
		return 0, nil
	}
	if total > 0 {
		total-- // exclude header row
	}
	return total, nil
}

func (e *excelRowReader) Next() (*ImportRow, *RowError, bool, error) {
	if !e.rows.Next() {
		return nil, nil, false, e.rows.Error()
	}
	e.rowNumber++
	cells, err := e.rows.Columns()
	if err != nil {
		return nil, nil, true, fmt.Errorf("read row %d: %w", e.rowNumber, err)
	}
	row, rowErr := parseRow(e.rowNumber, cells, e.header)
	return row, rowErr, true, nil
}

func (e *excelRowReader) Close() error { return e.file.Close() }

// --- CSV ---

type csvRowReader struct {
	reader    *csv.Reader
	header    map[string]int
	total     int
	rowNumber int
}

func newCSVRowReader(r io.ReadSeeker) (*csvRowReader, error) {
	delimiter, err := detectCSVDelimiter(r)
	if err != nil {
		return nil, fmt.Errorf("detect delimiter: %w", err)
	}

	total, err := countCSVDataRows(r)
	if err != nil {
		return nil, err
	}
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("rewind file: %w", err)
	}

	decoded, err := decodeToUTF8(r)
	if err != nil {
		return nil, fmt.Errorf("detect encoding: %w", err)
	}

	cr := csv.NewReader(bufio.NewReaderSize(decoded, 64*1024))
	cr.Comma = delimiter
	cr.ReuseRecord = true
	headerCells, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("read header row: %w", err)
	}
	header, err := indexHeader(headerCells)
	if err != nil {
		return nil, err
	}

	return &csvRowReader{reader: cr, header: header, total: total, rowNumber: 1}, nil
}

// detectCSVDelimiter picks between comma and semicolon by counting which one
// appears more often in the header line. Semicolon-delimited CSV is the
// default export format for Excel under many non-US regional settings
// (where comma is already the decimal separator) — without this, those
// files fail with a cryptic "extraneous ... in quoted-field" error, since
// the whole header ends up parsed as one quoted field.
func detectCSVDelimiter(r io.ReadSeeker) (rune, error) {
	buf := make([]byte, 4096)
	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return ',', err
	}
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return ',', fmt.Errorf("rewind file: %w", err)
	}

	firstLine := buf[:n]
	if idx := bytes.IndexAny(firstLine, "\r\n"); idx != -1 {
		firstLine = firstLine[:idx]
	}
	if bytes.Count(firstLine, []byte{';'}) > bytes.Count(firstLine, []byte{','}) {
		return ';', nil
	}
	return ',', nil
}

// decodeToUTF8 transcodes the stream if it isn't already valid UTF-8. Real
// exports from older tools/Excel are routinely Windows-1252 — a file like
// that parses as CSV just fine (encoding/csv only cares about delimiters and
// quotes, not encoding) but then fails downstream with a cryptic Postgres
// "invalid byte sequence for encoding UTF8" the moment its text reaches a
// TEXT column, aborting the whole batch the offending row landed in. This
// only ever fires on files that genuinely aren't valid UTF-8, so a real
// UTF-8 file is never touched.
func decodeToUTF8(r io.ReadSeeker) (io.Reader, error) {
	sample := make([]byte, 32*1024)
	n, err := r.Read(sample)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("rewind file: %w", err)
	}

	// Drop the last few bytes before validating: a truncated multi-byte UTF-8
	// sequence right at the sample boundary would otherwise look invalid even
	// in a genuinely UTF-8 file.
	checkLen := n
	if checkLen > 4 {
		checkLen -= 4
	}
	if utf8.Valid(sample[:checkLen]) {
		return r, nil
	}
	return transform.NewReader(r, charmap.Windows1252.NewDecoder()), nil
}

func countCSVDataRows(r io.ReadSeeker) (int, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	count := 0
	for scanner.Scan() {
		count++
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("count rows: %w", err)
	}
	if count > 0 {
		count-- // header
	}
	return count, nil
}

func (c *csvRowReader) EstimateTotal() (int, error) { return c.total, nil }

func (c *csvRowReader) Next() (*ImportRow, *RowError, bool, error) {
	cells, err := c.reader.Read()
	if err == io.EOF {
		return nil, nil, false, nil
	}
	if err != nil {
		return nil, nil, true, fmt.Errorf("read row %d: %w", c.rowNumber+1, err)
	}
	c.rowNumber++
	row, rowErr := parseRow(c.rowNumber, cells, c.header)
	return row, rowErr, true, nil
}

func (c *csvRowReader) Close() error { return nil }

// --- shared parsing helpers ---

func indexHeader(cells []string) (map[string]int, error) {
	index := make(map[string]int, len(cells))
	for i, cell := range cells {
		index[strings.ToLower(strings.TrimSpace(cell))] = i
	}
	var missing []string
	for _, h := range expectedHeaders {
		if _, ok := index[h]; !ok {
			missing = append(missing, h)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required column(s): %s", strings.Join(missing, ", "))
	}
	return index, nil
}

func cell(cells []string, header map[string]int, name string) string {
	idx, ok := header[name]
	if !ok || idx >= len(cells) {
		return ""
	}
	// Belt-and-braces: even after CSV encoding detection (or for .xlsx, which
	// excelize already guarantees is UTF-8), strip any byte sequence that
	// still isn't valid UTF-8 rather than let it reach Postgres, which
	// rejects the entire batch — not just the one bad cell — on this.
	return strings.TrimSpace(strings.ToValidUTF8(cells[idx], ""))
}

func parseRow(rowNumber int, cells []string, header map[string]int) (*ImportRow, *RowError) {
	book := cell(cells, header, "book")
	if book == "" {
		return nil, &RowError{RowNumber: rowNumber, Message: "book is required"}
	}
	chapter, err := strconv.Atoi(cell(cells, header, "chapter"))
	if err != nil || chapter <= 0 {
		return nil, &RowError{RowNumber: rowNumber, Message: "chapter must be a positive integer"}
	}
	verse, err := strconv.Atoi(cell(cells, header, "verse"))
	if err != nil || verse <= 0 {
		return nil, &RowError{RowNumber: rowNumber, Message: "verse must be a positive integer"}
	}

	return &ImportRow{
		RowNumber: rowNumber, Book: book, Chapter: chapter, Verse: verse,
		TextEN: cell(cells, header, "text_en"), TextID: cell(cells, header, "text_id"),
		TitleEN: cell(cells, header, "title_en"), TitleID: cell(cells, header, "title_id"),
	}, nil
}
