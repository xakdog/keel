package operand

import (
	"github.com/samber/lo"
	"github.com/teamkeel/keel/schema/expressions"
	"github.com/teamkeel/keel/schema/parser"
	"github.com/teamkeel/keel/schema/query"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
)

type ExpressionScope struct {
	Parent   *ExpressionScope
	Entities []*ExpressionScopeEntity
}

func (a *ExpressionScope) Merge(b *ExpressionScope) *ExpressionScope {
	return &ExpressionScope{
		Entities: append(a.Entities, b.Entities...),
	}
}

type ExpressionObjectEntity struct {
	Name   string
	Fields []*ExpressionScopeEntity
}

type ExpressionScopeEntity struct {
	Name string

	Object    *ExpressionObjectEntity
	Model     *parser.ModelNode
	Field     *parser.FieldNode
	Literal   *expressions.Operand
	Enum      *parser.EnumNode
	EnumValue *parser.EnumValueNode
	Array     []*ExpressionScopeEntity
	Type      string

	Parent *ExpressionScopeEntity
}

func (e *ExpressionScopeEntity) GetType() string {
	if e.Object != nil {
		return e.Object.Name
	}

	if e.Model != nil {
		return e.Model.Name.Value
	}

	if e.Field != nil {
		return e.Field.Type
	}

	if e.Literal != nil {
		return e.Literal.Type()
	}

	if e.EnumValue != nil {
		return e.Parent.Enum.Name.Value
	}

	if e.Array != nil {
		return expressions.TypeArray
	}

	if e.Type != "" {
		return e.Type
	}

	return ""
}

func (e *ExpressionScopeEntity) AllowedOperators() (operators []string) {
	if e.IsRepeated() {
		operators = append(operators, expressions.ArrayOperators...)
		return operators
	}

	switch {
	case e.Literal != nil:
		t := e.Literal.Type()

		switch t {
		case expressions.TypeBoolean:
			operators = append(operators, expressions.OperatorEquals)
			operators = append(operators, expressions.OperatorAssignment)
		case expressions.TypeNumber:
			operators = append(operators, expressions.LogicalOperators...)
			operators = append(operators, expressions.OperatorAssignment)
		case expressions.TypeNull:
			operators = append(operators, expressions.OperatorEquals)
			operators = append(operators, expressions.AssignmentCondition)
		case expressions.TypeText:
			operators = append(operators, expressions.OperatorEquals)
			operators = append(operators, expressions.OperatorAssignment)
		case expressions.TypeArray:
			operators = append(operators, expressions.ArrayOperators...)
		}
	case e.Model != nil:
		operators = append(operators, expressions.OperatorEquals)
		operators = append(operators, expressions.OperatorAssignment)
	case e.Field != nil || e.Type != "":
		switch e.GetType() {
		case expressions.TypeText, parser.FieldTypeText:
			operators = append(operators, expressions.OperatorEquals)
			operators = append(operators, expressions.OperatorAssignment)
		case expressions.TypeBoolean:
			operators = append(operators, expressions.OperatorEquals)
			operators = append(operators, expressions.OperatorAssignment)
		case expressions.TypeNumber:
			operators = append(operators, expressions.OperatorEquals)
			operators = append(operators, expressions.OperatorAssignment)
			operators = append(operators, expressions.NumericalOperators...)
		default:
			operators = append(operators, expressions.OperatorEquals)
			operators = append(operators, expressions.OperatorAssignment)
		}
	}

	return operators
}

func DefaultExpressionScope(asts []*parser.AST) *ExpressionScope {
	entities := []*ExpressionScopeEntity{
		{
			Name: "ctx",
			Object: &ExpressionObjectEntity{
				Name: "Context",
				Fields: []*ExpressionScopeEntity{
					{
						Name:  "identity",
						Model: query.Model(asts, "Identity"),
					},
					{
						Name: "now",
						Type: parser.FieldTypeDatetime,
					},
				},
			},
		},
	}

	for _, enum := range query.Enums(asts) {
		entities = append(entities, &ExpressionScopeEntity{
			Name: enum.Name.Value,
			Enum: enum,
		})
	}

	return &ExpressionScope{
		Entities: entities,
	}
}

// IsRepeated returns true if the entity is a repeated value
// This can be because it is a literal array e.g. [1,2,3]
// or because it's a repeated field or at least one parent
// entity is a repeated field e.g. order.items.product.price
// would be a list of prices (assuming order.items is an
// array of items)
func (e *ExpressionScopeEntity) IsRepeated() bool {
	entity := e
	if len(entity.Array) > 0 {
		return true
	}
	if entity.Field != nil && entity.Field.Repeated {
		return true
	}
	for entity.Parent != nil {
		entity = entity.Parent
		if entity.Field != nil && entity.Field.Repeated {
			return true
		}
	}
	return false
}

func scopeFromModel(parentScope *ExpressionScope, parentEntity *ExpressionScopeEntity, model *parser.ModelNode) *ExpressionScope {
	newEntities := []*ExpressionScopeEntity{}

	for _, field := range query.ModelFields(model) {
		newEntities = append(newEntities, &ExpressionScopeEntity{
			Name:   field.Name.Value,
			Field:  field,
			Parent: parentEntity,
		})
	}

	return &ExpressionScope{
		Entities: newEntities,
		Parent:   parentScope,
	}
}

func scopeFromObject(parentScope *ExpressionScope, parentEntity *ExpressionScopeEntity) *ExpressionScope {
	newEntities := []*ExpressionScopeEntity{}

	for _, entity := range parentEntity.Object.Fields {
		// create a shallow copy by getting the _value_ of entity
		entityCopy := *entity
		// update parent (this does _not_ mutate entity)
		entityCopy.Parent = parentEntity
		// then add a pointer to the _copy_
		newEntities = append(newEntities, &entityCopy)
	}

	return &ExpressionScope{
		Entities: newEntities,
		Parent:   parentScope,
	}
}

func scopeFromEnum(parentScope *ExpressionScope, parentEntity *ExpressionScopeEntity) *ExpressionScope {
	newEntities := []*ExpressionScopeEntity{}

	for _, value := range parentEntity.Enum.Values {
		newEntities = append(newEntities, &ExpressionScopeEntity{
			Name:      value.Name.Value,
			EnumValue: value,
			Parent:    parentEntity,
		})
	}

	return &ExpressionScope{
		Entities: newEntities,
		Parent:   parentScope,
	}
}

// Given an operand of a condition, tries to resolve the relationships defined within the operand
// e.g if the operand is of type "Ident", and the ident is post.author.name
func ResolveOperand(asts []*parser.AST, operand *expressions.Operand, scope *ExpressionScope) (entity *ExpressionScopeEntity, err error) {
	if ok, _ := operand.IsLiteralType(); ok {

		// If it is an array literal then handle differently.
		if operand.Type() == expressions.TypeArray {

			array := []*ExpressionScopeEntity{}

			for _, item := range operand.Array.Values {
				array = append(array,
					&ExpressionScopeEntity{
						Literal: item,
					},
				)
			}

			entity = &ExpressionScopeEntity{
				Array: array,
			}

			return entity, nil
		} else {
			entity = &ExpressionScopeEntity{
				Literal: operand,
			}
			return entity, nil
		}

	}

	// We want to loop over every fragment in the Ident, each time checking if the Ident matches anything
	// stored in the expression scope.
	// e.g if the first ident fragment is "ctx", and the ExpressionScope has a matching key
	// (which it does if you use the DefaultExpressionScope)
	// then it will continue onto the next fragment, setting the new scope to Ctx
	// so that the next fragment can be compared to fields that exist on the Ctx object
fragments:
	for _, fragment := range operand.Ident.Fragments {
		for _, e := range scope.Entities {
			if e.Name != fragment.Fragment {
				continue
			}

			switch {
			case e.Model != nil:
				scope = scopeFromModel(scope, e, e.Model)
			case e.Field != nil:
				model := query.Model(asts, e.Field.Type)

				if model == nil {
					// Did not find the model matching the field
					scope = &ExpressionScope{
						Parent: scope,
					}
				} else {
					scope = scopeFromModel(scope, e, model)
				}
			case e.Object != nil:
				scope = scopeFromObject(scope, e)
			case e.Enum != nil:
				scope = scopeFromEnum(scope, e)
			case e.EnumValue != nil:
				scope = &ExpressionScope{
					Parent: scope,
				}
			case e.Type != "":
				scope = &ExpressionScope{
					Parent: scope,
				}
			}

			entity = e
			continue fragments
		}

		// Suggest the names of all things in scope
		inScope := lo.Map(scope.Entities, func(e *ExpressionScopeEntity, _ int) string {
			return e.Name
		})

		suggestions := errorhandling.NewCorrectionHint(inScope, fragment.Fragment)

		parent := ""
		if entity != nil {
			parent = entity.GetType()
		}

		err = errorhandling.NewValidationError(
			errorhandling.ErrorUnresolvableExpression,
			errorhandling.TemplateLiterals{
				Literals: map[string]string{
					"Fragment":   fragment.Fragment,
					"Parent":     parent,
					"Suggestion": suggestions.ToString(),
				},
			},
			fragment,
		)

		return nil, err
	}

	return entity, nil
}
