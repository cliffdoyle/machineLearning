package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
)

type Data struct {
	Outlook    string
	Temprature string
	Humidity   string
	Wind       string
	PlayTennis string
}

func LoadCsv(s string) ([]string, []Data, error) {
	file, err := os.Open(s)
	if err != nil {
		fmt.Println("Error openning file", err)
		return nil, nil, fmt.Errorf("Error opening file %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, nil, fmt.Errorf("Error reading file:%v", err)
	}

	var dataset []Data

	header := records[0]

	for _, row := range records[1:] {
		dataset = append(dataset, Data{
			Outlook:    row[0],
			Temprature: row[1],
			Humidity:   row[2],
			Wind:       row[3],
			PlayTennis: row[4],
		})
	}

	return header, dataset, nil
}

// CountClass counts the occurrence of the target class in
// our dataset
func countClassOccurrences(dataset []Data) map[string]int {
	classCounts := make(map[string]int)

	for _, row := range dataset {
		classCounts[row.PlayTennis]++
	}

	return classCounts
}

// Calculates probability of each class
func computeProbabilities(classCounts map[string]int, totalSamples int) map[string]float64 {
	probabilities := make(map[string]float64)

	for class, count := range classCounts {
		probabilities[class] = float64(count) / float64(totalSamples)
	}
	return probabilities
}

// Split the dataset based on the attribute
func SplitDataset(dataset []Data, attribute string) map[string][]Data {
	subsets := make(map[string][]Data)

	for _, row := range dataset {
		var key string
		switch attribute {
		case "Outlook":
			key = row.Outlook
		case "Temperature":
			key = row.Temprature
		case "Humidity":
			key = row.Humidity
		case "Wind":
			key = row.Wind

		}
		subsets[key] = append(subsets[key], row)
	}
	// fmt.Println(subsets)
	return subsets
}

// Calculates entropy based on probabilities to determine the impurity of the dataset
func Entropy(dataset []Data) float64 {
	countClassOccurrences := countClassOccurrences(dataset)
	// fmt.Println(countClassOccurrences)
	totalSamples := len(dataset)
	probabilities := computeProbabilities(countClassOccurrences, totalSamples)
	// fmt.Println(probabilities)
	entropy := 0.0

	for _, probability := range probabilities {
		if probability > 0 {
			entropy -= probability * math.Log2(probability)
		}
	}
	return entropy
}

func InformationGain(dataset []Data, attribute string) float64 {
	splitted := SplitDataset(dataset, attribute)
	totalSamples := len(dataset)
	entropy := Entropy(dataset)

	weightedEntropy := 0.0
	// count := 0
	for _, subset := range splitted {
		// fmt.Println(count)
		// count++
		// fmt.Println(subset)
		proportion := float64(len(subset)) / float64(totalSamples)
		weightedEntropy += proportion * Entropy(subset)
	}

	infogain := entropy - weightedEntropy
	return infogain
}

// Gain ratio function calculates the split information of the target attribute
func GainRatio(dataset []Data, attribute string) float64 {
	splitted := SplitDataset(dataset, attribute)
	// fmt.Println("Splitted subsets:", splitted)


	totalSamples := len(dataset)

	infoGain := InformationGain(dataset, attribute)
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

// function to find the best attribute for splitting
func BestAttribute(dataset []Data,header []string) string {
	// fmt.Println("Loaded Headers:", header)
	bestAttr := ""
	bestGainRatio := -1.0

	for _, attr := range header[:len(header)-1] {
		gainRatio := GainRatio(dataset, attr)
		// fmt.Printf("Gain Ratio for %s: %f\n", attr, gainRatio)
		if gainRatio > bestGainRatio {
			bestGainRatio = gainRatio
			// fmt.Printf("best gain ratio %f",bestGainRatio)
			bestAttr = attr

		}

	}

	return bestAttr
}

type TreeNode struct{
	Attribute string
	Children map[string]*TreeNode
	Class string
	IsLeaf bool
}

func BuildDecisionTree(dataset []Data,header []string) *TreeNode{
	classCounts:=countClassOccurrences(dataset)
	if len(classCounts)==1{
		for class:= range classCounts{
			return &TreeNode{Class:class,IsLeaf:true}
		}
		
		
	}
	bestAttr:=BestAttribute(dataset,header)
	if bestAttr==""{
		mostCommonClass:=""
		maxCount:=0
		for class,count:=range classCounts{
			if count > maxCount{
				maxCount=count
				mostCommonClass=class
			}
		}
		return &TreeNode{Class:mostCommonClass,IsLeaf:true}
	}
	node:=&TreeNode{Attribute:bestAttr,Children:make(map[string]*TreeNode)}
	splitted:=SplitDataset(dataset,bestAttr)
	for attrValue,subset:=range splitted{
		node.Children[attrValue]=BuildDecisionTree(subset,header)
	}
	return node
}

// PrintDecisionTree prints the tree structure
func PrintDecisionTree(node *TreeNode, indent string) {
	if node.IsLeaf {
		fmt.Println(indent + "Class: " + node.Class)
		return
	}
	fmt.Println(indent + "Attribute: " + node.Attribute)
	for value, child := range node.Children {
		fmt.Println(indent + "  ├── Value:", value)
		PrintDecisionTree(child, indent+"  |  ")
	}
}


func main() {
	 header,dataset, err := LoadCsv("dataset.csv")
	if err != nil {
		fmt.Println("error openning file")
		return
	}

	tree := BuildDecisionTree(dataset, header)
	fmt.Println("Decision Tree Structure:")
	PrintDecisionTree(tree, "")
	
	// bestAttribute:=BestAttribute(dataset,header)

	// fmt.Printf("best attribute for our dataset is %v\n",bestAttribute)
	// fmt.Println(heado)

	// files := countClassOccurrences(header)
	// totalSamples := len(header)
	// probabilities := computeProbabilities(files, totalSamples)
	// splitted:=SplitDataset(header,"outlook")
	// entropy:=0.0
	// for _,value:=range splitted{
	// 	// fmt.Println(value)
	// 	entropy:=Entropy(value)
	// 	fmt.Printf("%.2f\n",entropy)
	// }

	// infoGain := InformationGain(dataset, "Outlook")
	// fmt.Printf("Information gain for outlook attribute is:%.2f\n", infoGain)
}
