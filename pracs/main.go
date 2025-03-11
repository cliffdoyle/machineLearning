package main

import "fmt"

func main() {
	m := map[string]int{"yes": 3}

	for _,key:= range m {
		fmt.Println(key)
	}
}