package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

/*
This hashtable is based on open addressing, which is also called closed hashing.  This is the 5th hashing exercise, and will introduce double hashing.

The main downside to quadratic probing is that it doesn't always visit every slot.

Here we use a key's hashed value to directly calculate the location to store its value.  In assembly, you could probably calculate a memory address.
Here in Go, we'll store the data in a big array (slice) and calculate the key's index in this array.  We'll take the hashed key modulo the slices size to map the key to its location in the slice.
Linear probing collision resolution policy added a constant and repeat until you find an empty slot to store your value.
The sequence of slots to check for a vacancy is called the probe sequence.

The 3rd project added delete.  The complexity of delete was that this entry needs to be non-nil because other elements may depend on this being non-nil for their probe sequence to succeed.
This is why it's hard to delete an element from this type of hashing.
This example will use lazy deletion, in which an item is marked as deleted so the probe sequence is still intact, but adding a new entry (that's not already present) can occupy a deleted slot.

The 4th project added quadratic probing.
This one, the 5th project, is adding double hashing, which needs 2 independent hash functions.  This eliminates both primary clustering and not always checking every slot.
The formula is idx = (hash1 + i * hash2) mod capacity.  And hash2 and capacity must be relatively prime.

If the capacity is a power of 2, then hash2 must be odd.

Or, if the capacity is a prime number, then hash2 must not be zero; but it can be 1.  This is the approach he uses here.  When initializing the table, I look for the smallest prime number
that is larger than the desired capacity.

The crypto algorithm RSA uses prime and relatively prime numbers.  That's covered in more detail in a subsequent exercise.
*/

type Employee struct {
	name    string
	phone   string
	deleted bool // flag to allow lazy deletion.
}

type DoubleHashTable struct {
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

func hash2(value string) int {
	// Jenkins one_at_a_time hash function.
	// See https://en.wikipedia.org/wiki/Jenkins_hash_function
	hash := 0
	for _, ch := range value {
		hash += int(ch)
		hash += hash << 10
		hash ^= hash >> 6
	}

	// Make sure the result is non-negative.
	if hash < 0 {
		hash = -hash
	}

	// Make sure the result is not 0.
	if hash == 0 {
		hash = 1
	}
	return hash
}

func NewDoubleHashTable(capacity int) *DoubleHashTable {
	c := capacity
	if c%2 == 0 {
		c++
	}

	var n int
	for n = c; !IsPrime(n); n += 2 {
	} // this should finish w/ the smallest prime number that is greater than the entered capacity.

	LPHT := DoubleHashTable{ // LPHT = Linear Probing Hash Table, of course.
		capacity:  n,
		employees: make([]*Employee, n),
	}
	return &LPHT
}

func (hashTable *DoubleHashTable) dump() {
	for i, ht := range hashTable.employees {
		fmt.Printf(" %2d:", i)
		if ht == nil {
			fmt.Printf(" ---\n")
			continue
		}
		if ht.deleted {
			fmt.Printf(" xxx\n")
			continue
		}
		fmt.Printf(" %20s %s\n", ht.name, ht.phone)
	}
	fmt.Println()
}

func (hashTable *DoubleHashTable) find(name string) (int, int) {
	// return index of a key's location or the index where it should be inserted, and the probe sequence length.  If not found and table is full, return -1 for the index.
	// (A + B) mod C = ((A mod C) + (B mod C)) mod C
	// I found a problem when h is near the end of the table, and those items are not empty.  I want to cycle around before declaring the table full.
	// In this 3rd example, deleted entries are allowed.  If a deleted entry is found, record its index and keep searching.  If name is not found, return the deleted index.
	var jumps int
	deletedIndex := -1
	h1 := hash(name)
	h1 = h1 % hashTable.capacity
	h2 := hash2(name)
	h2 = h2 % hashTable.capacity
	//fmt.Printf(" In find.  capacity = %d, h = %d\n", hashTable.capacity, h)

	for i := 0; i < hashTable.capacity; i++ {
		idx := (h1 + i*h2) % hashTable.capacity // now it's a quadratic.  Note that the first time thru, i = 0, so i^2 is also zero.
		jumps++
		if hashTable.employees[idx] == nil { // key not found, but came to an empty slot so here's where a new key may be added, but if there's a deleted Index, return that.
			if deletedIndex != -1 {
				return deletedIndex, jumps
			}
			return idx, jumps
		}
		if hashTable.employees[idx].deleted { // this will return the first deleted entry in a probe sequence if the key is not found.
			if deletedIndex == -1 {
				deletedIndex = idx
			}
		} else if hashTable.employees[idx].name == name { // found key and it's not been deleted.
			return idx, jumps
		}
	}
	// If got here, name not found and i == capacity
	if deletedIndex != -1 { // return the index of a deleted entry, so the table is not actually full when you consider deleted entries.
		return deletedIndex, hashTable.capacity
	}
	return -1, hashTable.capacity
}

func (hashTable *DoubleHashTable) set(name, phone string) {
	idx, _ := hashTable.find(name)
	//fmt.Printf(" Set %s, %s.  idx = %d, numJumps = %d\n", name, phone, idx, numJumps)
	if idx < 0 { // hashtable is at capacity, can't add any more entries.
		panic("hashtable at capacity and cannot add any more entries")
	}
	if hashTable.employees[idx] == nil || hashTable.employees[idx].deleted { // name is not in the table, so add it
		empl := Employee{
			name:  name,
			phone: phone,
			//deleted: false,  this is the default value, so don't need to explicitly set it.
		}
		hashTable.employees[idx] = &empl // this uses pointer semantics
		return
	}

	// Employee record is here, so need to update the phone field
	hashTable.employees[idx].phone = phone
}

func (hashTable *DoubleHashTable) get(name string) string {
	idx, _ := hashTable.find(name)
	if idx < 0 { // employee not found (and table is full)
		return ""
	}
	if hashTable.employees[idx] == nil || hashTable.employees[idx].deleted { // name not found
		return ""
	}
	return hashTable.employees[idx].phone
}

func (hashTable *DoubleHashTable) contains(name string) bool {
	idx, _ := hashTable.find(name)
	if idx < 0 {
		return false
	}
	return !(hashTable.employees[idx] == nil || hashTable.employees[idx].deleted)
}

func (hashTable *DoubleHashTable) delete(name string) {
	idx, _ := hashTable.find(name)
	if idx < 0 { // name not found.  Do nothing.
		return
	}

	// Employee record is here, so need to delete it.
	if hashTable.employees[idx] != nil {
		hashTable.employees[idx].deleted = true
	}
}

// Make a display showing whether each array entry is nil.
func (hashTable *DoubleHashTable) dumpConcise() {
	// Loop through the array.
	for i, employee := range hashTable.employees {
		if employee == nil {
			// This spot is empty.
			fmt.Printf(".")
		} else if employee.deleted {
			fmt.Printf("x")
		} else {
			fmt.Printf("O")
		}
		if i%50 == 49 {
			fmt.Println()
		}
	}
	fmt.Println()
}

func (hashTable *DoubleHashTable) aveProbeSequenceLength() float32 {
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

func (hashTable *DoubleHashTable) probe(name string) int {
	// Show this key's probe sequence.
	// Hash the key.
	h1 := hash(name) % hashTable.capacity
	h2 := hash2(name) % hashTable.capacity
	fmt.Printf("Probing %s (%d, %d)\n", name, h1, h2)

	// Keep track of a deleted spot if we find one.
	deletedIndex := -1

	// Probe up to hashTable.capacity times.
	for i := 0; i < hashTable.capacity; i++ {
		index := (h1 + i*h2) % hashTable.capacity // now has double hashing

		fmt.Printf("    %d: ", index)
		if hashTable.employees[index] == nil {
			fmt.Printf("---\n")
		} else if hashTable.employees[index].deleted {
			fmt.Printf("xxx\n")
		} else {
			fmt.Printf("%s\n", hashTable.employees[index].name)
		}

		// If this spot is empty, the value isn't in the table.
		if hashTable.employees[index] == nil {
			// If we found a deleted spot, return its index.
			if deletedIndex >= 0 {
				fmt.Printf("    Returning deleted index %d\n", deletedIndex)
				return deletedIndex
			}

			// Return this index, which holds nil.
			fmt.Printf("    Returning nil index %d\n", index)
			return index
		}

		// If this spot is deleted, remember where it is.
		if hashTable.employees[index].deleted {
			if deletedIndex < 0 {
				deletedIndex = index
			}
		} else if hashTable.employees[index].name == name {
			// If this cell holds the key, return its data.
			fmt.Printf("    Returning found index %d\n", index)
			return index
		}

		// Otherwise continue the loop.
	}

	// If we get here, then the key is not in the table and the table is full.

	// If we found a deleted spot, return it.
	if deletedIndex >= 0 {
		fmt.Printf("    Returning deleted index %d\n", deletedIndex)
		return deletedIndex
	}

	// There's nowhere to put a new entry.
	fmt.Printf("    Table is full\n")
	return -1
}

func main() {
	// Make some names.
	employees := []Employee{
		Employee{"Ann Archer", "202-555-0101", false},
		Employee{"Bob Baker", "202-555-0102", false},
		Employee{"Cindy Cant", "202-555-0103", false},
		Employee{"Dan Deever", "202-555-0104", false},
		Employee{"Edwina Eager", "202-555-0105", false},
		Employee{"Fred Franklin", "202-555-0106", false},
		Employee{"Gina Gable", "202-555-0107", false},
	}

	hashTable := NewDoubleHashTable(100)
	for _, employee := range employees {
		hashTable.set(employee.name, employee.phone)
	}
	hashTable.dump()

	hashTable.probe("Hank Hardy")
	fmt.Printf("Table contains Sally Owens: %t\n", hashTable.contains("Sally Owens"))
	fmt.Printf("Table contains Dan Deever: %t\n", hashTable.contains("Dan Deever"))
	fmt.Println("Deleting Dan Deever")
	hashTable.delete("Dan Deever")
	fmt.Printf("Table contains Dan Deever: %t\n", hashTable.contains("Dan Deever"))
	fmt.Printf("Sally Owens: %s\n", hashTable.get("Sally Owens"))
	fmt.Printf("Fred Franklin: %s\n", hashTable.get("Fred Franklin"))
	fmt.Println("Changing Fred Franklin")
	hashTable.set("Fred Franklin", "202-555-0100")
	fmt.Printf("Fred Franklin: %s\n", hashTable.get("Fred Franklin"))
	hashTable.dump()

	hashTable.probe("Ann Archer")
	hashTable.probe("Bob Baker")
	hashTable.probe("Cindy Cant")
	hashTable.probe("Dan Deever")
	hashTable.probe("Edwina Eager")
	hashTable.probe("Fred Franklin")
	hashTable.probe("Gina Gable")
	hashTable.set("Hank Hardy", "202-555-0108")
	hashTable.probe("Hank Hardy")

	// Look at clustering.
	fmt.Println(time.Now())                   // Print the time so it will compile if we use a fixed seed.
	random := rand.New(rand.NewSource(12345)) // Initialize with a fixed seed
	// random := rand.New(rand.NewSource(time.Now().UnixNano())) // Initialize with a changing seed
	bigCapacity := 1009
	bigHashTable := NewDoubleHashTable(bigCapacity)
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

func IsPrime(i int) bool { // The real input is to allow from stack.

	var t uint = 3
	var RoundSqrt uint

	Uint := uint(i)

	if Uint == 0 || Uint == 1 {
		return false
	} else if Uint == 2 || Uint == 3 {
		return true
	} else if Uint%2 == 0 {
		return false
	}

	sqrt := math.Sqrt(float64(Uint))
	RoundSqrt = uint(math.Round(sqrt))

	for t <= RoundSqrt {
		if Uint%t == 0 {
			return false
		}
		t += 2
	}
	return true
} // IsPrime
