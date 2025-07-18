package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// TestUnstructuredLogsCheck tests the unstructuredLogs function
func TestUnstructuredLogsCheck(t *testing.T) { //nolint:paralleltest // Uses environment variables
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"Default Case", "", true},
		{"Explicitly True", "true", true},
		{"Explicitly False", "false", false},
		{"Invalid Value", "not-a-bool", true},
	}

	for _, tt := range tests { //nolint:paralleltest // Uses environment variables
		t.Run(tt.name, func(t *testing.T) { //nolint:paralleltest // Uses environment variables
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("UNSTRUCTURED_LOGS", tt.envValue)
				defer os.Unsetenv("UNSTRUCTURED_LOGS")
			} else {
				os.Unsetenv("UNSTRUCTURED_LOGS")
			}

			if got := unstructuredLogs(); got != tt.expected {
				t.Errorf("unstructuredLogs() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestStructuredLogger tests the structured logger functionality
// TODO: Keeping this for migration but can be removed as we don't need really need to test zap
func TestStructuredLogger(t *testing.T) { //nolint:paralleltest // Uses environment variables
	// Test cases for basic logging methods (Debug, Info, Warn, etc.)
	basicLogTestCases := []struct {
		level   string // The log level to test
		message string // The message to log
	}{
		{"debug", "debug message"},
		{"info", "info message"},
		{"warn", "warn message"},
		{"error", "error message"},
		{"dpanic", "dpanic message"},
		{"panic", "panic message"},
	}

	for _, tc := range basicLogTestCases {
		t.Run("BasicLogs_"+tc.level, func(t *testing.T) {
			// we create a pipe to capture the output of the log
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			os.Setenv("UNSTRUCTURED_LOGS", "false")
			defer os.Unsetenv("UNSTRUCTURED_LOGS")

			viper.SetDefault("debug", true)

			logger, err := NewLogger()
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			// Handle panic and fatal recovery
			defer func() {
				if r := recover(); r != nil {
					if tc.level != "panic" && tc.level != "dpanic" {
						t.Errorf("Unexpected panic for level %s: %v", tc.level, r)
					}
				}
			}()

			// Log using basic methods
			switch tc.level {
			case "debug":
				logger.Debug(tc.message)
			case "info":
				logger.Info(tc.message)
			case "warn":
				logger.Warn(tc.message)
			case "error":
				logger.Error(tc.message)
			case "dpanic":
				logger.DPanic(tc.message)
			case "panic":
				logger.Panic(tc.message)
			}

			w.Close()
			os.Stdout = originalStdout

			// Read the captured output
			var capturedOutput bytes.Buffer
			io.Copy(&capturedOutput, r)
			output := capturedOutput.String()

			// Parse JSON output
			var logEntry map[string]any
			if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
				t.Fatalf("Failed to parse JSON log output: %v", err)
			}

			// Check level
			if level, ok := logEntry["level"].(string); !ok || level != tc.level {
				t.Errorf("Expected level %s, got %v", tc.level, logEntry["level"])
			}

			// Check message
			if msg, ok := logEntry["msg"].(string); !ok || msg != tc.message {
				t.Errorf("Expected message %s, got %v", tc.message, logEntry["msg"])
			}
		})
	}

	// Test cases for structured logging methods (Debugw, Infow, etc.)
	structuredLogTestCases := []struct {
		level   string // The log level to test
		message string // The message to log
		key     string // Key for structured logging
		value   string // Value for structured logging
	}{
		{"debug", "debug message", "key", "value"},
		{"info", "info message", "key", "value"},
		{"warn", "warn message", "key", "value"},
		{"error", "error message", "key", "value"},
		{"dpanic", "dpanic message", "key", "value"},
		{"panic", "panic message", "key", "value"},
	}

	for _, tc := range structuredLogTestCases {
		t.Run("StructuredLogs_"+tc.level, func(t *testing.T) {
			// we create a pipe to capture the output of the log
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			os.Setenv("UNSTRUCTURED_LOGS", "false")
			defer os.Unsetenv("UNSTRUCTURED_LOGS")

			viper.SetDefault("debug", true)

			logger, err := NewLogger()
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			// Handle panic and fatal recovery
			defer func() {
				if r := recover(); r != nil {
					if tc.level != "panic" && tc.level != "dpanic" {
						t.Errorf("Unexpected panic for level %s: %v", tc.level, r)
					}
				}
			}()

			// Log using structured methods
			switch tc.level {
			case "debug":
				logger.Debugw(tc.message, tc.key, tc.value)
			case "info":
				logger.Infow(tc.message, tc.key, tc.value)
			case "warn":
				logger.Warnw(tc.message, tc.key, tc.value)
			case "error":
				logger.Errorw(tc.message, tc.key, tc.value)
			case "dpanic":
				logger.DPanicw(tc.message, tc.key, tc.value)
			case "panic":
				logger.Panicw(tc.message, tc.key, tc.value)
			}

			w.Close()
			os.Stdout = originalStdout

			// Read the captured output
			var capturedOutput bytes.Buffer
			io.Copy(&capturedOutput, r)
			output := capturedOutput.String()

			// Parse JSON output
			var logEntry map[string]any
			if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
				t.Fatalf("Failed to parse JSON log output: %v", err)
			}

			// Check level
			if level, ok := logEntry["level"].(string); !ok || level != tc.level {
				t.Errorf("Expected level %s, got %v", tc.level, logEntry["level"])
			}

			// Check message
			if msg, ok := logEntry["msg"].(string); !ok || msg != tc.message {
				t.Errorf("Expected message %s, got %v", tc.message, logEntry["msg"])
			}

			// Check key-value pair
			if value, ok := logEntry[tc.key].(string); !ok || value != tc.value {
				t.Errorf("Expected %s=%s, got %v", tc.key, tc.value, logEntry[tc.key])
			}
		})
	}

	// Test cases for formatted logging methods (Debugf, Infof, etc.)
	formattedLogTestCases := []struct {
		level    string
		message  string
		key      string
		value    string
		expected string
		contains bool
	}{
		{"debug", "debug message %s and %s", "key", "value", "debug message key and value", true},
		{"info", "info message %s and %s", "key", "value", "info message key and value", true},
		{"warn", "warn message %s and %s", "key", "value", "warn message key and value", true},
		{"error", "error message %s and %s", "key", "value", "error message key and value", true},
		{"dpanic", "dpanic message %s and %s", "key", "value", "dpanic message key and value", true},
		{"panic", "panic message %s and %s", "key", "value", "panic message key and value", true},
	}

	for _, tc := range formattedLogTestCases {
		t.Run("FormattedLogs_"+tc.level, func(t *testing.T) {
			// we create a pipe to capture the output of the log
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			os.Setenv("UNSTRUCTURED_LOGS", "false")
			defer os.Unsetenv("UNSTRUCTURED_LOGS")

			viper.SetDefault("debug", true)

			logger, err := NewLogger()
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			// Handle panic and fatal recovery
			defer func() {
				if r := recover(); r != nil {
					if tc.level != "panic" && tc.level != "dpanic" {
						t.Errorf("Unexpected panic for level %s: %v", tc.level, r)
					}
				}
			}()

			// Log using formatted methods
			switch tc.level {
			case "debug":
				logger.Debugf(tc.message, tc.key, tc.value)
			case "info":
				logger.Infof(tc.message, tc.key, tc.value)
			case "warn":
				logger.Warnf(tc.message, tc.key, tc.value)
			case "error":
				logger.Errorf(tc.message, tc.key, tc.value)
			case "dpanic":
				logger.DPanicf(tc.message, tc.key, tc.value)
			case "panic":
				logger.Panicf(tc.message, tc.key, tc.value)
			}

			w.Close()
			os.Stdout = originalStdout

			// Read the captured output
			var capturedOutput bytes.Buffer
			io.Copy(&capturedOutput, r)
			output := capturedOutput.String()

			// Parse JSON output
			var logEntry map[string]any
			if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
				t.Fatalf("Failed to parse JSON log output: %v", err)
			}

			// Check level
			if level, ok := logEntry["level"].(string); !ok || level != tc.level {
				t.Errorf("Expected level %s, got %v", tc.level, logEntry["level"])
			}

			// Check message
			if msg, ok := logEntry["msg"].(string); !ok || msg != tc.expected {
				t.Errorf("Expected message %s, got %v", tc.expected, logEntry["msg"])
			}
		})
	}
}

// TestUnstructuredLogger tests the unstructured logger functionality
// TODO: Keeping this for migration but can be removed as we don't need really need to test zap
func TestUnstructuredLogger(t *testing.T) { //nolint:paralleltest // Uses environment variables
	// we only test for the formatted logs here because the unstructured logs
	// do not contain the key/value pair format that the structured logs do
	formattedLogTestCases := []struct {
		level    string
		message  string
		key      string
		value    string
		expected string
	}{
		{"DEBUG", "debug message %s and %s", "key", "value", "debug message key and value"},
		{"INFO", "info message %s and %s", "key", "value", "info message key and value"},
		{"WARN", "warn message %s and %s", "key", "value", "warn message key and value"},
		{"ERROR", "error message %s and %s", "key", "value", "error message key and value"},
		{"DPANIC", "error message %s and %s", "key", "value", "dpanic message key and value"},
		{"PANIC", "error message %s and %s", "key", "value", "panic message key and value"},
	}

	for _, tc := range formattedLogTestCases { //nolint:paralleltest // Uses environment variables
		t.Run("FormattedLogs_"+tc.level, func(t *testing.T) {

			// we create a pipe to capture the output of the log
			// so we can test that the logger logs the right message
			originalStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			viper.SetDefault("debug", true)

			logger, err := NewLogger()
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			// Handle panic recovery for DPANIC and PANIC levels
			defer func() {
				if r := recover(); r != nil {
					// Expected for panic levels
					if tc.level != "PANIC" && tc.level != "DPANIC" {
						t.Errorf("Unexpected panic for level %s: %v", tc.level, r)
					}
				}
			}()

			// Log the message based on the level
			switch tc.level {
			case "DEBUG":
				logger.Debugf(tc.message, tc.key, tc.value)
			case "INFO":
				logger.Infof(tc.message, tc.key, tc.value)
			case "WARN":
				logger.Warnf(tc.message, tc.key, tc.value)
			case "ERROR":
				logger.Errorf(tc.message, tc.key, tc.value)
			case "DPANIC":
				logger.DPanicf(tc.message, tc.key, tc.value)
			case "PANIC":
				logger.Panicf(tc.message, tc.key, tc.value)
			}

			w.Close()
			os.Stderr = originalStderr

			// Read the captured output
			var capturedOutput bytes.Buffer
			io.Copy(&capturedOutput, r)
			output := capturedOutput.String()

			assert.Contains(t, output, tc.level, "Expected log entry '%s' to contain log level '%s'", output, tc.level)
			assert.Contains(t, output, tc.expected, "Expected log entry '%s' to contain message '%s'", output, tc.expected)
		})
	}
}

// TestNewLogger tests the NewLogger function
func TestNewLogger(t *testing.T) { //nolint:paralleltest // Uses environment variables
	// Test structured logs (JSON)
	t.Run("Structured Logs", func(t *testing.T) { //nolint:paralleltest // Uses environment variables
		// Set environment to use structured logs
		os.Setenv("UNSTRUCTURED_LOGS", "false")
		defer os.Unsetenv("UNSTRUCTURED_LOGS")

		// Redirect stdout to capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Initialize the logger
		logger, err := NewLogger()
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		// Log a test message
		logger.Info("test message")

		// Restore stdout
		w.Close()
		os.Stdout = oldStdout

		// Read captured output
		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Verify JSON format
		var logEntry map[string]any
		if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
			t.Fatalf("Failed to parse JSON log output: %v", err)
		}

		if msg, ok := logEntry["msg"].(string); !ok || msg != "test message" {
			t.Errorf("Expected message 'test message', got %v", logEntry["msg"])
		}
	})

	// Test unstructured logs
	t.Run("Unstructured Logs", func(t *testing.T) { //nolint:paralleltest // Uses environment variables
		// Set environment to use unstructured logs
		os.Setenv("UNSTRUCTURED_LOGS", "true")
		defer os.Unsetenv("UNSTRUCTURED_LOGS")

		// Redirect stderr to capture output
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		// Initialize the logger
		logger, err := NewLogger()
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		// Log a test message
		logger.Info("test message", "key", "value")

		// Restore stderr
		w.Close()
		os.Stderr = oldStderr

		// Read captured output
		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Verify unstructured format (should contain message but not be JSON)
		if !strings.Contains(output, "test message") {
			t.Errorf("Expected output to contain 'test message', got %s", output)
		}

		if !strings.Contains(output, "INF") {
			t.Errorf("Expected output to contain 'INF', got %s", output)
		}
	})
}

// TestNamedNewLogger tests the NewLogger function with zap Named functionality
func TestNamedNewLogger(t *testing.T) { //nolint:paralleltest // Uses environment variables
	// Set up structured logger for testing
	os.Setenv("UNSTRUCTURED_LOGS", "false")
	defer os.Unsetenv("UNSTRUCTURED_LOGS")

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a new logger
	logger, err := NewLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create a named logger for the component
	componentLogger := logger.Named("test-component")

	// Log a test message
	componentLogger.Info("component message")

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON output
	var logEntry map[string]any
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log output: %v", err)
	}

	// Verify the component was added
	if logger, ok := logEntry["logger"].(string); !ok || logger != "test-component" {
		t.Errorf("Expected logger='test-component', got %v", logEntry["logger"])
	}
}
