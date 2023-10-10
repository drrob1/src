package main

import "fmt"

/*
This hashtable will use a collision resolution method called chaining.  The main slice holds buckets, and the mapping hash function maps a key to a bucket.  Collision resolution
adds this to the end of the bucket at that mapped location.  As these items in each bucket are sort of chained together, this strategy is called chaining.

Needed hash table functions are dump, set, get, contains and delete.

*/

type Employee struct {
	name  string
	phone string
}

type ChainingHashTable struct {
	numBuckets int // could be determined using len(), but is easy and helpful here.
	buckets    [][]*Employee
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

func newChainingHashTable(numBuckets int) *ChainingHashTable {
	slice := make([][]*Employee, 0, numBuckets)
	table := ChainingHashTable{
		numBuckets: numBuckets,
		buckets:    slice,
	}

	return &table
}

func (hashTable *ChainingHashTable) dump() {
	for i := 0; i < hashTable.numBuckets; i++ {
		fmt.Printf(" Bucket %2d:\n", i)
		if hashTable.buckets[i] != nil {
			for _, h := range hashTable.buckets[i] {
				fmt.Printf("    %s: %s\n", h.name, h.phone)
			}
		}

	}
	fmt.Println()
}

func (hashTable *ChainingHashTable) find(name string) (int, int) { // return bucket number and employee number within that bucket.
	h := hash(name)
	h = h % hashTable.numBuckets
	for i, bkt := range hashTable.buckets[h] {
		if bkt.name == name {
			return h, i
		}
	}
	return h, -1
}

func (hashTable *ChainingHashTable) set(name, phone string) {
	tblIdx, empIdx := hashTable.find(name)
	if empIdx < 0 { // add a new employee record because the employee is not here
		empl := Employee{
			name:  name,
			phone: phone,
		}
		hashTable.buckets[tblIdx] = append(hashTable.buckets[tblIdx], &empl) // this uses pointer semantics
	}

	// Employee record is here, so need to update the phone field
	hashTable.buckets[tblIdx][empIdx].phone = phone
}

func main() {

}
