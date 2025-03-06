package main

import (
	"encoding/csv"
	"fmt"
	"os"
)



type Data struct {
	Outlook string
	Temprature string
	Humidity string
	Wind string
	PlayTennis string
}


func LoadCsv(s string)([]string,[]Data,error){
	file,err :=os.Open(s)
	if err !=nil{
		fmt.Println("Error openning file",err)
		return nil,nil,fmt.Errorf("Error opening file %v",err)
	}
	defer file.Close()

	reader:=csv.NewReader(file)

	records,err:=reader.ReadAll()
	if err !=nil{
		fmt.Println("Error reading file:",err)
		return nil,nil,fmt.Errorf("Error reading file:%v",err)
	}

	var dataset []Data

	header:=records[0]

	for _,row:=range records[1:]{
		dataset = append(dataset,Data{
			Outlook: row[0],
			Temprature: row[1],
			Humidity: row[2],
			Wind: row[3],
			PlayTennis: row[4],
		})
	}

return header,dataset, nil

}

//CountClass counts the occurrence of the target class in 
//our dataset
func countClassOccurrences(dataset []Data) map[string]int{
	classCounts:=make(map[string]int)

	for _,row:=range dataset{
		classCounts[row.PlayTennis]++
	}

return classCounts
}

//Calculates probability of each class

func computeProbabilities(classCounts map[string]int, totalSamples int)map[string]float64{
	probabilities:=make(map[string]float64)

	for class,count:=range classCounts{
		probabilities[class]=float64(count)/float64(totalSamples)
	}
	return probabilities
}

func main(){
	_,header,err:=LoadCsv("dataset.csv")
	if err !=nil{
		fmt.Println("error openning file")
		return
	}

	files:=countClassOccurrences(header)
	totalSamples:=len(header)
	probabilities:=computeProbabilities(files,totalSamples)
	fmt.Println(files)
	fmt.Println(probabilities)
	
}