package events

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FilterEvents returns the events with the given types
func FilterEvents(events sdk.StringEvents, eventTypes string) sdk.StringEvents {
	var filteredEvents sdk.StringEvents
	for _, event := range events {
		if event.Type == eventTypes {
			filteredEvents = append(filteredEvents, event)
		}
	}
	return filteredEvents
}

// FindEventByType returns the event with the given type
func FindEventByType(events sdk.StringEvents, eventType string) (sdk.StringEvent, bool) {
	for _, event := range events {
		if event.Type == eventType {
			return event, true
		}
	}
	return sdk.StringEvent{}, false
}

// FindAttributeByKey returns the attribute with the given key
func FindAttributeByKey(event sdk.StringEvent, key string) (sdk.Attribute, bool) {
	for _, attribute := range event.Attributes {
		if attribute.Key == key {
			return attribute, true
		}
	}
	return sdk.Attribute{}, false
}
