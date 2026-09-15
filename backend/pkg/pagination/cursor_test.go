package pagination

import "testing"

func TestEncodeDecode_RoundTrip(t *testing.T) {
	token := Encode("0.42", 12345)

	decoded, err := Decode(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decoded.SortValue != "0.42" || decoded.ID != 12345 {
		t.Errorf("expected {0.42 12345}, got %+v", decoded)
	}
}

func TestDecode_EmptyStringIsNilCursor(t *testing.T) {
	decoded, err := Decode("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decoded != nil {
		t.Errorf("expected nil cursor for empty token, got %+v", decoded)
	}
}

func TestDecode_InvalidTokenErrors(t *testing.T) {
	if _, err := Decode("not-a-valid-cursor!!"); err == nil {
		t.Error("expected an error for a malformed cursor token")
	}
}

func TestNormalizeLimit(t *testing.T) {
	cases := map[int]int{
		0:    DefaultLimit,
		-5:   DefaultLimit,
		10:   10,
		1000: MaxLimit,
	}
	for input, want := range cases {
		if got := NormalizeLimit(input); got != want {
			t.Errorf("NormalizeLimit(%d) = %d, want %d", input, got, want)
		}
	}
}
