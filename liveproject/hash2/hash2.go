package main

import "fmt"

/*
This hashtable is based on open addressing, which is also called closed hashing.  This is the 2nd hashing exercise.  Here we use a key's hashed value to directly
calculate the location to store its value.  In assembly, you could probably calculate a memory address.  Here in Go, we'll store the data in a big array (slice)
and calculate the key's index in this array.  We'll take the hashed key modulo the slices size to map the key to its location in the slice.
Linear probing collision resolution policy is to add a constant and repeat until you find an empty slot to store your value.
The sequence of slots to check for a vacancy is called the probe sequence.

Needed hash table functions are dump, set, get, and contains.  The next project will add delete.

*/

type Employee struct {
	name  string
	phone string
}

type LinearProbingHashTable struct {
	capacity  int
	employees []*Employee
}

func hash(value string) int {
	hash := 5381
	for _, ch := range value {
		hash = ((hash << 5) + hash) + int(ch)
	}

	if hash < 0 {
		hash = -hash
	}
	return hash
}

func NewLinearProbingHashTable(capacity int) *LinearProbingHashTable {
	LPHT := LinearProbingHashTable{ // LPHT = Linear Probing Hash Table, of course.
		capacity:  capacity,
		employees: make([]*Employee, capacity),
	}
	return &LPHT
}

func (hashTable *LinearProbingHashTable) dump() {
	for i, ht := range hashTable.employees {
		fmt.Printf(" %2d:\n", i)
		if ht == nil {
			fmt.Printf(" ---\n")
			continue
		}
		fmt.Printf(" %20s %s\n", ht.name, ht.phone)
	}
	fmt.Println()
}

func (hashTable *LinearProbingHashTable) find(name string) (int, int) {
	// return index of a key's location or the index where it should be inserted, and the probe sequence length.  If not found and table is full, return -1 for the index.
	// (A + B) mod C = ((A mod C) + (B mod C)) mod C
	h := hash(name)
	h = h % hashTable.capacity
	//fmt.Printf(" In find.  capacity = %d, h = %d\n", hashTable.capacity, h)
	var i int // keep the i variable for past the for loop
	for i = h; i < hashTable.capacity; i++ {
		if hashTable.employees[i] == nil { // key not found, but came to an empty slot so here's where a new key would be added
			return i, i - h + 1
		}
		if hashTable.employees[i].name == name { // found key
			return i, i - h + 1
		}
	}
	// If got here, name not found and i == capacity
	return -1, hashTable.capacity
}

func (hashTable *LinearProbingHashTable) set(name, phone string) {
	idx, numJumps := hashTable.find(name)
	fmt.Printf(" Set %s, %s.  idx = %d, numJumps = %d\n", name, phone, idx, numJumps)
	if idx < 0 { // hashtable is at capacity, can't add any more entries.
		panic("hashtable at capacity and cannot add any more entries")
	}
	if hashTable.employees[idx] == nil { // name is not in the table, so add it
		empl := Employee{
			name:  name,
			phone: phone,
		}
		hashTable.employees[idx] = &empl // this uses pointer semantics
		return
	}

	// Employee record is here, so need to update the phone field
	hashTable.employees[idx].phone = phone
}

func (hashTable *LinearProbingHashTable) get(name string) string {
	idx, _ := hashTable.find(name)
	if idx < 0 { // employee not found (and table is full)
		return ""
	}
	if hashTable.employees[idx] == nil { // name not found
		return ""
	}
	return hashTable.employees[idx].phone
}

func (hashTable *LinearProbingHashTable) contains(name string) bool {
	idx, _ := hashTable.find(name)
	if idx < 0 {
		return false
	}
	return hashTable.employees[idx] != nil
}

// not yet.
//func (hashTable *ChainingHashTable) delete(name string) {
//	tblIdx, empIdx := hashTable.find(name)
//	if empIdx < 0 { // name not found.  Do nothing.
//		return
//	}
//
//	// Employee record is here, so need to delete it.
//	for EmpIdx := range hashTable.buckets[tblIdx] {
//		if hashTable.buckets[tblIdx][EmpIdx].name == name {
//			slice := append(hashTable.buckets[tblIdx][:EmpIdx], hashTable.buckets[tblIdx][empIdx+1:]...)
//			hashTable.buckets[tblIdx] = slice
//		}
//	}
//
//}

// Make a display showing whether each array entry is nil.
func (hashTable *LinearProbingHashTable) dumpConcise() {
	// Loop through the array.
	for i, employee := range hashTable.employees {
		if employee == nil {
			// This spot is empty.
			fmt.Printf(".")
		} else {
			// Display this entry.
			fmt.Printf("O")
		}
		if i%50 == 49 {
			fmt.Println()
		}
	}
	fmt.Println()
}

func main() {
}

//type ChainingHashTable struct {
//	numBuckets int // could be determined using len(), but is easy and helpful here.
//	buckets    [][]*Employee
//}
//func newChainingHashTable(numBuckets int) *ChainingHashTable {
//	slice := make([][]*Employee, numBuckets)
//	for i := range slice {
//		slice[i] = []*Employee{}
//	}
//	table := ChainingHashTable{
//		numBuckets: numBuckets,
//		buckets:    slice,
//	}
//
//	return &table
//}
