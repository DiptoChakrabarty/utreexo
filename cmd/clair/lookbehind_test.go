package main

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

func TestSliceDelete(t *testing.T) {

	cacheLen := int32(1000)

	cacheSlice1 := make([]int, cacheLen)
	for i, _ := range cacheSlice1 {
		cacheSlice1[i] = int(rand.Int31())
	}
	// fmt.Printf("orig %v\n", cacheSlice1)

	cacheSlice2 := make([]int, cacheLen)
	copy(cacheSlice2, cacheSlice1)
	originalCache := make([]int, cacheLen)
	copy(originalCache, cacheSlice1)
	deletionMap := make(map[int]bool)
	for i := 0; i < 5; i++ {
		del := int(rand.Int31n(cacheLen))
		deletionMap[del] = true
	}
	deletions := make([]int, 0)
	for key, _ := range deletionMap {
		//fmt.Println("key:",key)
		deletions = append(deletions, key)
		//fmt.Println(deletions)
	}
	sort.Ints(deletions)
	fmt.Println(deletions)

	for z, deletePosition := range deletions {
		//fmt.Println(deletePosition)
		cacheSlice1 = append(cacheSlice1[:deletePosition-z], cacheSlice1[deletePosition-z+1:]...)
	}

	//newCache2 := make([]int, len(cacheSlice2)-len(deletions))
	prevPos := 0
	for z, deletePosition := range deletions {
		/*fmt.Println("prevPos: ",prevPos)
		fmt.Println("z: ",z)
		fmt.Println("deletePos: ",deletePosition)
		fmt.Println("length of cache: ",len(cacheSlice2))*/
		//copy(cacheSlice2[prevPos-1:], cacheSlice2[prevPos:deletePosition-z])
		if prevPos == 0 {
			copy(cacheSlice2[0:], cacheSlice2[0:deletePosition])
		} else {
			copy(cacheSlice2[prevPos-z:], cacheSlice2[prevPos:deletePosition])
		}
		prevPos = deletePosition + 1
	}
	copy(cacheSlice2[prevPos-len(deletions):], cacheSlice2[prevPos:])
	//fmt.Println(len(cacheSlice1))
	//fmt.Println(len(cacheSlice2))
	cacheSlice2 = cacheSlice2[:int(cacheLen)-len(deletions)]
	// fmt.Printf("cs1 %v\ncs2 %v\n", cacheSlice1, cacheSlice2)
	if len(cacheSlice1) != len(cacheSlice2) {
		t.Fatalf("CS1 len %d != CS2 len %d", len(cacheSlice1), len(cacheSlice2))
	}

	for ind, _ := range cacheSlice1 {
		if cacheSlice1[ind] != cacheSlice2[ind] {
			/*fmt.Println(originalCache[ind:ind + 5])
			fmt.Println(cacheSlice1[ind:ind + 5])
			fmt.Println(cacheSlice2[ind:ind + 5])*/
			t.Fatal("Doesn't match for index ", ind, " ; is actually ", cacheSlice1[ind], " but got ", cacheSlice2[ind])

		}

	}
	fmt.Println("ALL GOOD!")

}
