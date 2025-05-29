package storage

import (
	"os"
	"testing"
)

func TestNewFileStore(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "with custom path",
			path:     "./test.json",
			expected: "./test.json",
		},
		{
			name:     "with empty path",
			path:     "",
			expected: defaultPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFileStore(tt.path)
			if fs.Path != tt.expected {
				t.Errorf("NewFileStore() got = %v, want %v", fs.Path, tt.expected)
			}
		})
	}
}

func TestFileStore_SetAll_GetAll(t *testing.T) {
	// Create temporary file for testing
	tmpFile, err := os.CreateTemp("", "storage_test_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	fs := NewFileStore(tmpFile.Name())

	// Test data
	records := []Record{
		{Hostname: "host1", IP: "192.168.1.1", ID: "1"},
		{Hostname: "host2", IP: "192.168.1.2", ID: "2"},
	}

	// Test SetAll
	if err := fs.SetAll(records); err != nil {
		t.Errorf("SetAll() error = %v", err)
	}

	// Test GetAll
	got, err := fs.GetAll()
	if err != nil {
		t.Errorf("GetAll() error = %v", err)
	}

	if len(got) != len(records) {
		t.Errorf("GetAll() got %d records, want %d", len(got), len(records))
	}

	// Compare records
	for i, record := range records {
		if got[i].Hostname != record.Hostname ||
			got[i].IP != record.IP ||
			got[i].ID != record.ID {
			t.Errorf("GetAll() got = %v, want %v", got[i], record)
		}
	}
}

func TestFileStore_GetAll_EmptyFile(t *testing.T) {
	// Create empty temporary file
	tmpFile, err := os.CreateTemp("", "storage_test_empty_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	fs := NewFileStore(tmpFile.Name())

	// Test GetAll on empty file
	got, err := fs.GetAll()
	if err != nil {
		t.Errorf("GetAll() error = %v", err)
	}

	if len(got) != 0 {
		t.Errorf("GetAll() got %d records, want 0", len(got))
	}
}

func TestFileStore_GetAll_NonexistentFile(t *testing.T) {
	fs := NewFileStore("nonexistent.json")

	// Test GetAll on nonexistent file
	_, err := fs.GetAll()
	if err == nil {
		t.Error("GetAll() expected error for nonexistent file, got nil")
	}
}

func TestFileStore_DeleteAll(t *testing.T) {
	fs := NewFileStore("test.json")

	// Test DeleteAll
	if err := fs.DeleteAll(); err != nil {
		t.Errorf("DeleteAll() error = %v", err)
	}
}
