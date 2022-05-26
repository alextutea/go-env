package env_test

import (
	"github.com/alextutea/go-env"
	"github.com/alextutea/go-env/errs"
	ttest "github.com/alextutea/go-table-tests"
	"testing"
)

type UnitTest struct{}

func Test(t *testing.T) {
	ut := UnitTest{}

	t.Run("UnmarshalMap", ut.TestUnmarshalMap)
	t.Run("UnmarshalMapNestedStructs", ut.TestUnmarshalMapNestedStructs)
}

func (ut *UnitTest) TestUnmarshalMap(t *testing.T) {
	type TestStruct struct {
		Var                string `env:"VAR"`
		MissingVar         string `env:"MISSING_VAR,default=missing"`
		RequiredVar        string `env:"REQUIRED_VAR,required=true"`
		RequiredMissingVar string `env:"REQUIRED_MISSING_VAR,required=true"`
	}

	type TestInput struct {
		EnvMap    map[string]string
		OutStruct TestStruct
	}

	testCases := []ttest.Case{
		{
			In: TestInput{
				EnvMap: map[string]string{
					"VAR":                  "Something",
					"MISSING_VAR":          "Not missing",
					"REQUIRED_VAR":         "Here",
					"REQUIRED_MISSING_VAR": "All good",
				},
				OutStruct: TestStruct{},
			},
			ExpectedOut:      TestStruct{"Something", "Not missing", "Here", "All good"},
			Desc:             "When passing a struct with fields that have the env tag and the corresponding env keys exist",
			ExpectedBehavior: "Should unmarshal the value from env into the given struct",
		},
		{
			In: TestInput{
				EnvMap: map[string]string{
					"VAR":                  "Something",
					"REQUIRED_VAR":         "Here",
					"REQUIRED_MISSING_VAR": "All good",
				},
				OutStruct: TestStruct{},
			},
			ExpectedOut:      TestStruct{"Something", "missing", "Here", "All good"},
			Desc:             "When passing a struct with a field that has the env tag and the default option and the env key does not exist",
			ExpectedBehavior: "Should unmarshal the value from env into the given struct and for the missing env values it should use the defaults",
		},
		{
			In: TestInput{
				EnvMap: map[string]string{
					"VAR":          "Something",
					"MISSING_VAR":  "Not missing",
					"REQUIRED_VAR": "Here",
				},
				OutStruct: TestStruct{},
			},
			ExpectedOut:      TestStruct{"Something", "Not missing", "Here", ""},
			Desc:             "When passing a struct with a required field that is not present in the env values",
			ExpectedBehavior: "Should return a new RequiredKeysNotPresentError",
			ErrTypeCheckFunc: errs.IsRequiredKeyNotPresentError,
		},
	}

	for _, tc := range testCases {
		t.Logf("\t%s", tc.Desc)
		{
			in, _ := tc.In.(TestInput)
			err := env.UnmarshalMap(in.EnvMap, &in.OutStruct)
			ok, msg := tc.Check(in.OutStruct, err)
			if !ok {
				t.Logf(ttest.FailureMessage(tc.ExpectedBehavior))
				t.Fatal(msg)
			}
			t.Logf(ttest.SuccessMessage(tc.ExpectedBehavior))
		}
	}
}

func (ut *UnitTest) TestUnmarshalMapNestedStructs(t *testing.T) {
	type TestSubSubStruct struct {
		SubSubString string `env:"SUB_STRING,default=defaultsubsubstring"`
	}

	type TestSubStruct struct {
		SubBool    bool             `env:"SUB_BOOL,default=false"`
		SubInt64   int64            `env:"SUB_INT64,default=44"`
		SubInt32   int32            `env:"SUB_INT32,default=34"`
		SubInt16   int16            `env:"SUB_INT16,default=54"`
		SubInt8    int8             `env:"SUB_INT8,default=64"`
		SubInt     int              `env:"SUB_INT,default=24"`
		SubFloat64 float64          `env:"SUB_FLOAT64,default=24.654"`
		SubFloat32 float32          `env:"SUB_FLOAT32,default=24.321"`
		SubString  string           `env:"SUB_STRING,default=defaultsubstring"`
		SubStruct  TestSubSubStruct `env:"SUB"`
	}

	type TestSuperStruct struct {
		SuperVar TestSubStruct `env:"SUPER"`
	}

	type TestInput struct {
		EnvMap    map[string]string
		OutStruct TestSuperStruct
	}

	testCases := []ttest.Case{
		{
			In: TestInput{
				EnvMap: map[string]string{
					"SUPER_SUB_BOOL":       "true",
					"SUPER_SUB_STRING":     "substr",
					"SUPER_SUB_INT":        "42",
					"SUPER_SUB_INT32":      "43",
					"SUPER_SUB_INT64":      "44",
					"SUPER_SUB_INT16":      "45",
					"SUPER_SUB_INT8":       "46",
					"SUPER_SUB_FLOAT32":    "42.123",
					"SUPER_SUB_FLOAT64":    "42.456",
					"SUPER_SUB_SUB_STRING": "subsubstr",
				},
				OutStruct: TestSuperStruct{},
			},
			ExpectedOut: TestSuperStruct{SuperVar: TestSubStruct{
				SubBool:    true,
				SubString:  "substr",
				SubInt:     42,
				SubInt32:   43,
				SubInt64:   44,
				SubInt16:   45,
				SubInt8:    46,
				SubFloat32: 42.123,
				SubFloat64: 42.456,
				SubStruct:  TestSubSubStruct{SubSubString: "subsubstr"},
			}},
			Desc:             "When passing nested env vars with different specific types",
			ExpectedBehavior: "Should unmarshal flat env keys correctly into a nested struct and cast them to the appropriate types",
		},
		{
			In: TestInput{
				EnvMap:    map[string]string{},
				OutStruct: TestSuperStruct{},
			},
			ExpectedOut: TestSuperStruct{SuperVar: TestSubStruct{
				SubBool:    false,
				SubString:  "defaultsubstring",
				SubInt:     24,
				SubInt32:   34,
				SubInt64:   44,
				SubInt16:   54,
				SubInt8:    64,
				SubFloat32: 24.321,
				SubFloat64: 24.654,
				SubStruct:  TestSubSubStruct{SubSubString: "defaultsubsubstring"},
			}},
			Desc:             "When passing nested no env vars into a nested struct with default values of various types",
			ExpectedBehavior: "Should unmarshal flat env keys correctly into a nested struct and cast them to the appropriate types",
		},
	}

	for _, tc := range testCases {
		t.Logf("\t%s", tc.Desc)
		{
			in, _ := tc.In.(TestInput)
			err := env.UnmarshalMap(in.EnvMap, &in.OutStruct)
			ok, msg := tc.Check(in.OutStruct, err)
			if !ok {
				t.Logf(ttest.FailureMessage(tc.ExpectedBehavior))
				t.Fatal(msg)
			}
			t.Logf(ttest.SuccessMessage(tc.ExpectedBehavior))
		}
	}
}
