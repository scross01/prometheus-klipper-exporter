package test

import (
	"testing"

	"github.com/scross01/prometheus-klipper-exporter/collector"
)

func TestGetValidLabelName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "hyphens to underscores",
			input:    "test-label-name",
			expected: "test_label_name",
		},
		{
			name:     "spaces removed",
			input:    "test label name",
			expected: "testlabelname",
		},
		{
			name:     "special characters removed",
			input:    "test@label#name$",
			expected: "testlabelname",
		},
		{
			name:     "mixed invalid characters",
			input:    "test-label name!@#",
			expected: "test_labelname",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "already valid",
			input:    "validLabelName123",
			expected: "validLabelName123",
		},
		{
			name:     "only invalid characters",
			input:    "!@#$%^&*()",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.GetValidLabelName(tt.input)
			if result != tt.expected {
				t.Errorf("GetValidLabelName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
