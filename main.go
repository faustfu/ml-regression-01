// https://github.com/PacktPublishing/Go-Machine-Learning-Projects/tree/master/Chapter02
package main

import (
	"encoding/csv"
	"fmt"
	"gorgonia.org/tensor"
	"io"
	"log"
	"os"
	"strconv"
)

func main() {
	f, err := os.Open("data/train.csv")
	if err != nil {
		log.Fatalln("Open unknown error", err)
	}

	headers, data, indices, err := ingest(f)
	if err != nil {
		log.Fatalln("ingest unknown error", err)
	}

	c := cardinality(indices)
	for i, h := range headers {
		fmt.Printf("%s: %v\n", h, c[i])
	}

	rowLen, colLen, xsBack, ysBack, _, _ := clean(headers, data, indices, datahints, ignored)
	xs := tensor.New(tensor.WithShape(rowLen, colLen), tensor.WithBacking(xsBack))
	ys := tensor.New(tensor.WithShape(rowLen, 1), tensor.WithBacking(ysBack))
	fmt.Printf("xs:\n%+1.1sys:\n%1.1s", xs, ys)
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
				y, _ := convert(raw, false, headers[j], nil, "")
				ys = append(ys, y...)

				continue
			}

			if inList(headers[j], ignored) {
				continue
			}

			x, xName := convert(raw, hints[j], headers[j], indices[j], modes[j]) // Convert raw string to []float64 and new columns

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

func convert(raw string, hint bool, header string, index map[string][]int, mode string) ([]float64, []string) {
	if hint {
		raw = imputeCategorical(raw, header, mode)

		return convertCategorical(raw, header, index)
	}

	// Convert string to []float64
	v, _ := strconv.ParseFloat(raw, 64)

	return []float64{v}, []string{header}
}

// Convert categorical string to []float64
func convertCategorical(raw string, header string, index map[string][]int) ([]float64, []string) {
	result := make([]float64, len(index)-1) // Modified one-hot method costs len-1 space.

	// Get all categorical values
	keys := make([]string, 0, len(index))
	for k := range index {
		keys = append(keys, k) // keys is unsorted in map.
	}
	keys = sortKeys(index, keys) // keys is sorted.

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
		keys[i] = fmt.Sprintf("%s_%s", header, key)
	}

	return result, keys[1:]
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
func imputeCategorical(raw string, header string, mode string) string {
	if raw != "NA" || raw != "" {
		return raw
	}

	switch header {
	case "MSZoning", "BsmtFullBath", "BsmtHalfBath", "Utilities", "Functional", "Electrical", "KitchenQual", "SaleType", "Exterior1st", "Exterior2nd":
		return mode
	}

	return raw
}
