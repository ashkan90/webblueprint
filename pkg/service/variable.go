package service

import (
	"context"
	"fmt"
	"webblueprint/pkg/models"
	"webblueprint/pkg/repository"
)

type BlueprintVariableService struct {
	blueprintRepo         repository.BlueprintRepository
	blueprintVariableRepo repository.BlueprintVariableRepository
}

func NewBlueprintVariableService(blueprintRepo repository.BlueprintRepository, blueprintVariableRepo repository.BlueprintVariableRepository) *BlueprintVariableService {
	return &BlueprintVariableService{
		blueprintRepo:         blueprintRepo,
		blueprintVariableRepo: blueprintVariableRepo,
	}
}

func (s *BlueprintVariableService) CreateVariable(ctx context.Context, version int, blueprintID, varID, varName, varType string, varValue interface{}) (*models.Variable, error) {
	versionModel, err := s.blueprintRepo.GetVersion(ctx, blueprintID, version)
	if err != nil {
		return nil, fmt.Errorf("an error ocurred while getting version: %w", err)
	}

	return s.blueprintVariableRepo.CreateVariable(ctx, blueprintID, versionModel.ID, varID, varName, varType, varValue)
}
