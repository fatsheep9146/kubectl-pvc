package app

import (
	"github.com/gsamokovarov/assert"
	"testing"
)

func TestVersionParser(t *testing.T) {
	t.Parallel() // marks TLog as capable of running in parallel with other tests
	tests := []struct {
		name    string
		chart   string
		version string
	}{
		{"chart-v2.8-b.9", "chart", "v2.8-b.9"},
		{"chart-v2.8.1", "chart", "v2.8.1"},
	}

	for _, tt := range tests {
		tt := tt // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // marks each test case as capable of running in parallel with each other
			chart, version := parseVersion(tt.name)
			assert.Equal(t, chart, tt.chart)
			assert.Equal(t, version, tt.version)
		})
	}
}
