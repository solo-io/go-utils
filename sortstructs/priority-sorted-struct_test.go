package sortstructs_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/go-utils/sortstructs"
)

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func getCount(arr []int, value int) int {
	count := 0
	for _, v := range arr {
		if v == value {
			count++
		}
	}
	return count
}

func matchInts(val int) int {
	return val
}

func createNumberPriorityList(pl [][]int) (*PrioritySortedStruct[int, int], []PriorityIndex) {
	listOfInserts := make([]PriorityIndex, 0)

	p := NewPrioritySortedStruct(pl, matchInts)
	numbersToNotMatch := make([]int, 0)
	for _, p := range pl {
		numbersToNotMatch = append(numbersToNotMatch, p...)
	}
	for i := 0; i <= 100; i++ {
		if contains(numbersToNotMatch, i) {
			continue
		}
		listOfInserts = append(listOfInserts, p.Add(i))
	}
	return p, listOfInserts
}

func addValue[P comparable, K any](p *PrioritySortedStruct[P, K], values []K, numberToAdd int) {
	for _, v := range values {
		for i := 0; i < numberToAdd; i++ {
			p.Add(v)
		}
	}
}

type MyStruct struct {
	value string
}

var _ = Describe("sort structs", func() {

	BeforeEach(func() {
	})
	const numberInserted = 10

	Describe("GetPriorityList", func() {
		It("it return a priority list and the first values are the most prioritized", func() {
			numberToMatch := 5
			priorityList := [][]int{{numberToMatch}}
			p, _ := createNumberPriorityList(priorityList)
			addValue(p, []int{numberToMatch}, numberInserted)

			pl := p.GetPriorityList()
			numberCountedInPriority := 0
			for _, el := range pl {
				if el == numberToMatch {
					numberCountedInPriority++
				} else {
					break
				}
			}
			Expect(numberCountedInPriority).To(Equal(numberInserted))
			Expect(p.Len()).To(Equal(100 + numberInserted))
		})
		It("returns a multi layer priority list", func() {
			numbersToMatch := []int{5, 16, 7}
			priorityList := [][]int{{numbersToMatch[0]}, {numbersToMatch[1]}, {numbersToMatch[2]}}
			p, _ := createNumberPriorityList(priorityList)
			addValue(p, numbersToMatch, numberInserted)

			pl := p.GetPriorityList()
			numberCountedInPriorityMap := make(map[int]int)
			plIndex := 0
			currentCount := 0
			var numbersAndCount [3][2]int
			lastIndexNumber := numbersToMatch[len(numbersToMatch)-1]
			for priority, numToMatch := range numbersToMatch {
				for {
					currentEl := pl[plIndex]
					if currentEl == numToMatch {
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
			numbersToMatch := []int{5, 16, 7}
			priorityList := [][]int{{numbersToMatch[0]}, {numbersToMatch[1]}, {numbersToMatch[2]}}
			p, indexes := createNumberPriorityList(priorityList)
			addValue(p, numbersToMatch, numberInserted)

			// -2 because the createNumberPriorityList inserts 101 - numberInPriorityList
			Expect(p.Len()).To(Equal(100 + (3 * numberInserted) - 2))

			v, exists := p.Get(indexes[0]) // gets 0
			Expect(exists).To(Equal(true))
			Expect(v).To(Equal(0))

			p.Delete(indexes[0]) // delete 0
			p.Delete(indexes[2]) // delete 2

			_, exists = p.Get(indexes[0]) // gets 0
			Expect(exists).To(Equal(false))

			Expect(p.Len()).To(Equal(100 + (3 * numberInserted) - 4))
		})
		It("Should always have a zero length when deleting all elements", func() {
			numbersToMatch := []int{5, 16, 7}
			priorityList := [][]int{{numbersToMatch[0]}, {numbersToMatch[1]}, {numbersToMatch[2]}}
			p := NewPrioritySortedStruct(priorityList, matchInts)
			Expect(p.Len()).To(Equal(0))
			p.Delete(PriorityIndex{0, 0})
			Expect(p.Len()).To(Equal(0))

			valueToAdd := 1
			pi := p.Add(valueToAdd)
			Expect(p.Len()).To(Equal(1))
			v, _ := p.Get(pi)
			Expect(v).To(Equal(valueToAdd))
			p.Delete(pi)
			Expect(p.Len()).To(Equal(0))
			v, exists := p.Get(pi)
			Expect(v).To(Equal(0))
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
			Expect(v).To(Equal(0))
			Expect(exists).To(Equal(false))
		})
	})
	Describe("GetPriorityIndexes", func() {
		It("should return the priority index list by priority", func() {
			numbersToMatch := []int{5, 16, 7}
			priorityList := [][]int{{numbersToMatch[0]}, {numbersToMatch[1]}, {numbersToMatch[2]}}
			p := NewPrioritySortedStruct(priorityList, matchInts)
			for i := 0; i <= 100; i++ {
				p.Add(i)
			}
			pis := p.GetPriorityIndexes()
			Expect(len(pis)).To(Equal(p.Len()))
			v, exists := p.Get(pis[0])
			Expect(v).To(Equal(numbersToMatch[0]))
			Expect(exists).To(Equal(true))
			v, exists = p.Get(pis[1])
			Expect(v).To(Equal(numbersToMatch[1]))
			Expect(exists).To(Equal(true))
			v, exists = p.Get(pis[2])
			Expect(v).To(Equal(numbersToMatch[2]))
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
			numbersToMatch := []int{5, 16, 7}
			priorityList := [][]int{{numbersToMatch[0]}, {numbersToMatch[1]}, {numbersToMatch[2]}}
			p := NewPrioritySortedStruct(priorityList, matchInts)
			indexesOfValues := make(map[int]PriorityIndex)
			for i := 0; i <= 100; i++ {
				pi := p.Add(i)
				indexesOfValues[i] = pi
			}
			firstThree := 0
			elementsReceived := make([]int, 0)
			processFunc := func(value int, pi PriorityIndex) {
				if firstThree < 3 {
					Expect(numbersToMatch[firstThree]).To(Equal(value))
					firstThree++
				}
				elementsReceived = append(elementsReceived, value)
				piValue, piExists := p.Get(pi)
				actualIndexValue, indexValueExists := p.Get(indexesOfValues[value])
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
			p := NewPrioritySortedStruct(nil, matchInts)
			pi := p.Add(1)
			v, exists := p.Get(pi)
			Expect(exists).To(Equal(true))
			Expect(v).To(Equal(1))
			d := p.Delete(pi)
			Expect(d).To(Equal(true))
		})
		It("Should work with structs", func() {
			p := NewPrioritySortedStruct([][]string{{"first", "fourth"}, {"second"}}, func(el MyStruct) string { return el.value })
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
			numbersToMatch := []int{5, 16, 7, 19, 25}
			priorityList := [][]int{{numbersToMatch[0], numbersToMatch[3]}, {numbersToMatch[1]}, {numbersToMatch[2], numbersToMatch[4]}}
			p := NewPrioritySortedStruct(priorityList, matchInts)
			indexesOfValues := make(map[int]PriorityIndex)
			for i := 0; i <= 100; i++ {
				pi := p.Add(i)
				indexesOfValues[i] = pi
			}
			pl := p.GetPriorityList()
			l := pl[0:2]
			Expect(contains(l, numbersToMatch[0])).To(Equal(true))
			Expect(contains(l, numbersToMatch[3])).To(Equal(true))
			l = pl[2:3]
			Expect(contains(l, numbersToMatch[1])).To(Equal(true))
			l = pl[3:5]
			Expect(contains(l, numbersToMatch[2])).To(Equal(true))
			Expect(contains(l, numbersToMatch[4])).To(Equal(true))
		})
		It("Should have the correct number given multiple priorities", func() {
			numbersToMatch := []int{5, 16, 7, 19, 25}
			priorityList := [][]int{{numbersToMatch[0], numbersToMatch[3]}, {numbersToMatch[1]}, {numbersToMatch[2], numbersToMatch[4]}}
			p := NewPrioritySortedStruct(priorityList, matchInts)
			indexesOfValues := make(map[int]PriorityIndex)
			for i := 0; i <= 100; i++ {
				pi := p.Add(i)
				indexesOfValues[i] = pi
			}
			addValue(p, []int{numbersToMatch[0]}, 15)
			addValue(p, []int{numbersToMatch[1]}, 19)
			addValue(p, []int{numbersToMatch[2]}, 27)
			addValue(p, []int{numbersToMatch[3]}, 6)
			addValue(p, []int{numbersToMatch[4]}, 3)
			pl := p.GetPriorityList()
			l := pl[0:23] // 15 + 6 = 21 + 2
			Expect(contains(l, numbersToMatch[0])).To(Equal(true))
			Expect(contains(l, numbersToMatch[3])).To(Equal(true))
			Expect(getCount(l, numbersToMatch[0])).To(Equal(16))
			Expect(getCount(l, numbersToMatch[3])).To(Equal(7))
			l = pl[23:43] // 20
			Expect(contains(l, numbersToMatch[1])).To(Equal(true))
			Expect(getCount(l, numbersToMatch[1])).To(Equal(20))
			l = pl[43:75] // 28 + 4 = 32
			Expect(contains(l, numbersToMatch[2])).To(Equal(true))
			Expect(contains(l, numbersToMatch[4])).To(Equal(true))
			Expect(getCount(l, numbersToMatch[2])).To(Equal(28))
			Expect(getCount(l, numbersToMatch[4])).To(Equal(4))
		})
	})
})
