package sortstructs

import (
	"fmt"
	"sort"
)

// The default Priority is the last index. The last index is not required in the prioirty list, thus any elements added
// to the struct will be added to the default priority (last).
const defaultPrioirty = -1

type PriorityValue[P comparable] interface {
	GetPriority() P
}

// PrioritySortedStruct sorts elements by prioirity. Priority lists are unordered.
// Needs a priority list and a get priority function.
// If a priority does not exist for an added element, it is added to the lowest priority.
type PrioritySortedStruct[P comparable, K PriorityValue[P]] struct {
	// prioritySets is the set lists for priorities, where P is the type used for priority.
	prioritySets map[int][]P
	// priorityMap maps the value to the priority.
	priorityMap map[P]int
	// elements are the map of elements structured by their priority.
	elements []map[int]K
	// numberOfPriorities is the number of priorities in the prioritySets
	numberOfPriorities int
	// currentElementIndex is the current element index to be inserted.
	currentElementIndex int
}

// PriorityIndex is the priority and index used to locate items.
type PriorityIndex struct {
	Priority int
	Index    int
}

func NewPrioritySortedStruct[P comparable, K PriorityValue[P]](prioritySets map[int][]P) *PrioritySortedStruct[P, K] {
	p := PrioritySortedStruct[P, K]{
		prioritySets: prioritySets,
	}
	p.Init()
	return &p
}

func (p *PrioritySortedStruct[P, K]) Init() {
	p.numberOfPriorities = len(p.prioritySets)
	// need to ensure that the prioriries are in order and there are no missing or skipped Priorities
	priorities := make([]int, 0)
	for priority, _ := range p.prioritySets {
		priorities = append(priorities, priority)
	}
	sort.Ints(priorities)
	currentP := 0
	for _, p := range priorities {
		if currentP == p {
			currentP++
		} else {
			panic(fmt.Sprintf("Priorities are not set correct, you are missing priority %d", currentP))
		}
	}
	p.elements = make([]map[int]K, p.numberOfPriorities+1)
	if p.prioritySets != nil && p.numberOfPriorities > 0 {
		for priorityIndex := range p.prioritySets {
			p.elements[priorityIndex] = make(map[int]K)
		}
		p.elements[p.numberOfPriorities] = make(map[int]K)
	} else {
		// there is only one map of elements
		p.elements[0] = make(map[int]K)
	}
	p.priorityMap = make(map[P]int)
	for index, pl := range p.prioritySets {
		for _, v := range pl {
			p.priorityMap[v] = index
		}
	}
	p.currentElementIndex = 0
}

// Get returns the element at the index, and if it exists.
func (p *PrioritySortedStruct[P, K]) Get(pi PriorityIndex) (K, bool) {
	v, ok := p.elements[pi.Priority][pi.Index]
	return v, ok
}

// Process will call the procesFunc over all the elements by priority.
func (p *PrioritySortedStruct[P, K]) Process(processFunc func(el K, pi PriorityIndex)) {
	for i := 0; i <= p.numberOfPriorities; i++ {
		m := p.elements[i]
		for index, v := range m {
			processFunc(v, PriorityIndex{Priority: i, Index: index})
		}
	}
}

// GetPriorityList returns an ordered list of the elements by priority.
func (p *PrioritySortedStruct[P, K]) GetPriorityList() []K {
	elements := make([]K, 0, p.Len())
	for priority := 0; priority <= p.numberOfPriorities; priority++ {
		mapOfElements := p.elements[priority]
		for _, el := range mapOfElements {
			elements = append(elements, el)
		}
	}
	return elements
}

// Add will add the element to the Priority Collection, returns the priority, and element number.
func (p *PrioritySortedStruct[P, K]) Add(element K) PriorityIndex {
	priority := p.getPriorityOfElement(element)
	if priority == defaultPrioirty {
		// add to the last index of the watches
		priority = p.numberOfPriorities
	}
	p.elements[priority][p.currentElementIndex] = element
	pi := PriorityIndex{Priority: priority, Index: p.currentElementIndex}
	p.currentElementIndex++
	return pi
}

// Delete will delete the element, returns true if it deleted.
func (p *PrioritySortedStruct[P, K]) Delete(pi PriorityIndex) bool {
	if p.Len() == 0 {
		return false
	}
	if _, ok := p.Get(pi); ok {
		delete(p.elements[pi.Priority], pi.Index)
		return ok
	} else {
		return false
	}
}

// Len will return the number of elements
func (p *PrioritySortedStruct[P, K]) Len() int {
	count := 0
	for _, el := range p.elements {
		count += len(el)
	}
	return count
}

// GetPriorityIndexes returns a list of all the indexes for all elements by priority.
func (p *PrioritySortedStruct[P, K]) GetPriorityIndexes() []PriorityIndex {
	pi := make([]PriorityIndex, 0, p.Len())
	for i := 0; i <= p.numberOfPriorities; i++ {
		m := p.elements[i]
		for index := range m {
			pi = append(pi, PriorityIndex{Priority: i, Index: index})
		}
	}
	return pi
}

// getPriorityOfElement returns the priority of element K.
func (p *PrioritySortedStruct[P, K]) getPriorityOfElement(element K) int {
	pv := element.GetPriority()
	if p, exists := p.priorityMap[pv]; exists {
		return p
	} else {
		// default priority is -1
		return defaultPrioirty
	}
}
