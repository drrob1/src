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
	slice := make([][]*Employee, numBuckets)
	for i := range slice {
		slice[i] = []*Employee{}
	}
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
	if hashTable.numBuckets == 0 {
		return h, -1
	}
	//fmt.Printf(" In find.  numBuckets = %d, h = %d\n", hashTable.numBuckets, h)
	for i, bkt := range hashTable.buckets[h] {
		if bkt.name == name {
			return h, i
		}
	}
	return h, -1
}

func (hashTable *ChainingHashTable) set(name, phone string) {
	tblIdx, empIdx := hashTable.find(name)
	//fmt.Printf(" Set %s, %s.  tblIdx = %d, empIdx = %d\n", name, phone, tblIdx, empIdx)
	if empIdx < 0 { // add a new employee record because the employee is not here
		empl := Employee{
			name:  name,
			phone: phone,
		}
		hashTable.buckets[tblIdx] = append(hashTable.buckets[tblIdx], &empl) // this uses pointer semantics
		// Don't increment numBuckets as that doesn't change when add an entry.
		return
	}

	// Employee record is here, so need to update the phone field
	hashTable.buckets[tblIdx][empIdx].phone = phone
}

func (hashTable *ChainingHashTable) get(name string) string {
	tblIdx, empIdx := hashTable.find(name)
	if empIdx < 0 { // employee not found
		return ""
	}
	return hashTable.buckets[tblIdx][empIdx].phone
}

func (hashTable *ChainingHashTable) contains(name string) bool {
	_, empIdx := hashTable.find(name)
	return empIdx >= 0
}

func main() {
	// Make some names.
	employees := []Employee{
		Employee{"Ann Archer", "202-555-0101"},
		Employee{"Bob Baker", "202-555-0102"},
		Employee{"Cindy Cant", "202-555-0103"},
		Employee{"Dan Deever", "202-555-0104"},
		Employee{"Edwina Eager", "202-555-0105"},
		Employee{"Fred Franklin", "202-555-0106"},
		Employee{"Gina Gable", "202-555-0107"},
		Employee{"Herb Henshaw", "202-555-0108"},
		Employee{"Ida Iverson", "202-555-0109"},
		Employee{"Jeb Jacobs", "202-555-0110"},
	}

	hashTable := newChainingHashTable(10)
	for _, employee := range employees {
		hashTable.set(employee.name, employee.phone)
	}
	hashTable.dump()

	//fmt.Printf("Table contains Sally Owens: %t\n", hashTable.contains("Sally Owens"))
	//fmt.Printf("Table contains Dan Deever: %t\n", hashTable.contains("Dan Deever"))
	//fmt.Println("Deleting Dan Deever")
	//hashTable.delete("Dan Deever")
	//fmt.Printf("Table contains Dan Deever: %t\n", hashTable.contains("Dan Deever"))
	//fmt.Printf("Sally Owens: %s\n", hashTable.get("Sally Owens"))
	//fmt.Printf("Fred Franklin: %s\n", hashTable.get("Fred Franklin"))
	//fmt.Println("Changing Fred Franklin")
	//hashTable.set("Fred Franklin", "202-555-0100")
	//fmt.Printf("Fred Franklin: %s\n", hashTable.get("Fred Franklin"))
}
