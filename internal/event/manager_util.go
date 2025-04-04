package event

import (
	"webblueprint/internal/core"
)

func ExtractEventManager(e core.EventManagerInterface) *EventManager {
	m, ok := e.(*eventManagerAdapter)
	if !ok {
		return nil
	}

	return m.manager
}
