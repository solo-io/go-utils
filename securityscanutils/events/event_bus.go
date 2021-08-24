package events

import (
	"sync"
	"time"
)

type EventTopic int

const (
	ScanStarted EventTopic = iota
	ScanCompleted

	RepoScanStarted
	RepoScanCompleted

	VulnerabilityFound
)

type Event struct {
	Topic EventTopic
	Data interface{}
}

type EventData struct {
	Time time.Time
	Err error
}

type RepositoryEventData struct {
	*EventData
	RepositoryName string
}

type VulnerabilityFoundEventData struct {
	*EventData
	RepositoryName string
	RepositoryOwner string
	Version string
	VulnerabilityMd string
}

type EventChannel chan Event
type EventChannelSlice []EventChannel


type EventBus struct {
	subscribersByTopic map[EventTopic]EventChannelSlice
	rm                 sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribersByTopic: make(map[EventTopic]EventChannelSlice),
		rm:                 sync.RWMutex{},
	}
}

func (eb *EventBus) Subscribe(topic EventTopic, ch EventChannel)  {
	eb.rm.Lock()
	defer eb.rm.Unlock()

	if prev, found := eb.subscribersByTopic[topic]; found {
		eb.subscribersByTopic[topic] = append(prev, ch)
	} else {
		eb.subscribersByTopic[topic] = append([]EventChannel{}, ch)
	}
}

func (eb *EventBus) Publish(topic EventTopic, data interface{}) {
	eb.rm.RLock()
	defer eb.rm.RUnlock()

	if chans, found := eb.subscribersByTopic[topic]; found {
		channels := append(EventChannelSlice{}, chans...)
		go func(event Event, eventChannels EventChannelSlice) {
			for _, ch := range eventChannels {
				ch <- event
			}
		}(Event{Topic: topic, Data: data}, channels)
	}
}