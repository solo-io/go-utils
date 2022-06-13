package sortstructs_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/sortstructs"
	"golang.org/x/exp/slices"
)

type PriorityInt int

func (p PriorityInt) GetPriority() int {
	return int(p)
}

func getCount(arr []PriorityInt, value int) int {
	count := 0
	for _, v := range arr {
		if int(v) == value {
			count++
		}
	}
	return count
}

func convert(nums []int) []PriorityInt {
	converted := make([]PriorityInt, 0, len(nums))
	for _, v := range nums {
		converted = append(converted, PriorityInt(v))
	}
	return converted
}

func createNumberPriorityList(pl map[int][]int) (*PrioritySortedStruct[int, PriorityInt], []PriorityIndex) {
	listOfInserts := make([]PriorityIndex, 0)

	p := NewPrioritySortedStruct[int, PriorityInt](pl)
	numbersToNotMatch := make([]int, 0)
	for _, p := range pl {
		numbersToNotMatch = append(numbersToNotMatch, p...)
	}
	for i := 0; i <= 100; i++ {
		if slices.Contains(numbersToNotMatch, i) {
			continue
		}
		listOfInserts = append(listOfInserts, p.Add(PriorityInt(i)))
	}
	return p, listOfInserts
}

func addValue[P comparable, K PriorityValue[P]](p *PrioritySortedStruct[P, K], values []K, numberToAdd int) {
	for _, v := range values {
		for i := 0; i < numberToAdd; i++ {
			p.Add(v)
		}
	}
}

type MyStruct struct {
	value string
}

func (m MyStruct) GetPriority() string {
	return m.value
}

var _ = Describe("sort structs", func() {

	BeforeEach(func() {
	})
	const numberInserted = 10

	var numbersToMatch = []int{5, 16, 7}
	var priorityList = map[int][]int{
		0: {numbersToMatch[0]},
		1: {numbersToMatch[1]},
		2: {numbersToMatch[2]},
	}

	var numbersToMatch2 = []int{5, 16, 7, 19, 25}
	var priorityList2 = map[int][]int{
		0: {numbersToMatch2[0], numbersToMatch2[3]},
		1: {numbersToMatch2[1]},
		2: {numbersToMatch2[2], numbersToMatch2[4]},
	}

	Describe("GetPriorityList", func() {
		It("it return a priority list and the first values are the most prioritized", func() {
			numberToMatch := 5
			priorityList := map[int][]int{0: {numberToMatch}}
			p, _ := createNumberPriorityList(priorityList)
			addValue(p, []PriorityInt{PriorityInt(numberToMatch)}, numberInserted)

			pl := p.GetPriorityList()
			numberCountedInPriority := 0
			for _, el := range pl {
				if int(el) == numberToMatch {
					numberCountedInPriority++
				} else {
					break
				}
			}
			Expect(numberCountedInPriority).To(Equal(numberInserted))
			Expect(p.Len()).To(Equal(100 + numberInserted))
		})
		It("returns a multi layer priority list", func() {
			p, _ := createNumberPriorityList(priorityList)
			converted := convert(numbersToMatch)
			addValue(p, converted, numberInserted)

			pl := p.GetPriorityList()
			numberCountedInPriorityMap := make(map[int]int)
			plIndex := 0
			currentCount := 0
			var numbersAndCount [3][2]int
			lastIndexNumber := numbersToMatch[len(numbersToMatch)-1]
			for priority, numToMatch := range numbersToMatch {
				for {
					currentEl := pl[plIndex]
					if int(currentEl) == numToMatch {
						numberCountedInPriorityMap[priority]++
					}
					plIndex++
					currentCount++
					if currentCount == numberInserted {
						numbersAndCount[priority] = [2]int{numToMatch, currentCount}
						lastIndexNumber = numToMatch
						currentCount = 0
						break
					}
				}
			}
			Expect(pl[plIndex]).ToNot(Equal(lastIndexNumber))
			for index, numberAndCount := range numbersAndCount {
				Expect(numberAndCount[0]).To(Equal(numbersToMatch[index]))
				Expect(numberAndCount[1]).To(Equal(numberInserted))
			}
			// -2 because the createNumberPriorityList inserts 101 - numberInPriorityList
			Expect(p.Len()).To(Equal(100 + (3 * numberInserted) - 2))
		})
	})
	Describe("Add, Delete, Len, Get", func() {
		It("Should return the correct length if elements are deleted", func() {
			p, indexes := createNumberPriorityList(priorityList)
			converted := convert(numbersToMatch)
			addValue(p, converted, numberInserted)

			// -2 because the createNumberPriorityList inserts 101 - numberInPriorityList
			Expect(p.Len()).To(Equal(100 + (3 * numberInserted) - 2))

			v, exists := p.Get(indexes[0]) // gets 0
			Expect(exists).To(Equal(true))
			Expect(int(v)).To(Equal(0))

			p.Delete(indexes[0]) // delete 0
			p.Delete(indexes[2]) // delete 2

			_, exists = p.Get(indexes[0]) // gets 0
			Expect(exists).To(Equal(false))

			Expect(p.Len()).To(Equal(100 + (3 * numberInserted) - 4))
		})
		It("Should always have a zero length when deleting all elements", func() {
			p := NewPrioritySortedStruct[int, PriorityInt](priorityList)
			Expect(p.Len()).To(Equal(0))
			p.Delete(PriorityIndex{0, 0})
			Expect(p.Len()).To(Equal(0))

			valueToAdd := PriorityInt(1)
			pi := p.Add(valueToAdd)
			Expect(p.Len()).To(Equal(1))
			v, _ := p.Get(pi)
			Expect(v).To(Equal(valueToAdd))
			p.Delete(pi)
			Expect(p.Len()).To(Equal(0))
			v, exists := p.Get(pi)
			Expect(int(v)).To(Equal(0))
			Expect(exists).To(Equal(false))

			p.Delete(PriorityIndex{0, 0})
			Expect(p.Len()).To(Equal(0))
			p.Delete(PriorityIndex{0, 0})
			Expect(p.Len()).To(Equal(0))

			valueToAdd = 2
			pi = p.Add(valueToAdd)
			Expect(p.Len()).To(Equal(1))
			v, _ = p.Get(pi)
			Expect(v).To(Equal(valueToAdd))
			p.Delete(pi)
			Expect(p.Len()).To(Equal(0))
			v, exists = p.Get(pi)
			Expect(int(v)).To(Equal(0))
			Expect(exists).To(Equal(false))
		})
	})
	Describe("GetPriorityIndexes", func() {
		It("should return the priority index list by priority", func() {
			p := NewPrioritySortedStruct[int, PriorityInt](priorityList)
			for i := 0; i <= 100; i++ {
				p.Add(PriorityInt(i))
			}
			pis := p.GetPriorityIndexes()
			Expect(len(pis)).To(Equal(p.Len()))
			v, exists := p.Get(pis[0])
			Expect(int(v)).To(Equal(numbersToMatch[0]))
			Expect(exists).To(Equal(true))
			v, exists = p.Get(pis[1])
			Expect(int(v)).To(Equal(numbersToMatch[1]))
			Expect(exists).To(Equal(true))
			v, exists = p.Get(pis[2])
			Expect(int(v)).To(Equal(numbersToMatch[2]))
			Expect(exists).To(Equal(true))

			for i := 3; i <= 100; i++ {
				_, exists = p.Get(pis[i])
				Expect(exists).To(Equal(true))
			}
			_, exists = p.Get(PriorityIndex{3, 101})
			Expect(exists).To(Equal(false))
		})
	})
	Describe("Process", func() {
		It("Should be able to iterate over all the values and process them", func() {
			p := NewPrioritySortedStruct[int, PriorityInt](priorityList)
			indexesOfValues := make(map[int]PriorityIndex)
			for i := 0; i <= 100; i++ {
				pi := p.Add(PriorityInt(i))
				indexesOfValues[i] = pi
			}
			firstThree := 0
			elementsReceived := make([]int, 0)
			processFunc := func(value PriorityInt, pi PriorityIndex) {
				if firstThree < 3 {
					Expect(numbersToMatch[firstThree]).To(Equal(int(value)))
					firstThree++
				}
				elementsReceived = append(elementsReceived, int(value))
				piValue, piExists := p.Get(pi)
				actualIndexValue, indexValueExists := p.Get(indexesOfValues[int(value)])
				Expect(piExists).To(Equal(true))
				Expect(indexValueExists).To(Equal(true))
				Expect(piValue).To(Equal(actualIndexValue))
				deleted := p.Delete(pi)
				Expect(deleted).To(Equal(true))
			}
			p.Process(processFunc)
			n := func(value, index int) {
				v := elementsReceived[index]
				Expect(v).To(Equal(value))
			}
			n(numbersToMatch[0], 0)
			n(numbersToMatch[1], 1)
			n(numbersToMatch[2], 2)
			Expect(p.Len()).To(Equal(0))
		})
	})
	Describe("One Offs", func() {
		It("Should work with no priorities", func() {
			p := NewPrioritySortedStruct[int, PriorityInt](nil)
			pi := p.Add(1)
			v, exists := p.Get(pi)
			Expect(exists).To(Equal(true))
			Expect(int(v)).To(Equal(1))
			d := p.Delete(pi)
			Expect(d).To(Equal(true))
		})
		It("Should work with structs", func() {
			prioirtySet := map[int][]string{
				0: {"first", "fourth"},
				1: {"second"},
			}
			p := NewPrioritySortedStruct[string, MyStruct](prioirtySet)
			m := MyStruct{value: "first"}
			pi := p.Add(m)
			Expect(pi.Priority).To(Equal(0))
			m1 := MyStruct{value: "second"}
			pi = p.Add(m1)
			Expect(pi.Priority).To(Equal(1))
			m2 := MyStruct{value: "third"}
			pi = p.Add(m2)
			Expect(pi.Priority).To(Equal(2))
			m3 := MyStruct{value: "fourth"}
			pi = p.Add(m3)
			Expect(pi.Priority).To(Equal(0))
		})
		It("Should be able to have priorities of 2 or more", func() {
			converted := convert(numbersToMatch2)
			p := NewPrioritySortedStruct[int, PriorityInt](priorityList2)
			indexesOfValues := make(map[int]PriorityIndex)
			for i := 0; i <= 100; i++ {
				pi := p.Add(PriorityInt(i))
				indexesOfValues[i] = pi
			}
			pl := p.GetPriorityList()
			l := pl[0:2]
			Expect(slices.Contains(l, converted[0])).To(Equal(true))
			Expect(slices.Contains(l, converted[3])).To(Equal(true))
			l = pl[2:3]
			Expect(slices.Contains(l, converted[1])).To(Equal(true))
			l = pl[3:5]
			Expect(slices.Contains(l, converted[2])).To(Equal(true))
			Expect(slices.Contains(l, converted[4])).To(Equal(true))
		})
		It("Should have the correct number given multiple priorities", func() {
			converted := convert(numbersToMatch2)
			p := NewPrioritySortedStruct[int, PriorityInt](priorityList2)
			indexesOfValues := make(map[int]PriorityIndex)
			for i := 0; i <= 100; i++ {
				pi := p.Add(PriorityInt(i))
				indexesOfValues[i] = pi
			}
			addValue(p, []PriorityInt{converted[0]}, 15)
			addValue(p, []PriorityInt{converted[1]}, 19)
			addValue(p, []PriorityInt{converted[2]}, 27)
			addValue(p, []PriorityInt{converted[3]}, 6)
			addValue(p, []PriorityInt{converted[4]}, 3)
			pl := p.GetPriorityList()
			l := pl[0:23] // 15 + 6 = 21 + 2
			Expect(slices.Contains(l, converted[0])).To(Equal(true))
			Expect(slices.Contains(l, converted[3])).To(Equal(true))
			Expect(getCount(l, numbersToMatch2[0])).To(Equal(16))
			Expect(getCount(l, numbersToMatch2[3])).To(Equal(7))
			l = pl[23:43] // 20
			Expect(slices.Contains(l, converted[1])).To(Equal(true))
			Expect(getCount(l, numbersToMatch2[1])).To(Equal(20))
			l = pl[43:75] // 28 + 4 = 32
			Expect(slices.Contains(l, converted[2])).To(Equal(true))
			Expect(slices.Contains(l, converted[4])).To(Equal(true))
			Expect(getCount(l, numbersToMatch2[2])).To(Equal(28))
			Expect(getCount(l, numbersToMatch2[4])).To(Equal(4))
		})
		It("Should not have the same index for a deleted element", func() {
			p := NewPrioritySortedStruct[int, PriorityInt](map[int][]int{})
			p.Add(0)
			piToDelete := p.Add(1)
			deleted := p.Delete(piToDelete)
			Expect(deleted).To(Equal(deleted))
			piToKeep := p.Add(2)
			Expect(piToKeep.Index).To(Equal(uint64(2)))
			Expect(piToDelete.Index).To(Equal(uint64(1)))
		})
	})
})
