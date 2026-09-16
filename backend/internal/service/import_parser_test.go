package service

import (
	"strings"
	"testing"
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
