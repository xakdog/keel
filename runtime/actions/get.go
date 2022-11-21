package actions

import (
	"errors"
	"fmt"

	"github.com/teamkeel/keel/proto"
)

func Get(scope *Scope, input map[string]any) (map[string]any, error) {
	query := NewQuery(scope.model)

	err := query.applyImplicitFilters(scope, input)
	if err != nil {
		return nil, err
	}

	err = query.applyExplicitFilters(scope, input)
	if err != nil {
		return nil, err
	}

	isAuthorised, err := query.isAuthorised(scope, input)
	if err != nil {
		return nil, err
	}

	if !isAuthorised {
		return nil, errors.New("not authorized to access this operation")
	}

	if scope.operation.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM {
		return ParseGetObjectResponse(scope.context, scope.operation, input)
	}

	// Select all columns and distinct on id
	query.AppendSelect(Field("*"))
	query.AppendDistinctOn(Field("id"))

	// Execute database request with results
	results, affected, err := query.SelectStatement().ExecuteWithResults(scope.context)
	if err != nil {
		return nil, err
	}

	if affected == 0 {
		return nil, errors.New("no records found for Get() operation")
	} else if affected > 1 {
		return nil, fmt.Errorf("Get() operation should find only one record, it found: %d", affected)
	}

	return results[0], nil
}
