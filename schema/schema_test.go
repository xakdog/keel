package schema_test

import (
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/nsf/jsondiff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamkeel/keel/schema"
	"google.golang.org/protobuf/encoding/protojson"
)

type Error struct {
	Code string `json:"code"`
}

type Errors struct {
	Errors []Error `json:"Errors"`
}

func TestSchema(t *testing.T) {
	testdataDir := "./testdata"
	testCases, err := ioutil.ReadDir(testdataDir)
	require.NoError(t, err)

	toRun := []fs.FileInfo{}
	for _, testCase := range testCases {
		if strings.HasSuffix(testCase.Name(), ".only") {
			toRun = append(toRun, testCase)
		}
	}

	if len(toRun) > 0 {
		testCases = toRun
	}

	for _, testCase := range testCases {
		if !testCase.IsDir() {
			continue
		}

		testCaseDir := testdataDir + "/" + testCase.Name()

		t.Run(testCase.Name(), func(t *testing.T) {

			files, err := ioutil.ReadDir(testCaseDir)
			require.NoError(t, err)

			filesByName := map[string][]byte{}
			for _, f := range files {
				if f.IsDir() {
					continue
				}
				b, err := ioutil.ReadFile(testCaseDir + "/" + f.Name())
				require.NoError(t, err)
				filesByName[f.Name()] = b
			}

			s2m := schema.Builder{}
			protoSchema, err := s2m.MakeFromDirectory(testCaseDir)

			var expectedJSON []byte
			var actualJSON []byte

			if expectedProto, ok := filesByName["proto.json"]; ok {
				require.NoError(t, err)
				expectedJSON = expectedProto
				actualJSON, err = protojson.Marshal(protoSchema)
				require.NoError(t, err)

			} else if expectedErrors, ok := filesByName["errors.json"]; ok {
				require.NotNil(t, err)

				expectedJSON = expectedErrors
				actualJSON, err = json.Marshal(err)
				require.NoError(t, err)

			} else {
				// if no proto.json file or errors.json file is provided then we assume this
				// is a test case that is just expected to parse and validate with no errors
				require.NoError(t, err)
				return
			}

			opts := jsondiff.DefaultConsoleOptions()

			diff, explanation := jsondiff.Compare(expectedJSON, actualJSON, &opts)

			switch diff {
			case jsondiff.FullMatch:
				// success
			case jsondiff.SupersetMatch, jsondiff.NoMatch:
				assert.Fail(t, "actual result does not match expected", explanation)
			case jsondiff.FirstArgIsInvalidJson:
				assert.Fail(t, "expected JSON is invalid")
			case jsondiff.SecondArgIsInvalidJson:
				// highly unlikely (almost impossible) to happen
				assert.Fail(t, "actual JSON (proto or errors) is invalid")
			case jsondiff.BothArgsAreInvalidJson:
				// also highly unlikely (almost impossible) to happen
				assert.Fail(t, "both expected and actual JSON are invalid")
			}

		})
	}
}
