package transformer

import (
	"fmt"
	"strings"
	"testing"
)

type Data struct {
	input          string
	rules          Transformers
	expectedOutput string
	expectedErr    error
}

type DataSet []Data

var (
	dataSet = DataSet{
		Data{
			input: "2.263",
			rules: Transformers{
				Transformer{
					AddPrefix: "alpha-",
				},
			},
			expectedOutput: "alpha-2.263",
		},
		Data{
			input: "2.263",
			rules: Transformers{
				Transformer{
					AddPrefix:           "alpha-",
					DeprecatedAddPrefix: "beta-",
				},
			},
			expectedOutput: "alpha-2.263",
		},
		Data{
			input: "2.263",
			rules: Transformers{
				Transformer{
					DeprecatedAddPrefix: "beta-",
				},
			},
			expectedOutput: "beta-2.263",
		},
		Data{
			input: "1.0.0",
			rules: Transformers{
				Transformer{
					SemVerInc: "wrong",
				},
			},
			expectedOutput: "",
			expectedErr:    fmt.Errorf("unsupported incremental semantic versioning rule \"wrong\", only accept a comma separated list between major, minor, patch"),
		},
		Data{
			input: "1.x.y",
			rules: Transformers{
				Transformer{
					SemVerInc: "major",
				},
			},
			expectedOutput: "",
			expectedErr:    fmt.Errorf("wrong semantic version input: \"1.x.y\""),
		},
		Data{
			input: "1.0.0",
			rules: Transformers{
				Transformer{
					SemVerInc:           "major,minor,patch",
					DeprecatedSemVerInc: "major",
				},
			},
			expectedOutput: "2.1.1",
		},
		Data{
			input: "1.0.0",
			rules: Transformers{
				Transformer{
					SemVerInc: "major,minor,patch",
				},
			},
			expectedOutput: "2.1.1",
		},
		Data{
			input: "1.0.0",
			rules: Transformers{
				Transformer{
					DeprecatedSemVerInc: "major",
				},
			},
			expectedOutput: "2.0.0",
		},
		Data{
			input: "2.263",
			rules: Transformers{
				Transformer{
					AddSuffix: "-jdk11",
				},
			},
			expectedOutput: "2.263-jdk11",
			expectedErr:    nil,
		},
		Data{
			input: "2.263",
			rules: Transformers{
				Transformer{
					AddSuffix:           "-jdk11",
					DeprecatedAddSuffix: "-jdk12",
				},
			},
			expectedOutput: "2.263-jdk11",
			expectedErr:    nil,
		},
		Data{
			input: "2.263",
			rules: Transformers{
				Transformer{
					DeprecatedAddSuffix: "-jdk12",
				},
			},
			expectedOutput: "2.263-jdk12",
			expectedErr:    nil,
		},
		Data{
			input: "alpha-2.263",
			rules: Transformers{
				Transformer{
					TrimPrefix:           "alpha-",
					DeprecatedTrimPrefix: "al",
				},
			},
			expectedOutput: "2.263",
			expectedErr:    nil,
		},
		Data{
			input: "alpha-2.263",
			rules: Transformers{
				Transformer{
					DeprecatedTrimPrefix: "al",
				},
			},
			expectedOutput: "pha-2.263",
			expectedErr:    nil,
		},
		Data{
			input: "alpha-2.263",
			rules: Transformers{
				Transformer{
					TrimPrefix: "alpha-",
				},
			},
			expectedOutput: "2.263",
			expectedErr:    nil,
		},
		Data{
			input: "2.263-jdk11",
			rules: Transformers{
				Transformer{
					TrimSuffix:           "-jdk11",
					DeprecatedTrimSuffix: "11",
				},
			},
			expectedOutput: "2.263",
			expectedErr:    nil,
		},
		Data{
			input: "2.263-jdk11",
			rules: Transformers{
				Transformer{
					DeprecatedTrimSuffix: "11",
				},
			},
			expectedOutput: "2.263-jdk",
			expectedErr:    nil,
		},
		Data{
			input: "2.263-jdk11",
			rules: Transformers{
				Transformer{
					TrimSuffix: "-jdk11",
				},
			},
			expectedOutput: "2.263",
			expectedErr:    nil,
		},
		Data{
			input: "alpha-2.263",
			rules: Transformers{
				Transformer{
					TrimPrefix: "alpha-",
				},
				Transformer{
					TrimPrefix: "2.",
				},
			},
			expectedOutput: "263",
			expectedErr:    nil,
		},
		Data{
			input: "alpha-2.263",
			rules: Transformers{
				Transformer{
					Replacers: Replacers{
						Replacer{
							From: "alpha",
							To:   "beta",
						},
					},
				},
			},
			expectedOutput: "beta-2.263",
			expectedErr:    nil,
		},
		Data{
			input: "alpha-2.263",
			rules: Transformers{
				Transformer{
					Replacer: Replacer{
						From: "al",
						To:   "be",
					},
				},
				Transformer{
					Replacer: Replacer{
						From: "pha",
						To:   "ta",
					},
				},
			},
			expectedOutput: "beta-2.263",
			expectedErr:    nil,
		},
		Data{
			input: "4b7f2b878a9854652493b2c94ac586586f2ab53f93e3baa55fc2199ccd5a042d  terraform_0.14.5_freebsd_amd64.zip",
			rules: Transformers{
				Transformer{
					Find: "terraform_(.*)$",
				},
			},
			expectedOutput: "terraform_0.14.5_freebsd_amd64.zip",
			expectedErr:    nil,
		},
		Data{
			input: "4b7f2b878a9854652493b2c94ac586586f2ab53f93e3baa55fc2199ccd5a042d  terraform_0.14.5_freebsd_amd64.zip",
			rules: Transformers{
				Transformer{
					Find: `^\S*`,
				},
			},
			expectedOutput: "4b7f2b878a9854652493b2c94ac586586f2ab53f93e3baa55fc2199ccd5a042d",
			expectedErr:    nil,
		},
		Data{
			input: "4b7f2b878a9854652493b2c94ac586586f2ab53f93e3baa55fc2199ccd5a042d  terraform_0.14.5_freebsd_amd64.zip",
			rules: Transformers{
				Transformer{
					Find: `\S*$`,
				},
			},
			expectedOutput: "terraform_0.14.5_freebsd_amd64.zip",
			expectedErr:    nil,
		},
		Data{
			input: "1.18.0",
			rules: Transformers{
				Transformer{
					DeprecatedFindSubMatch: `(\d*).(\d*)`,
				},
			},
			expectedOutput: "1.18",
			expectedErr:    nil,
		},
		Data{
			input: "1.18.0",
			rules: Transformers{
				Transformer{
					DeprecatedFindSubMatch: `(\d*).(\d*)`,
					FindSubMatch: FindSubMatch{
						Pattern:      `\d*.(\d*)`,
						CaptureIndex: 1,
					},
				},
			},
			expectedOutput: "18",
			expectedErr:    nil,
		},
		Data{
			input: "noalphanumericvalue",
			rules: Transformers{
				Transformer{
					DeprecatedFindSubMatch: `\d.*`,
				},
			},
			expectedOutput: "",
			expectedErr:    nil,
		},
		Data{
			input: "1.19.0",
			rules: Transformers{
				Transformer{
					FindSubMatch: FindSubMatch{
						Pattern:      `\d*.(\d*)`,
						CaptureIndex: 1,
					},
				},
			},
			expectedOutput: "19",
			expectedErr:    nil,
		},
		Data{
			input: "1.17.0",
			rules: Transformers{
				Transformer{
					FindSubMatch: FindSubMatch{
						Pattern:      `\d*.\d*`,
						CaptureIndex: 1,
					},
				},
			},
			expectedOutput: "",
			expectedErr:    nil,
		},
		Data{
			input: "1.17.0",
			rules: Transformers{
				Transformer{
					FindSubMatch: FindSubMatch{
						Pattern:      `\d*.(\d*).(\d*)`,
						CaptureIndex: 2,
					},
				},
			},
			expectedOutput: "0",
			expectedErr:    nil,
		},
		Data{
			input: "1.17.0",
			rules: Transformers{
				Transformer{
					FindSubMatch: FindSubMatch{
						Pattern:                `\d*.(\d*).(\d*)`,
						DeprecatedCaptureIndex: 2,
					},
				},
			},
			expectedOutput: "0",
			expectedErr:    nil,
		},
		Data{
			input: "1.17.0",
			rules: Transformers{
				Transformer{
					FindSubMatch: FindSubMatch{
						Pattern:                `\d*.(\d*).(\d*)`,
						CaptureIndex:           2,
						DeprecatedCaptureIndex: 1,
					},
				},
			},
			expectedOutput: "0",
			expectedErr:    nil,
		},
		Data{
			input: "1.17.0",
			rules: Transformers{
				Transformer{
					FindSubMatch: FindSubMatch{
						Pattern:      `\d*.(\d*).(\d*)`,
						CaptureIndex: 3,
					},
				},
			},
			expectedOutput: "",
			expectedErr:    nil,
		},
		Data{
			input: "1.17.0",
			rules: Transformers{
				Transformer{
					FindSubMatch: FindSubMatch{
						Pattern:        `(\d*).(\d*).(\d*)`,
						CapturePattern: "major: \\1 minor: \\2 patch: \\3",
						CaptureIndex:   3,
					},
				},
			},
			expectedOutput: "major: 1 minor: 17 patch: 0",
			expectedErr:    nil,
		},
		Data{
			input: "1.17.0",
			rules: Transformers{
				Transformer{
					FindSubMatch: FindSubMatch{
						Pattern:        `(\d*).(\d*).(\d*)`,
						CapturePattern: "\\0 major: \\1",
						CaptureIndex:   3,
					},
				},
			},
			expectedOutput: "1.17.0 major: 1",
			expectedErr:    nil,
		},
		Data{
			input: "1.17.0",
			rules: Transformers{
				Transformer{
					FindSubMatch: FindSubMatch{
						Pattern:        `(\d*).(\d*).(\d*)`,
						CapturePattern: "major: \\1 unknown: \\4",
						CaptureIndex:   3,
					},
				},
			},
			expectedOutput: "major: 1 unknown: \\4",
			expectedErr:    nil,
		},
		Data{
			input: "", // explicit empty value
			rules: Transformers{
				Transformer{
					AddPrefix: "alpha-",
				},
			},
			expectedOutput: "",
			expectedErr:    fmt.Errorf("validation error: transformer input is empty"),
		},
		Data{
			input: "1.17.0",
			rules: Transformers{
				Transformer{
					Quote: true,
				},
			},
			expectedOutput: "\"1.17.0\"",
			expectedErr:    nil,
		},
		Data{
			input: "\"1.17.0\"",
			rules: Transformers{
				Transformer{
					Unquote: true,
				},
			},
			expectedOutput: "1.17.0",
			expectedErr:    nil,
		},
		Data{
			input: "{ \"ver\": \"1.17.0\" }",
			rules: Transformers{
				Transformer{
					JsonMatch: JsonMatch{
						Key: ".ver",
					},
				},
			},
			expectedOutput: "1.17.0",
			expectedErr:    nil,
		},
		Data{
			input: "{ \"ver\": [ \"1.17.0\", \"1.17.1\" ] }",
			rules: Transformers{
				Transformer{
					JsonMatch: JsonMatch{
						Key: ".ver",
					},
				},
			},
			expectedOutput: "[1.17.0 1.17.1]",
			expectedErr:    nil,
		},
		Data{
			input: "{ \"ver\": [ \"1.17.0\", \"1.17.1\" ] }",
			rules: Transformers{
				Transformer{
					JsonMatch: JsonMatch{
						Key: "ver.all()",
					},
				},
			},
			expectedOutput: "",
			expectedErr:    fmt.Errorf("multiple results found for query \"ver.all()\""),
		},
		Data{
			input: "{ \"ver\": [ \"1.17.0\", \"1.17.1\" ] }",
			rules: Transformers{
				Transformer{
					JsonMatch: JsonMatch{
						Key:                 "ver.all()",
						JoinMultipleMatches: ",",
					},
				},
			},
			expectedOutput: "1.17.0,1.17.1",
			expectedErr:    nil,
		},
		Data{
			input: "{ \"ver\": [ \"1.17.0\", \"1.17.1\" ] }",
			rules: Transformers{
				Transformer{
					JsonMatch: JsonMatch{
						Key:                 "ver.all()",
						MultipleMatchAction: "first",
					},
				},
			},
			expectedOutput: "1.17.0",
			expectedErr:    nil,
		},
		Data{
			input: "{ \"ver\": [ \"1.17.0\", \"1.17.1\" ] }",
			rules: Transformers{
				Transformer{
					JsonMatch: JsonMatch{
						Key:                 "ver.all()",
						MultipleMatchAction: "last",
					},
				},
			},
			expectedOutput: "1.17.1",
			expectedErr:    nil,
		},
		Data{
			input: "{ \"key\": \"test\", \"ver\": \"1.17.0\" }",
			rules: Transformers{
				Transformer{
					JsonMatch: JsonMatch{
						Key: ".ver",
					},
				},
			},
			expectedOutput: "1.17.0",
			expectedErr:    nil,
		},
		Data{
			input: "{}",
			rules: Transformers{
				Transformer{
					JsonMatch: JsonMatch{
						Key: ".ver",
					},
				},
			},
			expectedOutput: "",
			expectedErr:    fmt.Errorf("could not access map index: property not found: ver"),
		},
		Data{
			input: "",
			rules: Transformers{
				Transformer{
					JsonMatch: JsonMatch{
						Key: ".ver",
					},
				},
			},
			expectedOutput: "",
			expectedErr:    fmt.Errorf("validation error: transformer input is empty"),
		},
		Data{
			input: "{ \"key\": \"test\" }",
			rules: Transformers{
				Transformer{
					JsonMatch: JsonMatch{
						Key: ".ver",
					},
				},
			},
			expectedOutput: "",
			expectedErr:    fmt.Errorf("could not access map index: property not found: ver"),
		},
		Data{
			input: "bad json",
			rules: Transformers{
				Transformer{
					JsonMatch: JsonMatch{
						Key: ".ver",
					},
				},
			},
			expectedOutput: "",
			expectedErr:    fmt.Errorf("invalid character 'b' looking for beginning of value"),
		},
	}
)

func TestApply(t *testing.T) {
	for i, d := range dataSet {
		got, err := d.rules.Apply(d.input)
		if err != nil {
			if d.expectedErr == nil {
				t.Errorf("Error [%d]:\n\tNot Expected\n\tGot:\t\t%q\n",
					i,
					err)
			} else if strings.Compare(d.expectedErr.Error(),
				err.Error()) != 0 {
				t.Errorf("Error [%d]:\n\tExpected:\t%q\n\tGot:\t\t%q\n",
					i,
					d.expectedErr,
					err)
			}
		}

		if got != d.expectedOutput {
			t.Errorf("[%d]Expected Output '%v', got '%v'", i, d.expectedOutput, got)
		}

	}
}
