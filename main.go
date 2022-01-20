package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
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

func ingest(f io.Reader) ([]string, [][]string, []map[string][]int, error) {
	r := csv.NewReader(f)

	header, err := r.Read() // Read first row as header
	if err != nil {
		return nil, nil, nil, err
	}

	rowCount, colCount := 0, len(header)
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

	return header, data, indices, nil
}

func cardinality(indices []map[string][]int) []int { // Count distinct values of columns.
	result := make([]int, len(indices))

	for i, m := range indices {
		result[i] = len(m)
	}

	return result
}
