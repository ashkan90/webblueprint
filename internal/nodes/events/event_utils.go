package events

import (
	"webblueprint/internal/core"
	"webblueprint/internal/event"
)

// convertEventHandlerContext converts an event.EventHandlerContext to a core.EventHandlerContext
func convertEventHandlerContext(ctx event.EventHandlerContext) *core.EventHandlerContext {
	return &core.EventHandlerContext{
		EventID:    ctx.EventID,
		Parameters: ctx.Parameters,
		SourceID:   ctx.SourceID,
		Timestamp:  ctx.Timestamp,
	}
}

// NOTE: Removed convertEventManager and eventManagerAdapter as they are likely
// redundant with the adapter provided by event.EventManager itself and were
// causing compilation errors due to interface mismatches after refactoring.
// Code should now use eventManager.AsEventManagerInterface() where needed.
