package main

import (
	"sort"
	"strconv"
)

// Sort categorical values by numeric or ASCII.
func sortKeys(index map[string][]int, cols []string) []string {
	result := make([]string, len(cols))
	for i, v := range cols {
		result[i] = v
	}

	isNumCat := true
	cats := make([]int, 0, len(index))
	for k := range index {
		i64, err := strconv.ParseInt(k, 10, 64)
		if err != nil && k != "NA" {
			isNumCat = false
			break
		}
		cats = append(cats, int(i64))
	}

	if isNumCat {
		sort.Ints(cats)
		for i := range cats {
			result[i] = strconv.Itoa(cats[i])
		}
		if _, ok := index["NA"]; ok {
			result[0] = "NA" // there are no negative numerical categories
		}
	} else {
		sort.Strings(result)
	}

	return result
}

func inList(header string, ignored []string) bool {
	for _, v := range ignored {
		if header == v {
			return true
		}
	}

	return false
}
