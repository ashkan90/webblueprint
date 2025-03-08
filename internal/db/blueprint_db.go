package db

import (
	"errors"
	"webblueprint/pkg/blueprint"
)

type BlueprintDB map[string]*blueprint.Blueprint

// Blueprint storage (in-memory for now, would be replaced with a database)
var Blueprints = make(BlueprintDB)

func (b BlueprintDB) AddBlueprint(bp *blueprint.Blueprint) {
	b[bp.ID] = bp
}

func (b BlueprintDB) GetBlueprint(name string) (*blueprint.Blueprint, error) {
	bp, ok := b[name]
	if !ok {
		return nil, errors.New("blueprint not found")
	}
	return bp, nil
}
