package nodes

import (
	"webblueprint/internal/node"
	"webblueprint/internal/nodes/data"
	"webblueprint/internal/nodes/logic"
	"webblueprint/internal/nodes/math"
	"webblueprint/internal/nodes/utility"
	"webblueprint/internal/nodes/web"
)

var (
	Core = map[string]node.NodeFactory{
		// Mantık düğümleri
		"if-condition": logic.NewIfConditionNode,
		"loop":         logic.NewLoopNode,
		"sequence":     logic.NewSequenceNode,
		"branch":       logic.NewBranchNode,

		// Web düğümleri
		"http-request": web.NewHTTPRequestNode,
		"dom-element":  web.NewDOMElementNode,
		"dom-event":    web.NewDOMEventNode,
		"storage":      web.NewStorageNode,

		// Veri düğümleri
		"constant-string":   data.NewStringConstantNode,
		"constant-number":   data.NewNumberConstantNode,
		"constant-boolean":  data.NewBooleanConstantNode,
		"variable-get":      data.NewVariableGetNode,
		"variable-set":      data.NewVariableSetNode,
		"json-processor":    data.NewJSONNode,
		"array-operations":  data.NewArrayNode,
		"object-operations": data.NewObjectNode,
		"type-conversion":   data.NewTypeConversionNode,

		// Matematik düğümleri
		"math-add":      math.NewAddNode,
		"math-subtract": math.NewSubtractNode,
		"math-multiply": math.NewMultiplyNode,
		"math-divide":   math.NewDivideNode,

		// Yardımcı düğümler
		"print": utility.NewPrintNode,
		"timer": utility.NewTimerNode,
	}
)
