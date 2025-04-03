package types

// GetPinTypeByID returns a pin type by its ID
func GetPinTypeByID(id string) (*PinType, bool) {
	switch id {
	case "execution":
		return PinTypes.Execution, true
	case "string":
		return PinTypes.String, true
	case "number":
		return PinTypes.Number, true
	case "boolean":
		return PinTypes.Boolean, true
	case "object":
		return PinTypes.Object, true
	case "array":
		return PinTypes.Array, true
	case "any":
		return PinTypes.Any, true
	default:
		return nil, false
	}
}
