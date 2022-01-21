package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
)

func main() {
	f, err := os.Open("data/train.csv")
	if err != nil {
		log.Fatalln("OPen unknown error", err)
	}

	headers, _, indices, err := ingest(f)
	if err != nil {
		log.Fatalln("ingest unknown error", err)
	}

	c := cardinality(indices)
	for i, h := range headers {
		fmt.Printf("%s: %v\n", h, c[i])
	}
}

// Parse header and indices from raw data
func ingest(f io.Reader) ([]string, [][]string, []map[string][]int, error) {
	r := csv.NewReader(f)

	headers, err := r.Read() // Read first row as header
	if err != nil {
		return nil, nil, nil, err
	}

	rowCount, colCount := 0, len(headers)
	indices := make([]map[string][]int, colCount) // By columns
	data := make([][]string, 0)                   // raw data

	for row, err := r.Read(); err == nil; row, err = r.Read() {
		if len(row) != colCount {
			return nil, nil, nil, fmt.Errorf("expected columns: %d. Got %d columns in row %d", colCount, len(row), rowCount)
		}
		data = append(data, row)

		for colIndex, v := range row {
			if indices[colIndex] == nil {
				indices[colIndex] = make(map[string][]int)
			}
			indices[colIndex][v] = append(indices[colIndex][v], rowCount) // Record positions of the value.
		}
		rowCount++
	}

	return headers, data, indices, nil
}

// Count distinct values of columns.
func cardinality(indices []map[string][]int) []int {
	result := make([]int, len(indices))

	for i, m := range indices {
		result[i] = len(m)
	}

	return result
}

func clean(headers []string, data [][]string, indices []map[string][]int, hints []bool, ignored []string) (int, int, []float64, []float64, []string, []bool) {
	modes := mode(indices)
	var xs, ys []float64 // xs is a rowLen * colLen matrix, ys is a rowLen * 1 matrix
	var newHints []bool
	var newHeaders []string
	var colLen int

	for i, row := range data {
		for j, raw := range row {
			if headers[j] == "Id" {
				continue // Ignore column:Id
			}

			// Use column:SalePrice as Y value
			if headers[j] == "SalePrice" {
				y, _ := convert(raw, headers[j])
				ys = append(ys, y...)

				continue
			}

			if inList(headers[j], ignored) {
				continue
			}

			// Convert raw string to []float64 and new columns
			var x []float64
			var xName []string
			if hints[j] {
				raw = imputeCategorical(raw, j, headers, modes)

				x, xName = convertCategorical(raw, indices[j], headers[j])
			} else {
				x, xName = convert(raw, headers[j])
			}
			xs = append(xs, x...)

			// Extend hints and headers
			if i == 0 {
				h := make([]bool, len(x))

				for k := range h {
					h[k] = hints[j] // Clone old hint to new hints
				}
				newHints = append(newHints, h...)
				newHeaders = append(newHeaders, xName...)
			}
		}

		// End of a row, calculate column number
		if i == 0 {
			colLen = len(xs)
		}
	}

	rowLen := len(data)
	if len(ys) == 0 {
		ys = make([]float64, rowLen)
	}

	return rowLen, colLen, xs, ys, newHeaders, newHints
}

func inList(header string, ignored []string) bool {
	for _, v := range ignored {
		if header == v {
			return true
		}
	}

	return false
}

// Convert string to []float64
func convert(raw string, name string) ([]float64, []string) {
	v, _ := strconv.ParseFloat(raw, 64)

	return []float64{v}, []string{name}
}

// Convert categorical string to []float64
func convertCategorical(raw string, index map[string][]int, name string) ([]float64, []string) {
	result := make([]float64, len(index)-1) // Modified one-hot method costs len-1 space.

	// Get all categorical values
	keys := make([]string, 0, len(index))
	for k := range index {
		keys = append(keys, k) // keys is unsorted in map.
	}
	keys = sortKeys(keys) // keys is sorted.

	// Move NA to index 0 if exists.
	naIndex := 0
	for i, key := range keys {
		if key == "NA" {
			naIndex = i
			break
		}
	}
	keys[0], keys[naIndex] = keys[naIndex], keys[0]

	// Convert categorical string to be modified one-hot array.
	for i, key := range keys[1:] { // First key/NA is default in modified one-hot method.
		if key == raw {
			result[i] = 1
			break
		}
	}

	// Naming extended categorical values
	for i, key := range keys {
		keys[i] = fmt.Sprintf("%s_%s", name, key)
	}

	return result, keys[1:]
}

// Sort categorical values by numeric or ASCII.
func sortKeys(cols []string) []string {
	result := make([]string, len(cols))
	for i, v := range cols {
		result[i] = v
	}

	sort.Slice(result, func(i, j int) bool {
		u, err1 := strconv.ParseFloat(result[i], 64)
		v, err2 := strconv.ParseFloat(result[j], 64)

		if err1 == nil && err2 == nil {
			return u < v // ascend
		}

		return result[i] < result[j] // ascend
	})

	return result
}

// Get most common categorical value by column
func mode(indices []map[string][]int) []string {
	result := make([]string, len(indices))

	for i, m := range indices {
		var max int

		for k, v := range m {
			if len(v) > max {
				max = len(v)

				result[i] = k
			}
		}
	}

	return result
}

// Replace "NA" or empty value by default categorical value.
func imputeCategorical(raw string, col int, headers []string, modes []string) string {
	if raw != "NA" || raw != "" {
		return raw
	}

	switch headers[col] {
	case "MSZoning", "BsmtFullBath", "BsmtHalfBath", "Utilities", "Functional", "Electrical", "KitchenQual", "SaleType", "Exterior1st", "Exterior2nd":
		return modes[col]
	}

	return raw
}
