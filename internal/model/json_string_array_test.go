package model

import "testing"

func TestJSONStringArrayValueAndScan(t *testing.T) {
	original := JSONStringArray{"心动", "温柔", "春天"}

	value, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	var scanned JSONStringArray
	if err := scanned.Scan(value); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(scanned) != len(original) {
		t.Fatalf("expected %d tags, got %d", len(original), len(scanned))
	}

	for i := range original {
		if scanned[i] != original[i] {
			t.Fatalf("expected tag %q at index %d, got %q", original[i], i, scanned[i])
		}
	}
}

func TestJSONStringArrayNilValue(t *testing.T) {
	var tags JSONStringArray

	value, err := tags.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	if value != "[]" {
		t.Fatalf("expected empty json array, got %#v", value)
	}

	if err := tags.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) error = %v", err)
	}

	if len(tags) != 0 {
		t.Fatalf("expected empty tags after nil scan, got %d", len(tags))
	}
}
