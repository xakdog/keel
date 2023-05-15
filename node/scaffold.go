package node

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
)

const (
	FUNCTIONS_DIR = "functions"
)

func Scaffold(dir string) (GeneratedFiles, error) {
	builder := schema.Builder{}

	schema, err := builder.MakeFromDirectory(dir)
	if err != nil {
		return nil, err
	}

	files, err := Generate(context.TODO(), schema)

	if err != nil {
		return nil, err
	}

	err = files.Write(dir)

	if err != nil {
		return nil, err
	}

	functionsDir := filepath.Join(dir, FUNCTIONS_DIR)
	if err := ensureDir(functionsDir); err != nil {
		return nil, err
	}

	generatedFiles := GeneratedFiles{}

	functions := proto.FilterOperations(schema, func(op *proto.Operation) bool {
		return op.Implementation == proto.OperationImplementation_OPERATION_IMPLEMENTATION_CUSTOM
	})

	for _, fn := range functions {
		path := filepath.Join(FUNCTIONS_DIR, fmt.Sprintf("%s.ts", fn.Name))

		_, err = os.Stat(filepath.Join(dir, path))

		if os.IsNotExist(err) {
			generatedFiles = append(generatedFiles, &GeneratedFile{
				Path:     path,
				Contents: writeFunctionWrapper(fn),
			})
		}

	}

	err = generatedFiles.Write(dir)
	if err != nil {
		return nil, err
	}

	return generatedFiles, nil
}

func ensureDir(dirName string) error {
	err := os.Mkdir(dirName, 0700)

	if err == nil || os.IsExist(err) {
		return nil
	} else {
		return err
	}
}

func writeFunctionWrapper(function *proto.Operation) string {
	functionName := casing.ToCamel(function.Name)

	suggestedImplementation := ""
	modelName := casing.ToLowerCamel(function.ModelName)

	switch function.Type {
	case proto.OperationType_OPERATION_TYPE_CREATE:
		suggestedImplementation = fmt.Sprintf(`const %s = await api.models.%s.create(inputs);
	return %s;`, modelName, modelName, modelName)
	case proto.OperationType_OPERATION_TYPE_LIST:
		// todo: fix bang! below
		suggestedImplementation = fmt.Sprintf(`const %ss = await api.models.%s.findMany(inputs.where!);
	return %ss;`, modelName, modelName, modelName)
	case proto.OperationType_OPERATION_TYPE_GET:
		suggestedImplementation = fmt.Sprintf(`const %s = await api.models.%s.findOne(inputs);
	return %s;`, modelName, modelName, modelName)
	case proto.OperationType_OPERATION_TYPE_UPDATE:
		suggestedImplementation = fmt.Sprintf(`const %s = await api.models.%s.update(inputs.where, inputs.values);
	return %s;`, modelName, modelName, modelName)
	case proto.OperationType_OPERATION_TYPE_DELETE:
		suggestedImplementation = fmt.Sprintf(`const %s = await api.models.%s.delete(inputs);
	return %s;`, modelName, modelName, modelName)
	case proto.OperationType_OPERATION_TYPE_READ, proto.OperationType_OPERATION_TYPE_WRITE:
		suggestedImplementation = "// Build something cool"
	}

	return fmt.Sprintf(`import { %s } from '@teamkeel/sdk';

export default %s(async (inputs, api, ctx) => {
	%s
});
	`, functionName, functionName, suggestedImplementation)
}
