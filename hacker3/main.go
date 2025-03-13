package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"time"
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
	fmt.Println("data[0]", data[0])
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

// SplitDataset handles both categorical and numerical attributes
func SplitDataset(dataset [][]interface{}, header []string, attribute string) map[string][][]interface{} {
	subsets := make(map[string][][]interface{})

	attrIndex := -1
	for i, col := range header {
		if col == attribute {
			attrIndex = i
			break
		}
	}

	// fmt.Printf("Attribute index in header %v\n",attrIndex)

	if attrIndex == -1 {
		fmt.Println("Error: Attribute not found in header")
		return subsets
	}

	// Check the type of the attribute (categorical or numerical)
	switch dataset[0][attrIndex].(type) {
	case string:
		// fmt.Printf("dataset[0][attrIndex]%v\n",dataset[0][attrIndex])
		// Categorical split
		for _, row := range dataset {
			if attrIndex < len(row) {
				key, _ := row[attrIndex].(string)
				// fmt.Printf("row[attrIndex]%v\n",row[attrIndex])
				subsets[key] = append(subsets[key], row)
				// fmt.Println("subseets:", subsets)
			}
		}
	default:
		// Numeric or date split (find best threshold)
		bestThreshold, leftSubset, rightSubset := FindBestThreshold(dataset, attrIndex)
		subsets[fmt.Sprintf("<=%.2f", bestThreshold)] = leftSubset
		subsets[fmt.Sprintf(">%.2f", bestThreshold)] = rightSubset
		// fmt.Println("subseets:", subsets)

	}

	return subsets
}
// FindBestThreshold finds the best threshold to split a numeric or time attribute
func FindBestThreshold(dataset [][]interface{}, attrIndex int) (float64, [][]interface{}, [][]interface{}) {
	var values []float64
	for _, row := range dataset {
		if v, ok := row[attrIndex].(float64); ok {
			values = append(values, v)
		} else if v, ok := row[attrIndex].(time.Time); ok {
			values = append(values, float64(v.Unix())) // Convert time to Unix timestamp
		}
	}

	if len(values) == 0 {
		fmt.Println("Error: No numeric or time values found in dataset for given attribute")
		return 0, nil, nil
	}

	sort.Float64s(values) // Sort values to find optimal threshold
	bestThreshold := values[len(values)/2] // Use median as the threshold

	var leftSubset, rightSubset [][]interface{}
	for _, row := range dataset {
		var val float64
		if v, ok := row[attrIndex].(float64); ok {
			val = v
		} else if v, ok := row[attrIndex].(time.Time); ok {
			val = float64(v.Unix())
		}

		if val <= bestThreshold {
			leftSubset = append(leftSubset, row)
		} else {
			rightSubset = append(rightSubset, row)
		}
	}

	return bestThreshold, leftSubset, rightSubset
}


// InformationGain calculates the information gain of splitting on an attribute
func InformationGain(dataset [][]interface{}, header []string, attrIndex int) float64 {
	totalSamples := len(dataset)
	if totalSamples == 0 {
		return 0
	}

	initialEntropy := Entropy(dataset)
	subsets := make(map[string][][]interface{})
	var bestThreshold float64
	var leftSubset, rightSubset [][]interface{}

	// Check attribute type
	switch dataset[0][attrIndex].(type) {
	case string:
		// Categorical split
		for _, row := range dataset {
			if attrIndex < len(row) {
				key, _ := row[attrIndex].(string)
				subsets[key] = append(subsets[key], row)
			}
		}

	default:
		// Numeric split - Find best threshold first
		bestThreshold, leftSubset, rightSubset = FindBestThreshold(dataset, attrIndex)
		subsets[fmt.Sprintf("<=%.2f", bestThreshold)] = leftSubset
		subsets[fmt.Sprintf(">%.2f", bestThreshold)] = rightSubset
	}

	// Compute weighted entropy
	weightedEntropy := 0.0
	for _, subset := range subsets {
		proportion := float64(len(subset)) / float64(totalSamples)
		weightedEntropy += proportion * Entropy(subset)
	}

	informationGain := initialEntropy - weightedEntropy
	fmt.Printf("information gain for %v: %v\n",informationGain)
	return informationGain
}

// func BestAttributeToSplit(dataset [][]interface{}, header []string) (string, int, float64) {
// 	bestAttr := ""
// 	bestAttrIndex := -1
// 	highestGain := 0.0

// 	for i := 0; i < len(header)-1; i++ { // Exclude the last column (target variable)
// 		ig := InformationGain(dataset, header, i)
// 		fmt.Printf("Attribute: %s, Information Gain: %.4f\n", header[i], ig)

// 		if ig > highestGain {
// 			highestGain = ig
// 			bestAttr = header[i]
// 			bestAttrIndex = i
// 		}
// 	}

// 	return bestAttr, bestAttrIndex, highestGain
// }


// GainRatio calculates the gain ratio of an attribute
func GainRatio(dataset [][]interface{}, header []string, attrIndex int) float64 {
	totalSamples := len(dataset)
	if totalSamples == 0 {
		return 0
	}

	infoGain := InformationGain(dataset, header, attrIndex)
	if infoGain == 0 {
		return 0
	}

	subsets := make(map[string][][]interface{})
	var bestThreshold float64
	var leftSubset, rightSubset [][]interface{}

	// Check attribute type
	switch dataset[0][attrIndex].(type) {
	case string:
		// Categorical split
		for _, row := range dataset {
			if attrIndex < len(row) {
				key, _ := row[attrIndex].(string)
				subsets[key] = append(subsets[key], row)
			}
		}
	default:
		// Numeric split
		bestThreshold, leftSubset, rightSubset = FindBestThreshold(dataset, attrIndex)
		subsets[fmt.Sprintf("<=%.2f", bestThreshold)] = leftSubset
		subsets[fmt.Sprintf(">%.2f", bestThreshold)] = rightSubset
	}

	// Compute split information
	splitInfo := 0.0
	for _, subset := range subsets {
		proportion := float64(len(subset)) / float64(totalSamples)
		if proportion > 0 {
			splitInfo -= proportion * math.Log2(proportion)
		}
	}

	if splitInfo == 0 {
		return 0
	}

	gainRatio := infoGain / splitInfo
	return gainRatio
}

func BestAttributeByGainRatio(dataset [][]interface{}, header []string) (string, int, float64) {
	bestAttr := ""
	bestAttrIndex := -1
	highestGainRatio := 0.0

	for i := 0; i < len(header)-1; i++ { // Exclude target variable
		gr := GainRatio(dataset, header, i)
		fmt.Printf("Attribute: %s, Gain Ratio: %.4f\n", header[i], gr)

		if gr > highestGainRatio {
			highestGainRatio = gr
			bestAttr = header[i]
			bestAttrIndex = i
		}
	}

	return bestAttr, bestAttrIndex, highestGainRatio
}

// BestAttribute finds the attribute with the highest Gain Ratio and returns it.
// func BestAttribute(dataset [][]interface{}, header []string) string {
// 	bestAttr := ""
// 	bestGainRatio := -1.0

// 	for _, attr := range header[:len(header)-1] { // Exclude target variable
// 		gainRatio := GainRatio(dataset, header, attr)

// 		if gainRatio > bestGainRatio {
// 			bestGainRatio = gainRatio
// 			bestAttr = attr
// 		}
// 	}

// 	return bestAttr
// }

// Node represents a decision tree node
type Node struct {
	Attribute   string                 // Attribute used for splitting
	Children    map[string]*Node       // Child nodes (key: attribute value, value: child node)
	IsLeaf      bool                   // True if this is a leaf node
	Class       string                 // Class label (if leaf)
}

// BuildTree constructs the decision tree recursively
func BuildTree(dataset [][]interface{}, header []string) *Node {
	// Base case: If all instances belong to the same class, return a leaf node
	if allSameClass(dataset) {
		return &Node{
			IsLeaf: true,
			Class:  dataset[0][len(dataset[0])-1].(string), // Last column is the target
		}
	}

	// Find the best attribute to split on
	bestAttr, _, _ := BestAttributeByGainRatio(dataset, header)

	// Create a new node for the best attribute
	node := &Node{
		Attribute: bestAttr,
		Children:  make(map[string]*Node),
	}

	// Split the dataset based on the best attribute
	subsets := SplitDataset(dataset, header, bestAttr)

	// Recursively build the tree for each subset
	for value, subset := range subsets {
		if len(subset) == 0 {
			// If the subset is empty, create a leaf node with the majority class
			node.Children[value] = &Node{
				IsLeaf: true,
				Class:  majorityClass(dataset),
			}
		} else {
			// Recursively build the tree for the subset
			node.Children[value] = BuildTree(subset, header)
		}
	}

	return node
}


// allSameClass checks if all instances in the dataset belong to the same class
func allSameClass(dataset [][]interface{}) bool {
	if len(dataset) == 0 {
		return true
	}

	targetClass := dataset[0][len(dataset[0])-1].(string)
	for _, row := range dataset {
		if row[len(row)-1].(string) != targetClass {
			return false
		}
	}
	return true
}

// majorityClass returns the majority class in the dataset
func majorityClass(dataset [][]interface{}) string {
	classCounts := CountClassOccurrences(dataset)
	majorityClass := ""
	maxCount := 0

	for class, count := range classCounts {
		if count > maxCount {
			maxCount = count
			majorityClass = class
		}
	}
	return majorityClass
}

func main() {
	header, dataset, colTypes, err := LoadCsv("data.csv")
	if err != nil {
		fmt.Println("Error loading data from the csv file", err)
		return
	}

	fmt.Println("Header of the csv file", header)
	for i, row := range dataset {
		fmt.Printf("row number: %v\n %v\n", i, row)
	}
	totalsamples := len(dataset)
	classCount := CountClassOccurrences(dataset)
	fmt.Println("counts of yes no in data", classCount)
	probabilities := ComputeProbabilities(classCount, totalsamples)
	fmt.Println("probabilities", probabilities)
	fmt.Println("entropies:", Entropy(dataset))
	bestAttr,bestAttrInd,highestGr:=BestAttributeByGainRatio(dataset,header)
	// splitted:=SplitDataset(dataset,header,bestAttr)
	fmt.Printf("Best attribute %v\n",bestAttr)
	fmt.Printf("Highest Gain Ratio %v\n",highestGr)
	fmt.Printf("BestAttrIndex %v\n",bestAttrInd)


	// fmt.Println("Splitted dataset",splitted)

	fmt.Printf("column types: %v\n", colTypes)
}
