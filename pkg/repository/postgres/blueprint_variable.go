package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"webblueprint/pkg/models"
	"webblueprint/pkg/repository"
)

type BlueprintVariableRepository struct {
	db *sql.DB
}

func NewBlueprintVariableRepository(db *sql.DB) repository.BlueprintVariableRepository {
	return &BlueprintVariableRepository{
		db: db,
	}
}

func (b *BlueprintVariableRepository) CreateVariable(ctx context.Context, bpID, bpVersionID, varID, varName, varType string, varValue interface{}) (*models.Variable, error) {
	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback()

	variableGetQuery := `SELECT variables FROM webblueprint.public.blueprint_versions WHERE id = $1`
	row := tx.QueryRowContext(ctx, variableGetQuery, bpVersionID)

	var oldVariables []models.Variable
	var buffer interface{}
	sErr := row.Scan(&buffer)
	if sErr != nil {
		return nil, fmt.Errorf("failed to scan old variables: %w", sErr)
	}

	_ = json.Unmarshal(buffer.([]byte), &oldVariables)

	variableQuery := `UPDATE webblueprint.public.blueprint_versions SET variables = $1 WHERE id = $2`
	variableInstance := &models.Variable{
		ID:           varID,
		BlueprintID:  bpID,
		Name:         varName,
		Type:         varType,
		DefaultValue: models.JSONB(map[string]interface{}{"data": varValue}),
		Description:  sql.NullString{},
		IsExposed:    true,
		Category:     sql.NullString{String: "User Defined", Valid: true},
		CreatedAt:    time.Now(),
	}

	oldVariables = append(oldVariables, *variableInstance)
	oldVariablesJSON, _ := json.Marshal(oldVariables)

	_, err = tx.ExecContext(ctx, variableQuery, oldVariablesJSON, bpVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to update blueprint version with variable: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return variableInstance, nil
}
