package main

type TreeNode struct {
	Left, Right, Parent *TreeNode
	Data                string
}

func BTreeInsertData(root *TreeNode, data string) *TreeNode {
	if root == nil {
		return &TreeNode{Data: data}
	}

	if data < root.Data {
		root.Left = BTreeInsertData(root.Left, data)
		if root.Left != nil {
			root.Left.Parent = root
		}
	}

	if data > root.Data {
		root.Right = BTreeInsertData(root.Right, data)
		if root.Right != nil {
			root.Right.Parent = root
		}
	}
	return root
}

func BTreeApplyInorder(root *TreeNode, f func(...interface{}) (int, error)) {
	if root == nil {
		return
	}
	BTreeApplyInorder(root.Left, f)
	f(root.Data)
	BTreeApplyInorder(root.Right, f)
}

func BTreeApplyPreorder(root *TreeNode, f func(...interface{}) (int, error)) {
	if root == nil {
		return
	}
	f(root.Data)
	BTreeApplyPreorder(root.Left, f)
	BTreeApplyPreorder(root.Right, f)
}

func BTreeSearchItem(root *TreeNode, elem string) *TreeNode {
	if root == nil {
		return nil
	}

	if elem == root.Data {
		return root
	}

	if elem < root.Data {
		return BTreeSearchItem(root.Left, elem)
	}
	if elem > root.Data {
		return BTreeSearchItem(root.Right, elem)
	}
	return root
}

func BTreeLevelCount(root *TreeNode) int {
	if root == nil {
		return 0
	}
	lheight := BTreeLevelCount(root.Left)
	rheight := BTreeLevelCount(root.Right)
	return max(lheight, rheight) + 1
}

func max(a, b int) int {
	max := a
	if b > a {
		max = b
	}
	return max
}
