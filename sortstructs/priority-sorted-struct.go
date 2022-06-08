package sortstructs

const defaultPrioirty = -1

// PrioritySortedStruct sorts elements by prioirity. Priority lists are unordered.
// Needs a priority list and a priority match function.
// If a priority does not exist for an added element, it is added to the lowest priority.
type PrioritySortedStruct[P comparable, K any] struct {
	// priorityList is the priority of all the elements, where P is the type
	priorityList [][]P
	// getPriorityValue returns the value of the priority from the element
	getPriorityValue func(el K) P
	// priorityMap maps the value to the priority
	priorityMap map[P]int
	// count is the number of elements in the struct
	count int
	// elements are the map of elements structured by their priority.
	elements []map[int]K
	// numberOfPriorities is the number of priorities in the priorityList + 1
	numberOfPriorities int
	// currentElementNumber is the current element number
	currentElementNumber int
}

// PriorityIndex is the priority and index used to locate items
type PriorityIndex struct {
	Priority int
	Index    int
}

func NewPrioritySortedStruct[P comparable, K any](priorityList [][]P, priorityMatchFunction func(el K) P) *PrioritySortedStruct[P, K] {
	p := PrioritySortedStruct[P, K]{
		priorityList:     priorityList,
		getPriorityValue: priorityMatchFunction,
	}
	p.Init()
	return &p
}

func (p *PrioritySortedStruct[P, K]) Init() {
	p.numberOfPriorities = len(p.priorityList)
	p.elements = make([]map[int]K, p.numberOfPriorities+1)
	if p.priorityList != nil {
		for priorityIndex := range p.priorityList {
			p.elements[priorityIndex] = make(map[int]K)
		}
		p.elements[p.numberOfPriorities] = make(map[int]K)
	} else {
		// there is only 1 map of elements
		p.elements[0] = make(map[int]K)
	}
	p.priorityMap = make(map[P]int)
	for index, pl := range p.priorityList {
		for _, v := range pl {
			p.priorityMap[v] = index
		}
	}
	p.count = 0
	p.currentElementNumber = 0
}

// Get returns the element at the index, and if it exists
func (p *PrioritySortedStruct[P, K]) Get(pi PriorityIndex) (K, bool) {
	v, ok := p.elements[pi.Priority][pi.Index]
	return v, ok
}

// Process will call the procesFunc over all the elements by priority
func (p *PrioritySortedStruct[P, K]) Process(processFunc func(el K, pi PriorityIndex)) {
	for i := 0; i <= p.numberOfPriorities; i++ {
		m := p.elements[i]
		for index, v := range m {
			processFunc(v, PriorityIndex{Priority: i, Index: index})
		}
	}
}

// GetPriorityList returns an ordered list of the elements by priority
func (p *PrioritySortedStruct[P, K]) GetPriorityList() []K {
	elements := make([]K, p.count)
	currentIndex := 0
	for priority := 0; priority <= p.numberOfPriorities; priority++ {
		mapOfElements := p.elements[priority]
		for _, el := range mapOfElements {
			elements[currentIndex] = el
			currentIndex++
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
	p.elements[priority][p.currentElementNumber] = element
	pi := PriorityIndex{Priority: priority, Index: p.currentElementNumber}
	p.count++
	p.currentElementNumber++
	return pi
}

// Delete will delete the element, return if it deleted
func (p *PrioritySortedStruct[P, K]) Delete(pi PriorityIndex) bool {
	if p.count == 0 {
		return false
	}
	if _, ok := p.Get(pi); ok {
		delete(p.elements[pi.Priority], pi.Index)
		p.count--
		return ok
	} else {
		return false
	}
}

// Len will return the number of elements
func (p *PrioritySortedStruct[P, K]) Len() int {
	return p.count
}

// GetPriorityIndexes returns a list of all the indexes for all elements by priority
func (p *PrioritySortedStruct[P, K]) GetPriorityIndexes() []PriorityIndex {
	pi := make([]PriorityIndex, p.Len())
	currentIndex := 0
	for i := 0; i <= p.numberOfPriorities; i++ {
		m := p.elements[i]
		for index := range m {
			pi[currentIndex] = PriorityIndex{Priority: i, Index: index}
			currentIndex++
		}
	}
	return pi
}

// getPriorityOfElement returns the priority of element K what ever that is
func (p *PrioritySortedStruct[P, K]) getPriorityOfElement(element K) int {
	pv := p.getPriorityValue(element)
	if p, exists := p.priorityMap[pv]; exists {
		return p
	} else {
		// default priority is -1
		return defaultPrioirty
	}
}
