package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type ColumnType int

const (
	Categorical ColumnType = iota
	Numeric
	Datetime
)

func (c ColumnType) String() string {
	return [...]string{"Categorical", "Numeric", "Datetime"}[c]
}

func LoadCsv(filename string) ([]string, [][]interface{}, []ColumnType, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error reading file: %v", err)
	}

	if len(records) < 2 {
		return nil, nil, nil, fmt.Errorf("insufficient data in CSV file")
	}

	header := records[0]
	dataset := make([][]interface{}, len(records)-1)
	colTypes := detectColumnTypes(records[1:])

	for i, row := range records[1:] {
		dataset[i] = make([]interface{}, len(row))
		for j, value := range row {
			dataset[i][j] = convertValue(value, colTypes[j])
		}
	}

	return header, dataset, colTypes, nil
}

func detectColumnTypes(records [][]string) []ColumnType {
	colCount := len(records[0])
	colTypes := make([]ColumnType, colCount)

	dateFormats := []string{
		"2006-01-02", "02-01-2006", "01/02/2006",
		"2006/01/02", "Jan 2, 2006", "02 Jan 2006",
		"Monday, Jan 2 2006",
	}

	for j := 0; j < colCount; j++ {
		isNumeric, isDatetime := true, true
		hasValidNumeric, hasValidDatetime := false, false

		for _, row := range records {
			value := strings.TrimSpace(row[j])
			if value == "" {
				continue
			}

			if _, err := strconv.ParseFloat(value, 64); err != nil {
				isNumeric = false
			} else {
				hasValidNumeric = true
			}

			validDate := false
			for _, format := range dateFormats {
				if _, err := time.Parse(format, value); err == nil {
					validDate = true
					hasValidDatetime = true
					break
				}
			}
			if !validDate {
				isDatetime = false
			}
		}

		if isNumeric && hasValidNumeric {
			colTypes[j] = Numeric
		} else if isDatetime && hasValidDatetime {
			colTypes[j] = Datetime
		} else {
			colTypes[j] = Categorical
		}
	}

	return colTypes
}

func convertValue(value string, colType ColumnType) interface{} {
	value = strings.TrimSpace(value)
	switch colType {
	case Numeric:
		num, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return value
		}
		return num
	case Datetime:
		dateFormats := []string{
			"2006-01-02", "02-01-2006", "01/02/2006",
			"2006/01/02", "Jan 2, 2006", "02 Jan 2006",
			"Monday, Jan 2 2006",
		}
		for _, format := range dateFormats {
			if date, err := time.Parse(format, value); err == nil {
				return date
			}
		}
		return value
	default:
		return value
	}
}

// CountClass counts the occurrence of the target class in
// our dataset
func CountClassOccurrences(dataset [][]string) map[string]int {
	classCounts := make(map[string]int)

	for _, row := range dataset {

		if len(row) == 0 {
			continue
		}
		targetClass := row[len(row)-1]
		classCounts[targetClass]++
	}

	return classCounts
}

// Calculates probability of each class
func ComputeProbabilities(classCounts map[string]int, totalSamples int) map[string]float64 {
	probabilities := make(map[string]float64)

	for class, count := range classCounts {
		probabilities[class] = float64(count) / float64(totalSamples)
	}
	return probabilities
}

// Calculates entropy based on probabilities to determine the impurity of the dataset
func Entropy(dataset [][]string) float64 {
	countClassOccurrences := CountClassOccurrences(dataset)
	totalSamples := len(dataset)
	probabilities := ComputeProbabilities(countClassOccurrences, totalSamples)
	entropy := 0.0

	for _, probability := range probabilities {
		if probability > 0 {
			entropy -= probability * math.Log2(probability)
		}
	}
	return entropy
}

func main() {
	head, dataset, col, err := LoadCsv("data.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Header:", head)

	fmt.Println("Dataset:")
	for _, row := range dataset {
		fmt.Println(row)
	}

	fmt.Print("Column Types: ")
	for _, c := range col {
		fmt.Print(c.String(), " ")
	}
	fmt.Println()
}
