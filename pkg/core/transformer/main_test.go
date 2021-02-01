package transformer

import (
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
					"addPrefix": "alpha-",
				},
			},
			expectedOutput: "alpha-2.263",
		},
		Data{
			input: "2.263",
			rules: Transformers{
				Transformer{
					"addSuffix": "-jdk11",
				},
			},
			expectedOutput: "2.263-jdk11",
			expectedErr:    nil,
		},
		Data{
			input: "alpha-2.263",
			rules: Transformers{
				Transformer{
					"trimPrefix": "alpha-",
				},
			},
			expectedOutput: "2.263",
			expectedErr:    nil,
		},
		Data{
			input: "2.263-jdk11",
			rules: Transformers{
				Transformer{
					"trimSuffix": "-jdk11",
				},
			},
			expectedOutput: "2.263",
			expectedErr:    nil,
		},
		Data{
			input: "alpha-2.263",
			rules: Transformers{
				Transformer{
					"trimPrefix": "alpha-",
				},
				Transformer{
					"trimPrefix": "2.",
				},
			},
			expectedOutput: "263",
			expectedErr:    nil,
		},
		Data{
			input: "alpha-2.263",
			rules: Transformers{
				Transformer{
					"replacers": Replacers{
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
					"replacer": Replacer{
						From: "alpha",
						To:   "beta",
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
					"find": "terraform_(.*)$",
				},
			},
			expectedOutput: "terraform_0.14.5_freebsd_amd64.zip",
			expectedErr:    nil,
		},
		Data{
			input: "4b7f2b878a9854652493b2c94ac586586f2ab53f93e3baa55fc2199ccd5a042d  terraform_0.14.5_freebsd_amd64.zip",
			rules: Transformers{
				Transformer{
					"find": `^\S*`,
				},
			},
			expectedOutput: "4b7f2b878a9854652493b2c94ac586586f2ab53f93e3baa55fc2199ccd5a042d",
			expectedErr:    nil,
		},
		Data{
			input: "4b7f2b878a9854652493b2c94ac586586f2ab53f93e3baa55fc2199ccd5a042d  terraform_0.14.5_freebsd_amd64.zip",
			rules: Transformers{
				Transformer{
					"find": `\S*$`,
				},
			},
			expectedOutput: "terraform_0.14.5_freebsd_amd64.zip",
			expectedErr:    nil,
		},
	}
)

func TestApply(t *testing.T) {
	for _, d := range dataSet {
		got, err := d.rules.Apply(d.input)
		if err != nil {
			t.Errorf("%v\n", err)
		}

		if got != d.expectedOutput {
			t.Errorf("Expected Output '%v', got '%v'", d.expectedOutput, got)
		}

	}
}
