package p

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrint(t *testing.T) {
	Error("Hello, kitty")
	Info("hello kitty")
	Success("hello kitty")
}

func TestStructuredPrint(t *testing.T) {
	require.NoError(t, PrintJSON(map[string]string{"hello": "world"}))
	require.NoError(t, PrintYAML(map[string]string{"hello": "world"}))
}

func TestPrintJSONError(t *testing.T) {
	err := PrintJSON(map[string]float64{"bad": math.Inf(1)})

	require.Error(t, err)
	var syntaxErr *json.UnsupportedValueError
	assert.ErrorAs(t, err, &syntaxErr)
}
