package flags

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestParseID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{"valid positive", "123", 123, false},
		{"valid negative", "-123", -123, false},
		{"zero", "0", 0, false},
		{"invalid", "abc", 0, true},
		{"empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseID(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseIDFromArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		index   int
		want    int64
		wantErr bool
	}{
		{"first arg", []string{"123"}, 0, 123, false},
		{"second arg", []string{"first", "456"}, 1, 456, false},
		{"index out of range", []string{"123"}, 5, 0, true},
		{"empty args", []string{}, 0, 0, true},
		{"invalid id", []string{"abc"}, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIDFromArgs(tt.args, tt.index)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateRequiredID(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		index   int
		idName  string
		want    int64
		wantErr bool
	}{
		{"valid", []string{"123"}, 0, "test-id", 123, false},
		{"missing", []string{}, 0, "test-id", 0, true},
		{"invalid", []string{"abc"}, 0, "test-id", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateRequiredID(tt.args, tt.index, tt.idName)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetFlagString(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("test-flag", "default", "test flag")
	cmd.SetArgs([]string{"--test-flag", "value"})
	cmd.Execute()

	assert.Equal(t, "value", GetFlagString(cmd, "test-flag"))
	assert.Equal(t, "", GetFlagString(cmd, "non-existent"))
}

func TestGetFlagBool(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("test-bool", false, "test bool")
	cmd.SetArgs([]string{"--test-bool"})
	cmd.Execute()

	assert.True(t, GetFlagBool(cmd, "test-bool"))
	assert.False(t, GetFlagBool(cmd, "non-existent"))
}
