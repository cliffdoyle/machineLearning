package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
	"math"
	"sort"
	"encoding/json"
	"flag"
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

	if attrIndex == -1 {
		fmt.Println("Error: Attribute not found in header")
		return subsets
	}

	// Check the type of the attribute (categorical or numerical)
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
		// Numeric or date split (find best threshold)
		bestThreshold, leftSubset, rightSubset := FindBestThreshold(dataset, attrIndex)
		subsets[fmt.Sprintf("<=%.2f", bestThreshold)] = leftSubset
		subsets[fmt.Sprintf(">%.2f", bestThreshold)] = rightSubset
	}

	return subsets
}

// FindBestThreshold finds the best threshold to split a numeric attribute
func FindBestThreshold(dataset [][]interface{}, attrIndex int) (float64, [][]interface{}, [][]interface{}) {
	var values []float64
	for _, row := range dataset {
		if v, ok := row[attrIndex].(float64); ok {
			values = append(values, v)
		} else if v, ok := row[attrIndex].(string); ok {
			parsedTime, err := time.Parse("2006-01-02", v) // Example: YYYY-MM-DD
			if err == nil {
				values = append(values, float64(parsedTime.Unix())) // Convert date to numeric value
			}
		}
	}

	sort.Float64s(values) // Sort values to find optimal threshold
	bestThreshold := values[len(values)/2]

	var leftSubset, rightSubset [][]interface{}
	for _, row := range dataset {
		val, _ := row[attrIndex].(float64)
		if val <= bestThreshold {
			leftSubset = append(leftSubset, row)
		} else {
			rightSubset = append(rightSubset, row)
		}
	}

	return bestThreshold, leftSubset, rightSubset
}

// InformationGain calculates how much information is gained by splitting on an attribute
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

	informationGain := initialEntropy - weightedEntropy
	return informationGain
}

// GainRatio calculates the gain ratio, a normalized version of information gain
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

	gainRatio := infoGain / splitInfo
	return gainRatio
}

// BestAttribute finds the attribute with the highest Gain Ratio and returns it.
func BestAttribute(dataset [][]interface{}, header []string) string {
	bestAttr := ""
	bestGainRatio := -1.0

	for _, attr := range header[:len(header)-1] { // Exclude target variable
		gainRatio := GainRatio(dataset, header, attr)

		if gainRatio > bestGainRatio {
			bestGainRatio = gainRatio
			bestAttr = attr
		}
	}

	return bestAttr
}

type TreeNode struct {
	Attribute  string
	Threshold  float64
	Children   map[string]*TreeNode
	Class      string
	IsLeaf     bool
}

// BuildDecisionTree constructs a decision tree based on the dataset.
func BuildDecisionTree(dataset [][]interface{}, header []string) *TreeNode {
	classCounts := CountClassOccurrences(dataset)

	// If all samples belong to the same class, return a leaf node
	if len(classCounts) == 1 {
		for class := range classCounts {
			return &TreeNode{Class: class, IsLeaf: true}
		}
	}

	bestAttr := BestAttribute(dataset, header)
	if bestAttr == "" {
		// If no good split is found, return the most common class
		mostCommonClass := ""
		maxCount := 0
		for class, count := range classCounts {
			if count > maxCount {
				maxCount = count
				mostCommonClass = class
			}
		}
		return &TreeNode{Class: mostCommonClass, IsLeaf: true}
	}

	attrIndex := -1
	for i, col := range header {
		if col == bestAttr {
			attrIndex = i
			break
		}
	}

	node := &TreeNode{Attribute: bestAttr, Children: make(map[string]*TreeNode)}

	// Determine whether the attribute is numeric or categorical
	switch dataset[0][attrIndex].(type) {
	case string:
		// Categorical split
		splitted := SplitDataset(dataset, header, bestAttr)
		for attrValue, subset := range splitted {
			node.Children[attrValue] = BuildDecisionTree(subset, header)
		}
	default:
		// Numeric split (find threshold)
		threshold, leftSubset, rightSubset := FindBestThreshold(dataset, attrIndex)
		node.Threshold = threshold
		node.Children[fmt.Sprintf("<=%.2f", threshold)] = BuildDecisionTree(leftSubset, header)
		node.Children[fmt.Sprintf(">%.2f", threshold)] = BuildDecisionTree(rightSubset, header)
	}

	return node
}

// Train decision tree and save model
func TrainModel(inputFile, targetCol, outputFile string) error {
	// Load dataset
	header, dataset, _, err := LoadCsv(inputFile) // Ignoring colTypes
	if err != nil {
		return err
	}

	// Train decision tree
	tree := BuildDecisionTree(dataset, header)

	// Save model as JSON
	modelFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("Error creating model file: %v", err)
	}
	defer modelFile.Close()

	encoder := json.NewEncoder(modelFile)
	err = encoder.Encode(tree)
	if err != nil {
		return fmt.Errorf("Error writing model: %v", err)
	}

	fmt.Println("Model saved to", outputFile)
	return nil
}

// Load model from JSON file
func LoadModel(modelFile string) (*TreeNode, error) {
	file, err := os.Open(modelFile)
	if err != nil {
		return nil, fmt.Errorf("Error opening model file: %v", err)
	}
	defer file.Close()

	var tree TreeNode
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&tree)
	if err != nil {
		return nil, fmt.Errorf("Error decoding model file: %v", err)
	}

	return &tree, nil
}

// Predict a single instance
func Predict(node *TreeNode, instance map[string]string) string {
	if node.IsLeaf {
		return node.Class
	}

	attrValue, exists := instance[node.Attribute]
	if !exists {
		return "Unknown"
	}

	// If value exists, navigate tree
	if child, found := node.Children[attrValue]; found {
		return Predict(child, instance)
	}

	// Fallback: If unseen value, return majority class
	return FindMostCommonClass(node)
}

func FindMostCommonClass(node *TreeNode) string {
	classCount := make(map[string]int)

	for _, child := range node.Children {
		if child.IsLeaf {
			classCount[child.Class]++
		} else {
			classCount[FindMostCommonClass(child)]++
		}
	}

	// Find most frequent class
	var mostCommonClass string
	maxCount := 0
	for class, count := range classCount {
		if count > maxCount {
			mostCommonClass = class
			maxCount = count
		}
	}
	return mostCommonClass
}


// Predict from test CSV using trained model
func PredictFromModel(inputFile, modelFile, outputFile string) error {
	// Load dataset
	header, dataset, _, err := LoadCsv(inputFile) // Ignoring colTypes
	if err != nil {
		return err
	}

	// Load model
	tree, err := LoadModel(modelFile)
	if err != nil {
		return err
	}

	// Open output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("Error creating output file: %v", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// Write header with "Prediction" column
	newHeader := append(header, "Prediction")
	writer.Write(newHeader)

	// Predict for each row
	for _, row := range dataset {
		instance := make(map[string]string)
		for i, value := range row {
			instance[header[i]] = fmt.Sprintf("%v", value) // Convert to string
		}

		prediction := Predict(tree, instance)
		newRow := append(interfaceSliceToStringSlice(row), prediction)
		writer.Write(newRow)
	}
	fmt.Println("Predictions saved to", outputFile)
	return nil
}

// Convert interface{} slice to string slice
func interfaceSliceToStringSlice(row []interface{}) []string {
	result := make([]string, len(row))
	for i, val := range row {
		result[i] = fmt.Sprintf("%v", val)
	}
	return result
}

func main() {
	// Define CLI flags
	command := flag.String("c", "", "Command: train or predict")
	inputFile := flag.String("i", "", "Input CSV file")
	targetCol := flag.String("t", "", "Target column (only for training)")
	modelFile := flag.String("m", "", "Model file (only for prediction)")
	outputFile := flag.String("o", "", "Output file")

	// Parse flags
	flag.Parse()

	// Execute command
	switch *command {
	case "train":
		if *inputFile == "" || *targetCol == "" || *outputFile == "" {
			fmt.Println("Usage: dt -c train -i <input.csv> -t <target> -o <model.dt>")
			return
		}
		err := TrainModel(*inputFile, *targetCol, *outputFile)
		if err != nil {
			fmt.Println("Error:", err)
		}

	case "predict":
		if *inputFile == "" || *modelFile == "" || *outputFile == "" {
			fmt.Println("Usage: dt -c predict -i <test.csv> -m <model.dt> -o <predictions.csv>")
			return
		}
		err := PredictFromModel(*inputFile, *modelFile, *outputFile)
		if err != nil {
			fmt.Println("Error:", err)
		}

	default:
		fmt.Println("Invalid command. Use 'train' or 'predict'.")
	}
}



// func main(){
// 	header := []string{"Color", "Size", "Weight", "Class"}
// dataset := [][]interface{}{
// 	{"Red", "Small", 1.5, "A"},
// 	{"Blue", "Large", 3.2, "B"},
// 	{"Green", "Medium", 2.1, "A"},
// 	{"Red", "Medium", 1.8, "B"},
// }

// bestAttr := BestAttribute(dataset, header)
// fmt.Println("Best attribute to split on:", bestAttr)


// tree := BuildDecisionTree(dataset, header)
// fmt.Println(tree)

// }



// // Example usage
// func main() {
// 	// Sample dataset with categorical, numerical, and date attributes
// 	dataset := [][]interface{}{
// 		{"Sunny", 85.0, "2023-01-01", "No"},
// 		{"Rainy", 75.0, "2023-01-03", "Yes"},
// 		{"Overcast", 78.0, "2023-01-05", "Yes"},
// 		{"Sunny", 90.0, "2023-01-07", "No"},
// 	}

// 	header := []string{"Weather", "Temperature", "Date", "PlayTennis"}

// 	// Test splitting
// 	splitted := SplitDataset(dataset, header, "Temperature")
// 	fmt.Println("Splitted Dataset:", splitted)

// 	infoGain := InformationGain(dataset, header, "Temperature")
// 	fmt.Println("Information Gain (Temperature):", infoGain)

// 	gainRatio := GainRatio(dataset, header, "Temperature")
// 	fmt.Println("Gain Ratio (Temperature):", gainRatio)
// }



// func main() {
	// header, dataset, colTypes, err := LoadCsv("data.csv")
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }

	// fmt.Println("Headers:", header)
	// fmt.Println("Column Types:", colTypes)
	// fmt.Println("Dataset:", dataset)

// 	// Example usage
// func main() {
// 	// Sample dataset with categorical class labels
// 	dataset := [][]interface{}{
// 		{"Sunny", 85.0, "Hot", "No"},
// 		{"Rainy", 75.0, "Cool", "Yes"},
// 		{"Overcast", 78.0, "Mild", "Yes"},
// 		{"Sunny", 90.0, "Hot", "No"},
// 	}

// 	classCounts := CountClassOccurrences(dataset)
// 	fmt.Println("Class Occurrences:", classCounts)

// 	totalSamples := len(dataset)
// 	probabilities := ComputeProbabilities(classCounts, totalSamples)
// 	fmt.Println("Class Probabilities:", probabilities)

// 	entropy := Entropy(dataset)
// 	fmt.Println("Entropy of dataset:", entropy)
// }

