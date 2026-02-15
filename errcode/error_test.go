package errcode

import (
	"errors"
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name: "reason and message",
			err: &Error{
				code:   "1014040001",
				reason: "NOT_FOUND",
				msg:    "Resource not found",
			},
			expected: "[NOT_FOUND:1014040001] Resource not found",
		},
		{
			name: "reason only",
			err: &Error{
				code:   "1014040001",
				reason: "NOT_FOUND",
			},
			expected: "[NOT_FOUND:1014040001]",
		},
		{
			name: "message only",
			err: &Error{
				code: "1014040001",
				msg:  "Resource not found",
			},
			expected: "[1014040001] Resource not found",
		},
		{
			name: "code only",
			err: &Error{
				code: "1014040001",
			},
			expected: "[1014040001]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := &Error{
		code:  "1014040001",
		cause: originalErr,
	}

	if unwrapped := err.Unwrap(); unwrapped != originalErr {
		t.Errorf("Error.Unwrap() = %v, want %v", unwrapped, originalErr)
	}

	// Test nil cause
	errWithoutCause := &Error{
		code: "1014040001",
	}
	if unwrapped := errWithoutCause.Unwrap(); unwrapped != nil {
		t.Errorf("Error.Unwrap() = %v, want nil", unwrapped)
	}
}

func TestError_Is(t *testing.T) {
	err1 := &Error{code: "1014040001"}
	err2 := &Error{code: "1014040001"}
	err3 := &Error{code: "1014040002"}
	regularErr := errors.New("regular error")

	// Test same code
	if !err1.Is(err2) {
		t.Error("Error.Is() should return true for same code")
	}

	// Test different code
	if err1.Is(err3) {
		t.Error("Error.Is() should return false for different code")
	}

	// Test different error type
	if err1.Is(regularErr) {
		t.Error("Error.Is() should return false for different error type")
	}
}

func TestError_Code(t *testing.T) {
	err := &Error{code: "1014040001"}
	if got := err.Code(); got != "1014040001" {
		t.Errorf("Error.Code() = %q, want %q", got, "1014040001")
	}
}

func TestError_Reason(t *testing.T) {
	err := &Error{reason: "NOT_FOUND"}
	if got := err.Reason(); got != "NOT_FOUND" {
		t.Errorf("Error.Reason() = %q, want %q", got, "NOT_FOUND")
	}
}

func TestError_Message(t *testing.T) {
	err := &Error{msg: "Resource not found"}
	if got := err.Message(); got != "Resource not found" {
		t.Errorf("Error.Message() = %q, want %q", got, "Resource not found")
	}
}

func TestError_CodeInt(t *testing.T) {
	err := &Error{code: "1014040001"}
	expected := 1014040001
	if got := err.CodeInt(); got != expected {
		t.Errorf("Error.CodeInt() = %d, want %d", got, expected)
	}
}

func TestError_HTTPStatus(t *testing.T) {
	err := &Error{code: "1014040001"}
	expected := 404
	if got := err.HTTPStatus(); got != expected {
		t.Errorf("Error.HTTPStatus() = %d, want %d", got, expected)
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		message  string
		wantErr  bool
		expected string
	}{
		{
			name:    "valid 9-digit code with valid HTTP status",
			code:    120000001,
			message: "Success",
			wantErr: true, // 000 is not a valid HTTP status
		},
		{
			name:    "valid 9-digit code with 201 status",
			code:    120100001,
			message: "Created",
			wantErr: true, // 001 is not a valid HTTP status
		},
		{
			name:    "valid 9-digit code with 404 status",
			code:    140400001,
			message: "Not found",
			wantErr: true, // 040 is not a valid HTTP status
		},
		{
			name:    "valid 9-digit code with 500 status",
			code:    150000001,
			message: "Internal error",
			wantErr: true, // 000 is not a valid HTTP status
		},
		{
			name:     "valid 10-digit code",
			code:     1014040001,
			message:  "Not found",
			wantErr:  false,
			expected: "[1014040001] Not found",
		},
		{
			name:    "invalid 8-digit code",
			code:    14040001,
			message: "Not found",
			wantErr: true,
		},
		{
			name:    "invalid 11-digit code",
			code:    10140400010,
			message: "Not found",
			wantErr: true,
		},
		{
			name:    "negative code",
			code:    -1,
			message: "Not found",
			wantErr: true,
		},
		{
			name:    "invalid http status in code",
			code:    1019990001,
			message: "Not found",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.code, tt.message)
			if tt.wantErr {
				if err == nil {
					t.Error("New() expected error but got nil")
				}
				return
			}

			if err == nil {
				t.Error("New() unexpected error")
				return
			}

			if got := err.Error(); got != tt.expected {
				t.Errorf("New() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNewWithReason(t *testing.T) {
	err := NewWithReason(1014040001, "NOT_FOUND", "Resource not found")
	if err == nil {
		t.Error("NewWithReason() unexpected error")
		return
	}

	expected := "[NOT_FOUND:1014040001] Resource not found"
	if got := err.Error(); got != expected {
		t.Errorf("NewWithReason() = %q, want %q", got, expected)
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")

	tests := []struct {
		name     string
		code     int
		err      error
		message  string
		wantErr  bool
		expected string
	}{
		{
			name:     "valid wrap",
			code:     1014040001,
			err:      originalErr,
			message:  "Wrapped error",
			wantErr:  false,
			expected: "[1014040001] Wrapped error",
		},
		{
			name:    "nil error",
			code:    1014040001,
			err:     nil,
			message: "Should not wrap",
			wantErr: false,
		},
		{
			name:    "invalid code",
			code:    14040001,
			err:     originalErr,
			message: "Invalid code",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Wrap(tt.code, tt.err, tt.message)
			if tt.wantErr {
				if err == nil {
					t.Error("Wrap() expected error but got nil")
				}
				return
			}

			if tt.err == nil {
				if err != nil {
					t.Error("Wrap() should return nil for nil error")
				}
				return
			}

			if err == nil {
				t.Error("Wrap() unexpected error")
				return
			}

			if got := err.Error(); got != tt.expected {
				t.Errorf("Wrap() = %q, want %q", got, tt.expected)
			}

			// Test unwrapping
			if !errors.Is(err, originalErr) {
				t.Error("Wrap() should preserve original error")
			}
		})
	}
}

func TestWrapWithReason(t *testing.T) {
	originalErr := errors.New("original error")
	err := WrapWithReason(1014040001, "NOT_FOUND", originalErr, "Resource not found")
	if err == nil {
		t.Error("WrapWithReason() unexpected error")
		return
	}

	expected := "[NOT_FOUND:1014040001] Resource not found"
	if got := err.Error(); got != expected {
		t.Errorf("WrapWithReason() = %q, want %q", got, expected)
	}

	// Test unwrapping
	if !errors.Is(err, originalErr) {
		t.Error("WrapWithReason() should preserve original error")
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		want    string
		wantErr bool
	}{
		{
			name:    "valid 9-digit code with valid HTTP status",
			code:    140400001,
			wantErr: true, // Becomes 140400001, HTTP status 040 is invalid
		},
		{
			name:    "valid 10-digit code",
			code:    1014040001,
			want:    "1014040001",
			wantErr: false,
		},
		{
			name:    "invalid 8-digit code",
			code:    14040001,
			wantErr: true,
		},
		{
			name:    "invalid 11-digit code",
			code:    10140400010,
			wantErr: true,
		},
		{
			name:    "negative code",
			code:    -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalize(tt.code)
			if tt.wantErr {
				if err == nil {
					t.Errorf("normalize() expected error for code %d", tt.code)
				}
				return
			}

			if err != nil {
				t.Errorf("normalize() unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("normalize() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		want    string
		wantErr bool
	}{
		{
			name:    "valid code",
			code:    "1014040001",
			want:    "1014040001",
			wantErr: false,
		},
		{
			name:    "invalid length",
			code:    "14040001",
			wantErr: true,
		},
		{
			name:    "non-digit characters",
			code:    "10140a0001",
			wantErr: true,
		},
		{
			name:    "invalid http status - too low",
			code:    "1010990001",
			wantErr: true,
		},
		{
			name:    "invalid http status - too high",
			code:    "1016000001",
			wantErr: true,
		},
		{
			name:    "valid http status boundaries",
			code:    "1011000001",
			want:    "1011000001",
			wantErr: false,
		},
		{
			name:    "valid http status boundaries",
			code:    "1015990001",
			want:    "1015990001",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validate(tt.code)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validate() expected error for code %s", tt.code)
				}
				return
			}

			if err != nil {
				t.Errorf("validate() unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("validate() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Test error compatibility with Go's error handling
func TestErrorCompatibility(t *testing.T) {
	// Test errors.Is with 9-digit code - all 9-digit codes are invalid due to HTTP status
	err9 := New(1014040001, "9-digit error")
	if err9 == nil {
		t.Fatal("New() returned nil for 9-digit code")
	}
	err9_2 := New(1014040001, "Another 9-digit error")
	if err9_2 == nil {
		t.Fatal("New() returned nil for 9-digit code")
	}

	if !errors.Is(err9, err9_2) {
		t.Error("errors.Is should return true for same 9-digit code")
	}

	// Test errors.As with 9-digit code
	var target9 *Error
	if !errors.As(err9, &target9) {
		t.Error("errors.As should succeed for 9-digit *Error type")
	}

	if target9.Code() != "1014040001" {
		t.Errorf("errors.As target has wrong 9-digit code: %s", target9.Code())
	}

	// Verify HTTP status is correctly parsed from 9-digit code
	if target9.HTTPStatus() != 404 {
		t.Errorf("9-digit code should have HTTP status 404, got %d", target9.HTTPStatus())
	}

	// Test errors.Is with 10-digit code
	err1 := New(1014040001, "Not found")
	if err1 == nil {
		t.Fatal("New() returned nil")
	}
	err2 := New(1014040001, "Also not found")
	if err2 == nil {
		t.Fatal("New() returned nil")
	}
	err3 := New(1014040002, "Different error")
	if err3 == nil {
		t.Fatal("New() returned nil")
	}

	if !errors.Is(err1, err2) {
		t.Error("errors.Is should return true for same code")
	}

	if errors.Is(err1, err3) {
		t.Error("errors.Is should return false for different code")
	}

	// Test errors.As
	var target *Error
	if !errors.As(err1, &target) {
		t.Error("errors.As should succeed for *Error type")
	}

	if target.Code() != "1014040001" {
		t.Errorf("errors.As target has wrong code: %s", target.Code())
	}

	// Test error wrapping with 9-digit code
	originalErr := errors.New("original")
	wrappedErr9 := Wrap(1014040001, originalErr, "wrapped 9-digit")
	if wrappedErr9 == nil {
		t.Fatal("Wrap() returned nil for 9-digit code")
	}

	if !errors.Is(wrappedErr9, originalErr) {
		t.Error("wrapped 9-digit error should contain original error")
	}

	var unwrapped9 *Error
	if !errors.As(wrappedErr9, &unwrapped9) {
		t.Error("errors.As should succeed for wrapped 9-digit error")
	}

	if unwrapped9.Code() != "1014040001" {
		t.Errorf("wrapped 9-digit error has wrong code: %s", unwrapped9.Code())
	}

	// Verify HTTP status is correctly parsed from 9-digit code
	if unwrapped9.HTTPStatus() != 404 {
		t.Errorf("wrapped 9-digit code should have HTTP status 404, got %d", unwrapped9.HTTPStatus())
	}

	// Test error wrapping with 10-digit code
	wrappedErr := Wrap(1014040001, originalErr, "wrapped")
	if wrappedErr == nil {
		t.Fatal("Wrap() returned nil")
	}

	if !errors.Is(wrappedErr, originalErr) {
		t.Error("wrapped error should contain original error")
	}

	var unwrapped *Error
	if !errors.As(wrappedErr, &unwrapped) {
		t.Error("errors.As should succeed for wrapped error")
	}

	if unwrapped.Code() != "1014040001" {
		t.Errorf("wrapped error has wrong code: %s", unwrapped.Code())
	}
}

// Benchmark tests
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New(1014040001, "Not found")
	}
}

func BenchmarkNew9Digit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New(1014040001, "Not found")
	}
}

func BenchmarkNewWithReason(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewWithReason(1014040001, "NOT_FOUND", "Resource not found")
	}
}

func BenchmarkNewWithReason9Digit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewWithReason(1014040001, "NOT_FOUND", "Resource not found")
	}
}

func BenchmarkWrap(b *testing.B) {
	err := errors.New("original error")
	for i := 0; i < b.N; i++ {
		_ = Wrap(1014040001, err, "Wrapped error")
	}
}

func BenchmarkWrap9Digit(b *testing.B) {
	err := errors.New("original error")
	for i := 0; i < b.N; i++ {
		_ = Wrap(1014040001, err, "Wrapped 9-digit error")
	}
}

func BenchmarkError_Error(b *testing.B) {
	err := &Error{
		code:   "1014040001",
		reason: "NOT_FOUND",
		msg:    "Resource not found",
	}
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}
