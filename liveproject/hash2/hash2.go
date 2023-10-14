package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

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
	// I found a problem when h is near the end of the table, and those items are not empty.  I want to cycle around before declaring the table full.
	var jumps int
	h := hash(name)
	h = h % hashTable.capacity
	//fmt.Printf(" In find.  capacity = %d, h = %d\n", hashTable.capacity, h)
	i := h
	var loopAround bool
	for {
		jumps++
		if hashTable.employees[i] == nil { // key not found, but came to an empty slot so here's where a new key would be added
			return i, jumps
		}
		if hashTable.employees[i].name == name { // found key
			return i, jumps
		}
		i++
		if i >= hashTable.capacity {
			if loopAround { // if already looped around, then the table is full.
				break
			}
			i = 0
			loopAround = true
		}
	}
	// If got here, name not found and i == capacity
	return -1, hashTable.capacity
}

func (hashTable *LinearProbingHashTable) set(name, phone string) {
	idx, _ := hashTable.find(name)
	//fmt.Printf(" Set %s, %s.  idx = %d, numJumps = %d\n", name, phone, idx, numJumps)
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

func (hashTable *LinearProbingHashTable) aveProbeSequenceLength() float32 {
	// Return the average probe sequence length for the items in the table.
	totalLength := 0
	numValues := 0
	for _, employee := range hashTable.employees {
		if employee != nil {
			_, probeLength := hashTable.find(employee.name)
			totalLength += probeLength
			numValues++
		}
	}
	return float32(totalLength) / float32(numValues)
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
	}

	hashTable := NewLinearProbingHashTable(10)
	for _, employee := range employees {
		hashTable.set(employee.name, employee.phone)
	}
	hashTable.dump()

	fmt.Printf("Table contains Sally Owens: %t\n", hashTable.contains("Sally Owens"))
	fmt.Printf("Table contains Dan Deever: %t\n", hashTable.contains("Dan Deever"))
	// fmt.Println("Deleting Dan Deever")
	// hashTable.delete("Dan Deever")
	// fmt.Printf("Table contains Dan Deever: %t\n", hashTable.contains("Dan Deever"))
	fmt.Printf("Sally Owens: %s\n", hashTable.get("Sally Owens"))
	fmt.Printf("Fred Franklin: %s\n", hashTable.get("Fred Franklin"))
	fmt.Println("Changing Fred Franklin")
	hashTable.set("Fred Franklin", "202-555-0100")
	fmt.Printf("Fred Franklin: %s\n", hashTable.get("Fred Franklin"))

	// Look at clustering.
	fmt.Println(time.Now()) // Print the time so it will compile if we use a fixed seed.
	random := rand.New(rand.NewSource(12345))
	// random := rand.New(rand.NewSource(time.Now().UnixNano())) // Initialize with a changing seed
	bigCapacity := 1009
	bigHashTable := NewLinearProbingHashTable(bigCapacity)
	numItems := int(float32(bigCapacity) * 0.9)
	for i := 0; i < numItems; i++ {
		str := fmt.Sprintf("%d-%d", i, random.Intn(1000000))
		bigHashTable.set(str, str)
	}
	bigHashTable.dumpConcise()
	fmt.Printf("Average probe sequence length: %f\n",
		bigHashTable.aveProbeSequenceLength())
}

func pause() bool {
	fmt.Print(" Pausing.  Hit <enter> to continue.  Or 'n' to exit  ")
	var ans string
	fmt.Scanln(&ans)
	ans = strings.ToLower(ans)
	if strings.Contains(ans, "n") {
		return true
	}
	return false
}

/*
Author's Solution
func (hash_table *LinearProbingHashTable) find(name string) (int, int) {
    // Return the key's index or where it would be if present and the probe sequence length.
    // If the key is not present and the table is full, return -1 for the index.

    // Hash the key.
    hash := hash(name) % hash_table.capacity

    // Probe up to hash_table.capacity times.
    for i := 0; i < hash_table.capacity; i++ {
        index := (hash + i) % hash_table.capacity // this handles wrap around w/ the modulo operation.  Neat!

        // If this spot is empty, the value isn't in the table.
        if hash_table.employees[index] == nil {
            return index, i + 1
        }

        // If this cell holds the key, return its data.
        if hash_table.employees[index].name == name {
            return index, i + 1
        }

        // Otherwise continue the loop.
    }

    // If we get here, then the key is not
    // in the table and the table is full.
    return -1, hash_table.capacity
}

*/
