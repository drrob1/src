package main

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"strings"
)

/*
  Nodes w/ no children are called leaf nodes.  And nodes w/ at least one child are called internal nodes.
  The inOrder, preOrder and postOrder traversal functions use recursion.
  The breadthFirst traversal function, uses a queue for the nodes and not recursion.

  Now called sortedBinaryTree.  A sorted tree is one in which Left child < node < right child.  This looks like the in order traversal from the previous lesson.
  I need to write insertValue, and findValue methods for this part of the live project.
*/

type Node struct {
	data        string
	left, right *Node
}

var foundValue *Node

func (root *Node) insertValue(value string) {
	// I have to first find the correct position to insert this new value.  If the new value is smaller than root, go down the left child.  If the new value is greater than root,
	// go down the right child.  If the child value is nil, insert there.  I'm going to use recursion to do this, which is easier than iteration.  He doesn't say what to do if the new value
	// is equal to root.  I'll add it to the right if equal.
	if value < root.data {
		if root.left == nil {
			node := Node{ // one way to construct a node.
				data:  value,
				left:  nil,
				right: nil,
			}
			root.left = &node
			return
		}
		again := root.left
		again.insertValue(value)
	} else {
		if root.right == nil {
			node := Node{} // another way to construct a node.
			node.data = value
			root.right = &node
			return
		}
		again := root.right
		again.insertValue(value)
	}
}

func (root *Node) findValue(value string) *Node {
	// I'm having a very hard time figuring out what to do w/ the rest of the routine after the recursion is already called.  I solved it here by using a global to handle the result.
	// The iterative solution does not have this problem.
	// I think I figured this out.  Recursion works well when the last line is the return statement.  So I have to re-write this so that the last line is return *Node or return nil.
	// Nope, I'm still stuck w/ the case of what to do after the recursion finished.  I'll try returning nil, as below.
	// So far, it seems to be working, that only the last return root is the result that is passed to the caller, the earlier return nil doesn't cause problems.  I have to check for
	// root == nil first, so I don't dereference a nil pointer when I do root.data, root.left or root.right.
	// Looks like this works.  Yay!  The first return root statement is what the caller gets as its result.  The subsequent return nil statements as the recursion unwinds are ignored.
	// IE, the return nil statements after the recursion call do not affect the result returned by this function.
	//
	fmt.Printf(" in findValue: value = %s, root = %+v\n", value, root)
	if root == nil {
		fmt.Printf(" took root == nil branch.  Value = %s, root = %+v\n", value, root)
		foundValue = nil
		return nil
	}
	if value < root.data {
		root.left.findValue(value)
		return nil
	} else if value > root.data {
		root.right.findValue(value)
		return nil
	}
	if value != root.data { // if get here, this should never be true.
		s := fmt.Sprintf(" took value != root branch.  Value = %s, root = %+v\n", value, root)
		panic(s)
	}
	foundValue = root
	return root
}

func (root *Node) findValueIterative(value string) *Node {
	n := root
	for {
		if n == nil {
			foundValue = nil
			break
		} else if value == n.data {
			foundValue = n // this is here for completeness in the global I had to create for the recursive findValue.
			break
		} else if value < n.data {
			n = n.left
		} else if value > n.data {
			n = n.right
		}
	}
	return n
}

func buildTree() *Node { // from binary tree live project.  I don't use this here, but I want to keep it so I can refer to it if needed.
	// Build a tree containing nodes A .. J and creating the tree in figure 1 of the exercise.  Note that pointer fields are initialized to nil, so I don't need to do that.
	// Need recursive functions to traverse the tree to do anything useful.  Basically, do something useful with a node, and then call that function recursively to do same w/ node's children.
	// make the nodes
	aNode := Node{
		data:  "A",
		left:  nil,
		right: nil,
	}
	bNode := Node{
		data:  "B",
		left:  nil,
		right: nil,
	}
	cNode := Node{data: "C"}
	dNode := Node{
		data: "D",
	}
	eNode := Node{data: "E"}
	fNode := Node{data: "F"}
	gNode := Node{data: "G"}
	hNode := Node{data: "H"}
	iNode := Node{data: "I"}
	jNode := Node{data: "J"}

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

func (nod *Node) displayIndented(indent string, depth int) string { // acts on the current node before it visits the children.  So it's a pre-order traversal.
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

func (nod *Node) preOrder() string {
	var sbuild strings.Builder

	sbuild.WriteString(nod.data)

	if nod.left != nil {
		sbuild.WriteRune(' ')
		sbuild.WriteString(nod.left.preOrder())
	}
	if nod.right != nil {
		sbuild.WriteRune(' ')
		sbuild.WriteString(nod.right.preOrder())
	}

	return sbuild.String()
}

func (nod *Node) inOrder() string {
	var sbuild strings.Builder

	if nod.left != nil {
		sbuild.WriteString(nod.left.inOrder())
		sbuild.WriteRune(' ')
	}

	sbuild.WriteString(nod.data)

	if nod.right != nil {
		sbuild.WriteRune(' ')
		sbuild.WriteString(nod.right.inOrder())
	}

	return sbuild.String()
}

func (nod *Node) postOrder() string {
	var sbuild strings.Builder

	if nod.left != nil {
		sbuild.WriteString(nod.left.postOrder())
		sbuild.WriteRune(' ')
	}

	if nod.right != nil {
		sbuild.WriteString(nod.right.postOrder())
		sbuild.WriteRune(' ')
	}

	sbuild.WriteString(nod.data)

	return sbuild.String()
}

func (nod *Node) breadthFirst() string {
	var sb strings.Builder
	queue := makeDoublyLinkedList()
	node := nod

	queue.enqueue(*node)

	for !queue.isEmpty() {
		node := queue.dequeue()
		sb.WriteString(node.data)
		if node.left != nil {
			queue.enqueue(*node.left)
		}
		if node.right != nil {
			queue.enqueue(*node.right)
		}
		if !queue.isEmpty() {
			sb.WriteRune(' ')
		}
	}

	return sb.String()
}

func oldmain() {
	aNode := buildTree()

	fmt.Println(aNode.displayIndented("  ", 0))
	fmt.Println()
	fmt.Println(" PreOrder:", aNode.preOrder())
	fmt.Println()
	fmt.Println(" InOrder:", aNode.inOrder())
	fmt.Println()
	fmt.Println(" PostOrder:", aNode.postOrder())
	fmt.Println()
	fmt.Println(" BreadthFirst:", aNode.breadthFirst())

}

func main() {
	oldmain()
	fmt.Printf("\n\n\n Time to run new main code.\n")

	// Make a root node to act as sentinel.
	root := Node{"", nil, nil}

	// Add some values.
	root.insertValue("I")
	//fmt.Printf(" after I: Sorted values: %s\n", root.right.inOrder())
	root.insertValue("G")
	//fmt.Printf(" after G: Sorted values: %s\n", root.right.inOrder())
	root.insertValue("C")
	//fmt.Printf(" after C: Sorted values: %s\n", root.right.inOrder())
	root.insertValue("E")
	//fmt.Printf(" after E: Sorted values: %s\n", root.right.inOrder())
	root.insertValue("B")
	//fmt.Printf(" after B: Sorted values: %s\n", root.right.inOrder())
	root.insertValue("K")
	//fmt.Printf(" after K: Sorted values: %s\n", root.right.inOrder())
	root.insertValue("S")
	//fmt.Printf(" after S: Sorted values: %s\n", root.right.inOrder())
	root.insertValue("Q")
	//fmt.Printf(" after Q: Sorted values: %s\n", root.right.inOrder())
	root.insertValue("M")
	//fmt.Printf(" after M: Sorted values: %s\n", root.right.inOrder())

	// Add F.
	root.insertValue("F")

	// Display the values in sorted order.
	fmt.Printf(" after F: Sorted values: %s\n", root.right.inOrder())

	// Let the user search for values.
	for {
		// Get the target value.
		target := ""
		fmt.Printf("String: ")
		fmt.Scanln(&target)
		if len(target) == 0 {
			break
		}

		// Find the value's node.
		target = strings.ToUpper(target)
		node := root.findValue(target)
		if foundValue == nil {
			ctfmt.Printf(ct.Red, false, "%s not found using findValue.  foundValue = %+v, node = %+v\n", target, foundValue, node)
		} else {
			ctfmt.Printf(ct.Green, false, "Found value %s using findValue in foundvalue = %+v,  node = %+v\n\n", target, foundValue, node)
		}

		node = root.findValueIterative(target)
		if node == nil {
			ctfmt.Printf(ct.Red, true, "%s not found using findValueIterative, node = %+v\n", target, node)
		} else {
			ctfmt.Printf(ct.Green, true, "Found value %s using findValueIterative in node = %+v\n\n", target, node)
		}
	}
}

type Cell struct {
	data       Node
	prev, next *Cell
}

type doublyLinkedList struct {
	topSentinel, bottomSentinel *Cell
}

func makeDoublyLinkedList() doublyLinkedList {
	// This is a factory function to create a new doubly linked list, init the sentinel pointers to a new cell, and return the linked list.
	// This is a linked list of nodes.
	var fCell, bCell Cell
	list := doublyLinkedList{}
	fNode := Node{ // one way to init a node.
		data:  "FrontSentinal",
		left:  nil,
		right: nil,
	}
	bNode := Node{} // another way to init a node.
	bNode.data = "BackSentinal"
	fCell.data = fNode
	bCell.data = bNode
	fCell.next = &bCell
	bCell.prev = &fCell
	list.topSentinel = &fCell
	list.bottomSentinel = &bCell
	return list
}

func (left *Cell) addAfter(btwn *Cell) {
	right := left.next // right now is a pointer to the initial right hand cell, so btwn can be inserted in btwn left and right cells.
	btwn.next = right  // so the next field of "btwn" now points to wherever "left" was pointing to, ie, the element after btwn that I'm calling "right".  So this inserts "btwn" between "left" and "right".
	btwn.prev = left
	left.next = btwn
	right.prev = btwn
	// no change to left.prev and right.next
	// fmt.Printf(" after assignment: me = %+v, btwnCell = %+v, right = %+v\n", btwn, btwn, right)
}

func (right *Cell) addBefore(btwn *Cell) {
	left := right.prev
	left.addAfter(btwn)
}

func (queue *doublyLinkedList) push(node Node) { // I'm going to change this so push and pop occur at the top of the list.  That's not what I did at first.
	cell := Cell{
		data: node,
		prev: nil,
		next: nil,
	}
	queue.topSentinel.addAfter(&cell)
}

func (me *Cell) delete() { // need to return the deleted cell.  If there is no cell after "me", panic.
	//fmt.Printf(" in delete.  me.data = %q\n", me.data)
	left := me.prev
	right := me.next
	left.next = right
	right.prev = left
}

func (queue *doublyLinkedList) enqueue(node Node) {
	queue.push(node)
}

func (queue *doublyLinkedList) dequeue() Node {
	cell := queue.bottomSentinel.prev
	cell.delete()
	return cell.data
}

func (queue *doublyLinkedList) length() int { // hear so I can use the _test functions
	if queue.topSentinel.next == queue.bottomSentinel {
		return 0
	}

	var counter int // this starts as zero, so I don't intend to count the sentinel elements.
	for cell := queue.topSentinel.next; cell != queue.bottomSentinel; cell = cell.next {
		counter++
	}

	return counter
}

func (list *doublyLinkedList) toSlice() []string {
	var stringSlice []string

	if list.isEmpty() {
		return []string{}
	}

	for cell := list.topSentinel.next; cell != list.bottomSentinel; cell = cell.next {
		stringSlice = append(stringSlice, cell.data.data)
	}

	return stringSlice
}

func (list *doublyLinkedList) isEmpty() bool {
	return list.topSentinel.next == list.bottomSentinel
}

func (list *doublyLinkedList) toString(separator string) string {
	var sb strings.Builder

	if list.isEmpty() {
		return ""
	}

	for cell := list.topSentinel.next; cell != list.bottomSentinel; cell = cell.next {
		//fmt.Printf("cell data %s, ", cell.data)
		sb.WriteString(cell.data.data)
		if cell.next != list.bottomSentinel { // I got this solution from the course hint.
			sb.WriteString(separator)
		}
	}
	return sb.String()
}

/*
func (root *Node) findValue(value string) *Node {
	// I'm having a very hard time figuring out what to do w/ the rest of the routine after the recursion is already called.  I solved it here by using a global to handle the result.
	// The iterative solution does not have this problem.
	// I think I figured this out.  Recursion works well when the last line is the return statement.  So I have to re-write this so that the last line is return *Node.
    // I'm preserving the version of this funcion before I started refactoring it.
	fmt.Printf(" in findValue: value = %s, root = %+v\n", value, root)
	if root == nil {
		fmt.Printf(" took root == nil branch.  Value = %s, root = %+v\n", value, root)
		foundValue = nil
		return nil
	}
	if value == root.data {
		fmt.Printf(" took value == root branch.  Value = %s, root = %+v\n", value, root)
		foundValue = root
		return root
	}
	if value < root.data {
		root.left.findValue(value)
	} else if value > root.data {
		root.right.findValue(value)
	}
	return nil
}
*/
