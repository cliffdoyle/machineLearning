package main

import (
	"fmt"
)

func main() {
	root := &TreeNode{Data: "4"}
	BTreeInsertData(root, "1")
	BTreeInsertData(root, "7")
	BTreeInsertData(root, "5")
	height:=BTreeLevelCount(root)
	fmt.Println(height)
	// selected := BTreeSearchItem(root, "7")
	// // BTreeApplyInorder(root, fmt.Println)
	// // BTreeApplyPreorder(root,fmt.Println)
	// // fmt.Println(root.Left.Data)
	// // fmt.Println(root.Data)
	// // fmt.Println(root.Right.Left.Data)
	// // fmt.Println(root.Right.Data)
	// // selected := piscine.BTreeSearchItem(root, "7")
	// fmt.Print("Item selected -> ")
	// if selected != nil {
	// 	fmt.Println(selected.Data)
	// } else {
	// 	fmt.Println("nil")
	// }

	// fmt.Print("Parent of selected item -> ")
	// if selected.Parent != nil {
	// 	fmt.Println(selected.Parent.Data)
	// } else {
	// 	fmt.Println("nil")
	// }

	// fmt.Print("Left child of selected item -> ")
	// if selected.Left != nil {
	// 	fmt.Println(selected.Left.Data)
	// } else {
	// 	fmt.Println("nil")
	// }

	// fmt.Print("Right child of selected item -> ")
	// if selected.Right != nil {
	// 	fmt.Println(selected.Right.Data)
	// } else {
	// 	fmt.Println("nil")
	// }
}
