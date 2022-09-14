package query

import (
	"github.com/teamkeel/keel/schema/parser"
)

func APIs(asts []*parser.AST) (res []*parser.APINode) {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.API != nil {
				res = append(res, decl.API)
			}
		}
	}
	return res
}

type ModelFilter func(m *parser.ModelNode) bool

func ExcludeBuiltInModels(m *parser.ModelNode) bool {
	return !m.BuiltIn
}

func Models(asts []*parser.AST, filters ...ModelFilter) (res []*parser.ModelNode) {
	for _, ast := range asts {
	models:
		for _, decl := range ast.Declarations {
			if decl.Model != nil {
				for _, filter := range filters {
					if !filter(decl.Model) {
						continue models
					}
				}

				res = append(res, decl.Model)
			}
		}
	}
	return res
}

func ModelNames(asts []*parser.AST, filters ...ModelFilter) (res []string) {
	for _, ast := range asts {

	models:
		for _, decl := range ast.Declarations {
			if decl.Model != nil {
				for _, filter := range filters {
					if !filter(decl.Model) {
						continue models
					}
				}

				res = append(res, decl.Model.Name.Value)
			}
		}
	}

	return res
}

func Model(asts []*parser.AST, name string) *parser.ModelNode {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Model != nil && decl.Model.Name.Value == name {
				return decl.Model
			}
		}
	}
	return nil
}

func IsModel(asts []*parser.AST, name string) bool {
	return Model(asts, name) != nil
}

func IsIdentityModel(asts []*parser.AST, name string) bool {
	return name == parser.ImplicitIdentityModelName
}

func ModelAttributes(model *parser.ModelNode) (res []*parser.AttributeNode) {
	for _, section := range model.Sections {
		if section.Attribute != nil {
			res = append(res, section.Attribute)
		}
	}
	return res
}

func Enums(asts []*parser.AST) (res []*parser.EnumNode) {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Enum != nil {
				res = append(res, decl.Enum)
			}
		}
	}
	return res
}

func Enum(asts []*parser.AST, name string) *parser.EnumNode {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Enum != nil && decl.Enum.Name.Value == name {
				return decl.Enum
			}
		}
	}
	return nil
}

func IsEnum(asts []*parser.AST, name string) bool {
	return Enum(asts, name) != nil
}

func Roles(asts []*parser.AST) (res []*parser.RoleNode) {
	for _, ast := range asts {
		for _, decl := range ast.Declarations {
			if decl.Role != nil {
				res = append(res, decl.Role)
			}
		}
	}
	return res
}

func IsUserDefinedType(asts []*parser.AST, name string) bool {
	return Model(asts, name) != nil || Enum(asts, name) != nil
}

func UserDefinedTypes(asts []*parser.AST) (res []string) {
	for _, model := range Models(asts) {
		res = append(res, model.Name.Value)
	}
	for _, enum := range Enums(asts) {
		res = append(res, enum.Name.Value)
	}
	return res
}

func ModelActions(model *parser.ModelNode) (res []*parser.ActionNode) {
	return append(ModelOperations(model), ModelFunctions(model)...)
}

func ModelOperations(model *parser.ModelNode) (res []*parser.ActionNode) {
	for _, section := range model.Sections {
		res = append(res, section.Operations...)
	}
	return res
}

func ModelFunctions(model *parser.ModelNode) (res []*parser.ActionNode) {
	for _, section := range model.Sections {
		res = append(res, section.Functions...)
	}
	return res
}

type ModelFieldFilter func(f *parser.FieldNode) bool

func ExcludeBuiltInFields(f *parser.FieldNode) bool {
	return !f.BuiltIn
}

func ModelFields(model *parser.ModelNode, filters ...ModelFieldFilter) (res []*parser.FieldNode) {
	for _, section := range model.Sections {
		if section.Fields == nil {
			continue
		}

	fields:
		for _, field := range section.Fields {
			for _, filter := range filters {
				if !filter(field) {
					continue fields
				}
			}

			res = append(res, field)
		}
	}
	return res
}

func ModelField(model *parser.ModelNode, name string) *parser.FieldNode {
	for _, section := range model.Sections {
		for _, field := range section.Fields {
			if field.Name.Value == name {
				return field
			}
		}
	}
	return nil
}

func FieldHasAttribute(field *parser.FieldNode, name string) bool {
	for _, attr := range field.Attributes {
		if attr.Name.Value == name {
			return true
		}
	}
	return false
}

func FieldIsUnique(field *parser.FieldNode) bool {
	return FieldHasAttribute(field, parser.AttributePrimaryKey) || FieldHasAttribute(field, parser.AttributeUnique)
}

func ModelFieldNames(model *parser.ModelNode) []string {
	names := []string{}
	for _, field := range ModelFields(model, ExcludeBuiltInFields) {
		names = append(names, field.Name.Value)
	}
	return names
}

// ResolveInputType returns a string represention of the type of the give input
// If the input is explicitly typed using a built in type that type is returned
//
//	example: (foo: Text) -> Text is returned
//
// If `i` refers to a field on the parent model (or a nested field) then the type of that field is returned
//
//	example: (foo: some.field) -> The type of `field` on the model referrred to by `some` is returned
func ResolveInputType(asts []*parser.AST, input *parser.ActionInputNode, parentModel *parser.ModelNode) string {
	// handle built-in type
	if parser.IsBuiltInFieldType(input.Type.ToString()) {
		return input.Type.ToString()
	}

	field := ResolveInputField(asts, input, parentModel)
	if field != nil {
		return field.Type
	}

	return ""
}

// ResolveInputField returns the field that the input's type references
func ResolveInputField(asts []*parser.AST, input *parser.ActionInputNode, parentModel *parser.ModelNode) (field *parser.FieldNode) {
	// handle built-in type
	if parser.IsBuiltInFieldType(input.Type.ToString()) {
		return nil
	}

	// follow the idents of the type from the current model to wherever it leads...
	model := parentModel
	for _, fragment := range input.Type.Fragments {
		if model == nil {
			return nil
		}
		field = ModelField(model, fragment.Fragment)
		if field == nil {
			return nil
		}
		model = Model(asts, field.Type)
	}

	return field
}
