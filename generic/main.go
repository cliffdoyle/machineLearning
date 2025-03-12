package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"

	// "flag"
	"fmt"
	"math"
	"os"
)

func LoadCsv(filename string) ([]string, [][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, nil, fmt.Errorf("error reading file: %v", err)
	}

	if len(records) < 2 {
		return nil, nil, fmt.Errorf("insufficient data in CSV file")
	}

	header := records[0]

	// Store rows as a slice of slices
	dataset := records[1:]

	return header, dataset, nil
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

func SplitDataset(dataset [][]string, header []string, attribute string) map[string][][]string {
	subsets := make(map[string][][]string)

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
			key := row[attrIndex]
			subsets[key] = append(subsets[key], row)
		}
	}

	return subsets
}

// How much information do we gain by using the selected attribute
func InformationGain(dataset [][]string, header []string, attribute string) float64 {
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

func GainRatio(dataset [][]string, header []string, attribute string) float64 {
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

func BestAttribute(dataset [][]string, header []string) string {
	bestAttr := ""
	bestGainRAtio := -1

	// Exclude the last column (target variable) from selection
	for i := 0; i < len(header)-1; i++ {
		attr := header[i]
		gainRatio := GainRatio(dataset, header, attr)
		if gainRatio > float64(bestGainRAtio) {
			bestGainRAtio = int(gainRatio)
			bestAttr = attr
		}
	}
	return bestAttr
}

type TreeNode struct {
	Attribute string
	Children  map[string]*TreeNode
	Class     string
	IsLeaf    bool
}

func BuildDecisionTree(dataset [][]string, header []string) *TreeNode {
	// Count occurrences of the target class (last column)
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

	// Create a new decision tree node
	node := &TreeNode{Attribute: bestAttr, Children: make(map[string]*TreeNode)}

	// Split the dataset based on the best attribute
	splitted := SplitDataset(dataset, header, bestAttr)

	for attrValue, subset := range splitted {
		node.Children[attrValue] = BuildDecisionTree(subset, header)
	}

	return node
}

// Train decision tree and save model
func TrainModel(inputFile, targetCol, outputFile string) error {
	// Load dataset
	header, dataset, err := LoadCsv(inputFile)
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

// Prediction function that loads model from json and makes prediction
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
func Predict(tree *TreeNode, instance map[string]string) string {
	if tree.IsLeaf {
		return tree.Class
	}
// fmt.Println("tree.attribute",instance[tree.Attribute])
	attributeValue, exists := instance[tree.Attribute]
	fmt.Println("exists:",exists)
	if !exists {
		return "Unknown"
	}
	// fmt.Println(attributeValue)
	child, found := tree.Children[attributeValue]
	fmt.Println(child)
	// fmt.Println(child.Children)
	if !found {
		return "Unknown"
	}

	return Predict(child, instance)
}

// Predict from test CSV using trained model
func PredictFromModel(inputFile, modelFile, outputFile string) error {
	// LOad dataset
	header, dataset, err := LoadCsv(inputFile)
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
			instance[header[i]] = value
		}

		prediction := Predict(tree, instance)
		newRow := append(row, prediction)
		writer.Write(newRow)
	}
	fmt.Println("Predictions saved to", outputFile)
	return nil
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
