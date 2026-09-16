package service

import (
	"strings"
	"testing"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func TestCSVRowReader_ParsesValidRows(t *testing.T) {
	csv := "book,chapter,verse,text_en,text_id,title_en,title_id\n" +
		"Genesis,1,1,In the beginning,Pada mulanya,Creation,Penciptaan\n" +
		"Genesis,1,2,The earth was,Bumi belum,Creation,Penciptaan\n"

	reader, err := NewRowReader(".csv", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer reader.Close()

	if total, _ := reader.EstimateTotal(); total != 2 {
		t.Errorf("expected EstimateTotal()=2, got %d", total)
	}

	row, rowErr, hasMore, err := reader.Next()
	if err != nil || rowErr != nil || !hasMore {
		t.Fatalf("unexpected first row result: row=%v rowErr=%v hasMore=%v err=%v", row, rowErr, hasMore, err)
	}
	if row.Book != "Genesis" || row.Chapter != 1 || row.Verse != 1 {
		t.Errorf("unexpected row values: %+v", row)
	}
	if row.TitleEN != "Creation" {
		t.Errorf("expected title_en=Creation, got %q", row.TitleEN)
	}

	_, _, hasMore, _ = reader.Next()
	if !hasMore {
		t.Fatal("expected a second row")
	}

	_, _, hasMore, _ = reader.Next()
	if hasMore {
		t.Fatal("expected no more rows after the second one")
	}
}

func TestCSVRowReader_MissingRequiredColumn(t *testing.T) {
	// Missing "verse" column entirely.
	csv := "book,chapter,text_en,text_id,title_en,title_id\nGenesis,1,text,teks,t,j\n"

	_, err := NewRowReader(".csv", strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected an error for a spreadsheet missing the required 'verse' column")
	}
}

func TestCSVRowReader_InvalidChapterOrVerseIsReportedAsRowError(t *testing.T) {
	csv := "book,chapter,verse,text_en,text_id,title_en,title_id\n" +
		"Genesis,notanumber,1,text,teks,t,j\n" +
		"Genesis,1,0,text,teks,t,j\n" + // verse must be > 0
		"Genesis,1,1,text,teks,t,j\n" // valid row

	reader, err := NewRowReader(".csv", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer reader.Close()

	row, rowErr, hasMore, err := reader.Next()
	if err != nil || !hasMore {
		t.Fatalf("unexpected result: err=%v hasMore=%v", err, hasMore)
	}
	if rowErr == nil || row != nil {
		t.Errorf("expected row 1 (non-numeric chapter) to be a row error, got row=%+v rowErr=%v", row, rowErr)
	}

	row, rowErr, hasMore, err = reader.Next()
	if err != nil || !hasMore {
		t.Fatalf("unexpected result: err=%v hasMore=%v", err, hasMore)
	}
	if rowErr == nil || row != nil {
		t.Errorf("expected row 2 (verse=0) to be a row error, got row=%+v rowErr=%v", row, rowErr)
	}

	row, rowErr, hasMore, err = reader.Next()
	if err != nil || !hasMore || rowErr != nil {
		t.Fatalf("expected row 3 to parse cleanly, got row=%+v rowErr=%v hasMore=%v err=%v", row, rowErr, hasMore, err)
	}
	if row.Chapter != 1 || row.Verse != 1 {
		t.Errorf("unexpected valid row values: %+v", row)
	}
}

func TestCSVRowReader_MissingBookIsRowError(t *testing.T) {
	csv := "book,chapter,verse,text_en,text_id,title_en,title_id\n,1,1,text,teks,t,j\n"

	reader, err := NewRowReader(".csv", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer reader.Close()

	row, rowErr, hasMore, err := reader.Next()
	if err != nil || !hasMore {
		t.Fatalf("unexpected result: err=%v hasMore=%v", err, hasMore)
	}
	if rowErr == nil || row != nil {
		t.Errorf("expected empty book to be a row error, got row=%+v rowErr=%v", row, rowErr)
	}
}

// Semicolon-delimited CSV is Excel's default export under many non-US
// regional settings (e.g. Indonesian, German) — a real file in this format
// was reported failing with a cryptic "extraneous ... in quoted-field" error
// before delimiter auto-detection was added.
func TestCSVRowReader_SemicolonDelimited(t *testing.T) {
	csv := "\"id\";\"book\";\"chapter\";\"verse\";\"text_en\";\"text_id\";\"title_en\";\"title_id\"\r\n" +
		"\"10\";\"Irenaeus 4\";\"1\";\"1\";\"Preface, part one\";\"Kata pengantar\";\"Intro\";\"Pendahuluan\"\r\n"

	reader, err := NewRowReader(".csv", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer reader.Close()

	row, rowErr, hasMore, err := reader.Next()
	if err != nil || rowErr != nil || !hasMore {
		t.Fatalf("unexpected result: row=%v rowErr=%v hasMore=%v err=%v", row, rowErr, hasMore, err)
	}
	if row.Book != "Irenaeus 4" || row.Chapter != 1 || row.Verse != 1 {
		t.Errorf("unexpected row values: %+v", row)
	}
	// A comma inside a quoted field must survive untouched now that ';' is
	// the actual field separator, not be mistaken for another delimiter.
	if row.TextEN != "Preface, part one" {
		t.Errorf("expected comma-containing text to be preserved, got %q", row.TextEN)
	}
}

func TestCSVRowReader_CommaDelimitedStillWorksAlongsideSemicolon(t *testing.T) {
	csv := "book,chapter,verse,text_en,text_id,title_en,title_id\n" +
		"Genesis,1,1,In the beginning,Pada mulanya,Creation,Penciptaan\n"

	reader, err := NewRowReader(".csv", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer reader.Close()

	row, rowErr, hasMore, err := reader.Next()
	if err != nil || rowErr != nil || !hasMore {
		t.Fatalf("unexpected result: row=%v rowErr=%v hasMore=%v err=%v", row, rowErr, hasMore, err)
	}
	if row.Book != "Genesis" {
		t.Errorf("expected comma-delimited files to still parse, got %+v", row)
	}
}

// Reproduces the exact reported failure: a Windows-1252-encoded CSV (Excel's
// default under many non-US locales) containing an em dash (byte 0x97 in
// CP1252) parses as CSV without error, but Postgres rejects that raw byte
// as invalid UTF-8 the moment it reaches a TEXT column — aborting the whole
// batch the row landed in, not just that row.
func TestCSVRowReader_Windows1252IsTranscodedToUTF8(t *testing.T) {
	utf8Text := "one thing—another thing"
	cp1252Bytes, _, err := transform.String(charmap.Windows1252.NewEncoder(), utf8Text)
	if err != nil {
		t.Fatalf("failed to build CP1252 fixture: %v", err)
	}

	csvContent := "book,chapter,verse,text_en,text_id,title_en,title_id\n" +
		"Genesis,1,1," + cp1252Bytes + ",teks,judul,title\n"

	if utf8.ValidString(csvContent) {
		t.Fatal("test fixture must actually be invalid UTF-8 to exercise the fallback")
	}

	reader, err := NewRowReader(".csv", strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer reader.Close()

	row, rowErr, hasMore, err := reader.Next()
	if err != nil || rowErr != nil || !hasMore {
		t.Fatalf("unexpected result: row=%v rowErr=%v hasMore=%v err=%v", row, rowErr, hasMore, err)
	}
	if !utf8.ValidString(row.TextEN) {
		t.Fatalf("expected transcoded text to be valid UTF-8, got %q", row.TextEN)
	}
	if row.TextEN != utf8Text {
		t.Errorf("expected %q, got %q", utf8Text, row.TextEN)
	}
}
