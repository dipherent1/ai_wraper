package utils

import "github.com/mark3labs/flyt"

// This struct is now shared across the application.
type Conversation struct {
	User string
	AI   any
}

type History struct {
	Conversations []Conversation
}

func GetHistory(shared *flyt.SharedStore) History {
	raw, _ := shared.Get("history")
	switch v := raw.(type) {
	case History:
		return v
	case []Conversation:
		return History{Conversations: v}
	case nil:
		return History{}
	default:
		// Best-effort conversion from []interface{} with map[string]interface{}
		if s, ok := raw.([]interface{}); ok {
			convs := make([]Conversation, 0, len(s))
			for _, it := range s {
				if m, ok := it.(map[string]interface{}); ok {
					var c Conversation
					if user, ok := m["User"].(string); ok {
						c.User = user
					}
					if ai, ok := m["AI"]; ok {
						c.AI = ai
					}
					convs = append(convs, c)
				}
			}
			return History{Conversations: convs}
		}
		return History{}
	}
}
