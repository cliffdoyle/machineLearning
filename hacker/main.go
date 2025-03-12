package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"sort"
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

func CountClassOccurrences(dataset [][]interface{}) map[string]int {
	classCounts := make(map[string]int)
	for _, row := range dataset {
		if len(row) == 0 {
			continue
		}
		targetClass := fmt.Sprintf("%v", row[len(row)-1])
		classCounts[targetClass]++
	}
	return classCounts
}

func ComputeProbabilities(classCounts map[string]int, totalSamples int) map[string]float64 {
	probabilities := make(map[string]float64)
	for class, count := range classCounts {
		probabilities[class] = float64(count) / float64(totalSamples)
	}
	return probabilities
}

func Entropy(dataset [][]interface{}) float64 {
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

func InformationGain(dataset [][]interface{}, header []string, attribute string) float64 {
	totalSamples := len(dataset)
	if totalSamples == 0 {
		return 0
	}

	initialEntropy := Entropy(dataset)
	splitted := SplitDataset(dataset, header, attribute)

	weightedEntropy := 0.0
	for _, subset := range splitted {
		proportion := float64(len(subset)) / float64(totalSamples)
		weightedEntropy += proportion * Entropy(subset)
	}

	return initialEntropy - weightedEntropy
}

func GainRatio(dataset [][]interface{}, header []string, attribute string) float64 {
	totalSamples := len(dataset)
	if totalSamples == 0 {
		return 0
	}

	infoGain := InformationGain(dataset, header, attribute)
	if infoGain == 0 {
		return 0
	}

	splitted := SplitDataset(dataset, header, attribute)

	splitInfo := 0.0
	for _, subset := range splitted {
		proportion := float64(len(subset)) / float64(totalSamples)
		if proportion > 0 {
			splitInfo -= proportion * math.Log2(proportion)
		}
	}

	if splitInfo == 0 {
		return 0
	}

	return infoGain / splitInfo
}

func SplitDataset(dataset [][]interface{}, header []string, attribute string) map[string][][]interface{} {
	subsets := make(map[string][][]interface{})
	attrIndex := -1

	for i, col := range header {
		if col == attribute {
			attrIndex = i
			break
		}
	}

	if attrIndex == -1 {
		fmt.Println("Error: Attribute not found in header")
		return subsets
	}

	for _, row := range dataset {
		if attrIndex < len(row) {
			key := fmt.Sprintf("%v", row[attrIndex])
			subsets[key] = append(subsets[key], row)
		}
	}

	return subsets
}

func FindBestThreshold(dataset [][]interface{}, attrIndex int) float64 {
	var values []float64
	for _, row := range dataset {
		if val, ok := row[attrIndex].(float64); ok {
			values = append(values, val)
		}
	}

	sort.Float64s(values)

	var bestThreshold float64
	bestInfoGain := -1.0

	for i := 0; i < len(values)-1; i++ {
		threshold := (values[i] + values[i+1]) / 2.0
		infoGain := EvaluateThreshold(dataset, attrIndex, threshold)
		if infoGain > bestInfoGain {
			bestInfoGain = infoGain
			bestThreshold = threshold
		}
	}

	return bestThreshold
}

func EvaluateThreshold(dataset [][]interface{}, attrIndex int, threshold float64) float64 {
	var leftSubset, rightSubset [][]interface{}

	for _, row := range dataset {
		if val, ok := row[attrIndex].(float64); ok {
			if val <= threshold {
				leftSubset = append(leftSubset, row)
			} else {
				rightSubset = append(rightSubset, row)
			}
		}
	}

	totalSamples := len(dataset)
	if totalSamples == 0 {
		return 0
	}

	initialEntropy := Entropy(dataset)
	weightedEntropy := (float64(len(leftSubset)) / float64(totalSamples) * Entropy(leftSubset)) +
		(float64(len(rightSubset)) / float64(totalSamples) * Entropy(rightSubset))

	return initialEntropy - weightedEntropy
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

	fmt.Println("Entropy:", Entropy(dataset))
}
