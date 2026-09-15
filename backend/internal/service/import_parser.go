package service

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
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
	total, err := countCSVDataRows(r)
	if err != nil {
		return nil, err
	}
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("rewind file: %w", err)
	}

	cr := csv.NewReader(bufio.NewReaderSize(r, 64*1024))
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
	return strings.TrimSpace(cells[idx])
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
