package main

import "fmt"

/*
  Tower of Hanoi problem.  Minimum # of moves is 2**N - 1.  Wikipedia.org has a long article.
  Recursive solution
  Label the pegs A, B and C.
  Number the disks from 1 .. n
  Need to move m disks from a source peg to a target peg using a spare peg.  Only can move 1 disk at a time, no disk can be placed upon a smaller disk, and
  a move consists of taking the top disk from a stack and placing it on another peg, which can be empty.

  Here we'll treat the posts as stacks.
  We'll use a slice of slices called posts, and each posts entry will be an int.
*/

const numDisks = 3
const numPosts = 3

//func push(post []int, disk int) []int {
//	post = append(post, disk)
//	return post
//}

func push(post []int, disk int) []int {
	return append([]int{disk}, post...)
}

//func pop(post []int) (int, []int) {
//	disk := post[len(post)-1]
//	post = post[:len(post)-1] // this changes the length of this stack, so must return it as a param since it's not being returned as a pointer receiver.
//	return disk, post
//}

func pop(post []int) (int, []int) {
	return post[0], post[1:]
}

func moveDisk(posts [][]int, fromPost, toPost int) {
	disk, post := pop(posts[fromPost])
	posts[fromPost] = post
	posts[toPost] = push(posts[toPost], disk)
}

func drawPosts(posts [][]int) {
	// make all posts the same length by pushing zeros
	for i := 0; i < numPosts; i++ {
		for len(posts[i]) < numDisks {
			posts[i] = push(posts[i], 0)
		}
	}

	// display the posts, row by row
	for r := 0; r < numDisks; r++ {
		for p := range posts[r] {
			fmt.Printf(" %d", posts[p][r])
		}
		fmt.Println()
	}
	fmt.Printf(" ------\n")

	// remove the zeros
	for p := 0; p < 3; p++ {
		for len(posts[p]) > 0 && posts[p][0] == 0 {
			_, posts[p] = pop(posts[p])
		}
	}
}

func moveDisks(posts [][]int, numToMove, fromPost, toPost, tempPost int) {
	// numToMove is the number of disks to move fromPost toPost.
	// fromPost is the index of the fromPost
	// toPost is the index of the toPost
	// tempPost is the index of the tempPost
	// I see a problem w/ posts, as moving is a push and a pop which changes the length of each slice, but that change in length is not returned as I can see.  This may not matter.

	if numToMove > 1 {
		moveDisks(posts, numToMove-1, fromPost, tempPost, toPost)
	}
	moveDisk(posts, fromPost, toPost)
	drawPosts(posts)
	if numToMove > 1 {
		moveDisks(posts, numToMove-1, tempPost, toPost, fromPost)
		//drawPosts(posts)
	}
}

func main() {
	// Make three posts.
	//posts := [][]int { }
	posts := make([][]int, numPosts)

	// Push the disks onto post 0 biggest first.
	posts = append(posts, []int{})
	for disk := numDisks; disk > 0; disk-- {
		posts[0] = push(posts[0], disk)
	}

	// Make the other posts empty.
	for p := 1; p < 3; p++ {
		posts = append(posts, []int{})
	}

	fmt.Println()
	// Draw the initial setup.
	drawPosts(posts)

	// Move the disks.
	moveDisks(posts, numDisks, 0, 1, 2)
}
