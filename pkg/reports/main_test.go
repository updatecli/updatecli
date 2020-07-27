package reports

import "testing"

func TestGet(t *testing.T) {
	dataSet := Reports{
		{
			Name:   "My strategy1",
			Result: "\u2714",
			Source: Stage{
				Name:   "Get pluginsite version",
				Kind:   "githubRelease",
				Result: "\u2714",
			},
			Conditions: []Stage{
				{
					Name:   "Test if value set to image",
					Kind:   "yaml",
					Result: "\u2714",
				},
				{
					Name:   "Test if docker image pluginsite exist",
					Kind:   "docker",
					Result: "\u2714",
				},
			},
			Targets: []Stage{
				{
					Name:   "Update value",
					Kind:   "yaml",
					Result: "\u2714",
				},
			},
		},
		{
			Name:   "My strategy2",
			Result: "\u2717",
			Source: Stage{
				Kind:   "mavenRelease",
				Result: "\u2714",
			},
			Conditions: []Stage{
				{
					Kind:   "yaml",
					Result: "\u2714",
				},
				{
					Kind:   "yaml",
					Result: "\u2717",
				},
				{
					Kind:   "docker",
					Result: "\u2717",
				},
			},
			Targets: []Stage{
				{
					Kind:   "yaml",
					Result: "\u2714",
				},
				{
					Kind:   "yaml",
					Result: "\u2714",
				},
			},
		},
	}

	got, err := dataSet.Get()
	if err != nil {
		t.Error(err)
	}
	t.Error(got)
}
