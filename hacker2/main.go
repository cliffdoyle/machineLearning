package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
	"math"
)

// LoadCsv loads a CSV file and detects data types (categorical, numeric, date)
func LoadCsv(filename string) ([]string, [][]interface{}, []string, error) {
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
	rawData := records[1:]

	// Detect column data types
	colTypes := detectColumnTypes(rawData)

	// Convert dataset based on detected types
	var dataset [][]interface{}
	for _, row := range rawData {
		var convertedRow []interface{}
		for i, val := range row {
			switch colTypes[i] {
			case "numeric":
				num, _ := strconv.ParseFloat(val, 64)
				convertedRow = append(convertedRow, num)
			case "date":
				parsedTime, _ := parseDate(val)
				convertedRow = append(convertedRow, parsedTime)
			default:
				convertedRow = append(convertedRow, val) // Keep as string
			}
		}
		dataset = append(dataset, convertedRow)
	}

	return header, dataset, colTypes, nil
}

// detectColumnTypes determines if each column is categorical, numeric, or a date
func detectColumnTypes(data [][]string) []string {
	colCount := len(data[0])
	colTypes := make([]string, colCount)

	for col := 0; col < colCount; col++ {
		isNumeric := true
		isDate := true

		for row := 0; row < len(data); row++ {
			if _, err := strconv.ParseFloat(data[row][col], 64); err != nil {
				isNumeric = false
			}
			if _, err := parseDate(data[row][col]); err != nil {
				isDate = false
			}
		}

		if isNumeric {
			colTypes[col] = "numeric"
		} else if isDate {
			colTypes[col] = "date"
		} else {
			colTypes[col] = "categorical"
		}
	}
	return colTypes
}

// parseDate tries to parse a string into a time.Time object
func parseDate(value string) (time.Time, error) {
	formats := []string{"2006-01-02", "02/01/2006", "01-02-2006", "2006/01/02"}
	for _, format := range formats {
		t, err := time.Parse(format, value)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date format: %s", value)
}


// CountClassOccurrences counts occurrences of each target class in the dataset
func CountClassOccurrences(dataset [][]interface{}) map[string]int {
	classCounts := make(map[string]int)

	for _, row := range dataset {
		if len(row) == 0 {
			continue
		}
		targetClass, ok := row[len(row)-1].(string) // Ensure it's categorical
		if !ok {
			continue // Skip if it's not a string (categorical class)
		}
		classCounts[targetClass]++
	}

	return classCounts
}


// ComputeProbabilities calculates the probability of each class in the dataset
func ComputeProbabilities(classCounts map[string]int, totalSamples int) map[string]float64 {
	probabilities := make(map[string]float64)

	for class, count := range classCounts {
		probabilities[class] = float64(count) / float64(totalSamples)
	}
	return probabilities
}


// Entropy calculates the entropy of the dataset (impurity measure)
func Entropy(dataset [][]interface{}) float64 {
	countClassOccurrences := CountClassOccurrences(dataset)
	totalSamples := len(dataset)
	if totalSamples == 0 {
		return 0.0
	}

	probabilities := ComputeProbabilities(countClassOccurrences, totalSamples)
	entropy := 0.0

	for _, probability := range probabilities {
		if probability > 0 {
			entropy -= probability * math.Log2(probability)
		}
	}
	return entropy
}


// func main() {
	// header, dataset, colTypes, err := LoadCsv("data.csv")
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }

	// fmt.Println("Headers:", header)
	// fmt.Println("Column Types:", colTypes)
	// fmt.Println("Dataset:", dataset)

	// Example usage
func main() {
	// Sample dataset with categorical class labels
	dataset := [][]interface{}{
		{"Sunny", 85.0, "Hot", "No"},
		{"Rainy", 75.0, "Cool", "Yes"},
		{"Overcast", 78.0, "Mild", "Yes"},
		{"Sunny", 90.0, "Hot", "No"},
	}

	classCounts := CountClassOccurrences(dataset)
	fmt.Println("Class Occurrences:", classCounts)

	totalSamples := len(dataset)
	probabilities := ComputeProbabilities(classCounts, totalSamples)
	fmt.Println("Class Probabilities:", probabilities)

	entropy := Entropy(dataset)
	fmt.Println("Entropy of dataset:", entropy)
}

