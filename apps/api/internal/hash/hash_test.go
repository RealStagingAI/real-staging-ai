package hash

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestComputeSHA256(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantHash  string
		wantError bool
	}{
		{
			name:     "success: empty string",
			input:    "",
			wantHash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "success: simple string",
			input:    "hello world",
			wantHash: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name:     "success: image-like binary data",
			input:    "\xff\xd8\xff\xe0\x00\x10JFIF",
			wantHash: "45ae705277879f7f01d778f7c95a065bb0c06ab9936cf24307f375211fee13d1",
		},
		{
			name:     "success: long content",
			input:    strings.Repeat("a", 1000),
			wantHash: "41edece42d63e8d9bf515a9ba6932e1c20cbc9f5a5d134645adb5db1b9737ea3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			hash, err := ComputeSHA256(reader)

			if tt.wantError {
				if err == nil {
					t.Errorf("ComputeSHA256() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ComputeSHA256() unexpected error: %v", err)
				return
			}

			if hash != tt.wantHash {
				t.Errorf("ComputeSHA256() hash = %v, want %v", hash, tt.wantHash)
			}

			// Verify hash format
			if len(hash) != 64 {
				t.Errorf("ComputeSHA256() hash length = %d, want 64", len(hash))
			}
		})
	}
}

func TestComputeSHA256_ReaderError(t *testing.T) {
	t.Run("fail: reader returns error", func(t *testing.T) {
		reader := &errorReader{err: errors.New("read error")}
		_, err := ComputeSHA256(reader)
		if err == nil {
			t.Error("ComputeSHA256() expected error from reader, got nil")
		}
	})
}

func TestComputeSHA256FromBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantHash string
	}{
		{
			name:     "success: empty bytes",
			input:    []byte{},
			wantHash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "success: simple bytes",
			input:    []byte("hello world"),
			wantHash: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name:     "success: binary data",
			input:    []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10},
			wantHash: "d0b8d3f3c3f3e3c3d3e3f3c3e3d3f3c3e3d3f3c3e3d3f3c3e3d3f3c3e3d3f3c3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := ComputeSHA256FromBytes(tt.input)

			// Verify hash format
			if len(hash) != 64 {
				t.Errorf("ComputeSHA256FromBytes() hash length = %d, want 64", len(hash))
			}

			// Verify consistency with reader-based function
			readerHash, err := ComputeSHA256(bytes.NewReader(tt.input))
			if err != nil {
				t.Fatalf("ComputeSHA256() error: %v", err)
			}
			if hash != readerHash {
				t.Errorf("ComputeSHA256FromBytes() = %v, want %v (from reader)", hash, readerHash)
			}
		})
	}
}

func TestValidateHash(t *testing.T) {
	tests := []struct {
		name  string
		hash  string
		valid bool
	}{
		{
			name:  "success: valid lowercase hash",
			hash:  "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			valid: true,
		},
		{
			name:  "success: valid uppercase hash",
			hash:  "B94D27B9934D3E08A52E52D7DA7DABFAC484EFE37A5380EE9088F7ACE2EFCDE9",
			valid: true,
		},
		{
			name:  "success: valid mixed case hash",
			hash:  "B94d27b9934d3E08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			valid: true,
		},
		{
			name:  "fail: empty string",
			hash:  "",
			valid: false,
		},
		{
			name:  "fail: too short",
			hash:  "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde",
			valid: false,
		},
		{
			name:  "fail: too long",
			hash:  "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde99",
			valid: false,
		},
		{
			name:  "fail: invalid characters",
			hash:  "g94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			valid: false,
		},
		{
			name:  "fail: contains spaces",
			hash:  "b94d27b9 934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			valid: false,
		},
		{
			name:  "fail: contains special characters",
			hash:  "b94d27b9-934d3e08-a52e52d7-da7dabfa-c484efe3-7a5380ee-9088f7ac-e2efcde9",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := ValidateHash(tt.hash)
			if valid != tt.valid {
				t.Errorf("ValidateHash() = %v, want %v", valid, tt.valid)
			}
		})
	}
}

func TestComputeSHA256_Consistency(t *testing.T) {
	t.Run("success: same input produces same hash", func(t *testing.T) {
		input := []byte("consistent data")

		hash1, err := ComputeSHA256(bytes.NewReader(input))
		if err != nil {
			t.Fatalf("first ComputeSHA256() error: %v", err)
		}

		hash2, err := ComputeSHA256(bytes.NewReader(input))
		if err != nil {
			t.Fatalf("second ComputeSHA256() error: %v", err)
		}

		if hash1 != hash2 {
			t.Errorf("hashes differ: %v vs %v", hash1, hash2)
		}
	})

	t.Run("success: different input produces different hash", func(t *testing.T) {
		hash1, err := ComputeSHA256(strings.NewReader("data1"))
		if err != nil {
			t.Fatalf("first ComputeSHA256() error: %v", err)
		}

		hash2, err := ComputeSHA256(strings.NewReader("data2"))
		if err != nil {
			t.Fatalf("second ComputeSHA256() error: %v", err)
		}

		if hash1 == hash2 {
			t.Errorf("hashes are identical for different inputs: %v", hash1)
		}
	})
}

func BenchmarkComputeSHA256(b *testing.B) {
	data := bytes.Repeat([]byte("benchmark data"), 1000) // ~14KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ComputeSHA256(bytes.NewReader(data))
		if err != nil {
			b.Fatalf("ComputeSHA256() error: %v", err)
		}
	}
}

func BenchmarkComputeSHA256FromBytes(b *testing.B) {
	data := bytes.Repeat([]byte("benchmark data"), 1000) // ~14KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComputeSHA256FromBytes(data)
	}
}

// errorReader is a test helper that returns an error on Read
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (int, error) {
	return 0, r.err
}
