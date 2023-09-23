package main

import (
	"fmt"
	"strings"
)

/*
  Nodes w/ no children are called leaf nodes.  And nodes w/ at least one child are called internal nodes.
*/

type node struct {
	data        string
	left, right *node
}

func buildTree() *node {
	// Build a tree containing nodes A .. J and creating the tree in figure 1 of the exercise.  Note that pointer fields are initialized to nil, so I don't need to do that.
	// Need recursive functions to traverse the tree to do anything useful.  Basically, do something useful with a node, and then call that function recursively to do same w/ node's children.
	// make the nodes
	aNode := node{
		data:  "A",
		left:  nil,
		right: nil,
	}
	bNode := node{
		data:  "B",
		left:  nil,
		right: nil,
	}
	cNode := node{data: "C"}
	dNode := node{
		data: "D",
	}
	eNode := node{data: "E"}
	fNode := node{data: "F"}
	gNode := node{data: "G"}
	hNode := node{data: "H"}
	iNode := node{data: "I"}
	jNode := node{data: "J"}

	// build the tree
	aNode.left = &bNode
	aNode.right = &cNode
	bNode.left = &dNode
	bNode.right = &eNode
	cNode.right = &fNode
	eNode.left = &gNode
	fNode.left = &hNode
	hNode.left = &iNode
	hNode.right = &jNode

	// return root node of the tree
	return &aNode
}

func (nod *node) displayIndented(indent string, depth int) string {
	var sbuild strings.Builder

	sbuild.WriteString(strings.Repeat(indent, depth))
	sbuild.WriteString(nod.data)
	sbuild.WriteRune('\n')

	if nod.left != nil {
		sbuild.WriteString(nod.left.displayIndented(indent, depth+1))
	}
	if nod.right != nil {
		sbuild.WriteString(nod.right.displayIndented(indent, depth+1))
	}

	return sbuild.String()
}

func main() {
	aNode := buildTree()

	fmt.Println(aNode.displayIndented("  ", 0))
}
