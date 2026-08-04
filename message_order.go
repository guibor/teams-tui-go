package main

import (
	"sort"
	"strings"
	"time"
)

func messageCreatedTime(message Message) (time.Time, bool) {
	value := strings.TrimSpace(message.CreatedDateTime)
	if value == "" {
		return time.Time{}, false
	}
	created, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, false
	}
	return created, true
}

// messageChronologyLess reports whether left belongs before right in the
// requested ordering. Valid Graph timestamps always precede malformed values.
func messageChronologyLess(left, right Message, newestFirst bool) bool {
	leftTime, leftOK := messageCreatedTime(left)
	rightTime, rightOK := messageCreatedTime(right)
	if leftOK != rightOK {
		return leftOK
	}
	if leftOK && !leftTime.Equal(rightTime) {
		if newestFirst {
			return leftTime.After(rightTime)
		}
		return leftTime.Before(rightTime)
	}
	if !leftOK && left.CreatedDateTime != right.CreatedDateTime {
		if newestFirst {
			return left.CreatedDateTime > right.CreatedDateTime
		}
		return left.CreatedDateTime < right.CreatedDateTime
	}
	if newestFirst {
		return left.ID > right.ID
	}
	return left.ID < right.ID
}

func messageCreatedAfter(left, right Message) bool {
	return messageChronologyLess(left, right, true)
}

func sortMessagesNewestFirst(messages []Message) {
	sort.SliceStable(messages, func(i, j int) bool {
		return messageChronologyLess(messages[i], messages[j], true)
	})
}

func sortMessagesOldestFirst(messages []Message) {
	sort.SliceStable(messages, func(i, j int) bool {
		return messageChronologyLess(messages[i], messages[j], false)
	})
}
