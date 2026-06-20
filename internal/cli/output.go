package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

// emitJSON pretty-prints v as JSON to stdout (with a trailing newline) and
// returns the success exit code, or routes an encoding failure through fail.
func (a *App) emitJSON(v any) int {
	enc := json.NewEncoder(a.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return a.fail(fmt.Errorf("encoding JSON: %w", err))
	}
	return exitOK
}

// fmtNum renders a float without scientific notation or trailing zeros, so
// rates and amounts print as "0.92" / "100" / "1234.5" rather than
// "0.920000" or "1.234500e+03".
func fmtNum(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// sortedKeys returns the keys of a string-keyed float map in sorted order.
func sortedKeys(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// maxLen returns the length of the longest string in s (0 for empty input).
func maxLen(s []string) int {
	n := 0
	for _, v := range s {
		if len(v) > n {
			n = len(v)
		}
	}
	return n
}
