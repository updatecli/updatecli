package jenkins

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSemVer(t *testing.T) {
	type test struct {
		v1       string
		v2       string
		lessThan bool
	}

	tests := []test{
		{v1: "1.0.0", v2: "2.0.0", lessThan: true},
		{v1: "1.0.0", v2: "1.1.0", lessThan: true},
		{v1: "1.0.0", v2: "1.0.1", lessThan: true},
		{v1: "1.0.0-A", v2: "1.0.0-B", lessThan: true},
		{v1: "2.0-alpha-1", v2: "2.5", lessThan: true},
		{v1: "2.5", v2: "2.0-alpha-1", lessThan: false},
		{v1: "2.5", v2: "2.0-alpha-2", lessThan: false},
		{v1: "2.5", v2: "2.0-alpha-3", lessThan: false},
		{v1: "2.5", v2: "2.0-alpha-4", lessThan: false},
		{v1: "2.5", v2: "2.0-beta-1", lessThan: false},
		{v1: "2.5", v2: "2.0-beta-2", lessThan: false},
		{v1: "2.4", v2: "2.0-rc-1", lessThan: false},
		{v1: "2.5", v2: "2.0-rc-1", lessThan: false},
		{v1: "2.0-rc-1", v2: "2.0-beta-1", lessThan: false},
		{v1: "2.0-alpha-1", v2: "2.0-beta-1", lessThan: true},
		{v1: "2.0-beta-1", v2: "2.0-alpha-1", lessThan: false},
		{v1: "2.0-beta-1", v2: "2.4", lessThan: true},
		{v1: "2.0-beta-1", v2: "2.0-alpha-3", lessThan: false},
		{v1: "2.0-beta-1", v2: "2.0-alpha-4", lessThan: false},
		{v1: "2.0-beta-1", v2: "2.0-beta-2", lessThan: true},
		{v1: "2.0-beta-1", v2: "2.0-alpha-2", lessThan: false},
		{v1: "2.0-beta-1", v2: "2.0", lessThan: true},
		{v1: "2.0-beta-1", v2: "2.1", lessThan: true},
		{v1: "2.0-beta-1", v2: "2.2", lessThan: true},
		{v1: "2.0-beta-1", v2: "2.3", lessThan: true},
		{v1: "2.3", v2: "2.0-beta-1", lessThan: false},
		{v1: "2.2", v2: "2.0-beta-1", lessThan: false},
		{v1: "2.1", v2: "2.0-beta-1", lessThan: false},
		{v1: "2.0", v2: "2.0-beta-1", lessThan: false},
		{v1: "2.0-alpha-2", v2: "2.0-beta-1", lessThan: true},
		{v1: "2.0-beta-2", v2: "2.0-beta-1", lessThan: false},
		{v1: "2.0-alpha-1", v2: "2.0-beta-1", lessThan: true},
		{v1: "2.0-alpha-2", v2: "2.0-beta-1", lessThan: true},
		{v1: "2.0-alpha-3", v2: "2.0-beta-1", lessThan: true},
		{v1: "2.0-alpha-4", v2: "2.0-beta-1", lessThan: true},
		{v1: "2.4", v2: "2.0-beta-1", lessThan: false},
		{v1: "1.518.JENKINS-14362-jzlib", v2: "1.518", lessThan: true},
		{v1: "1.518", v2: "1.518.JENKINS-14362-jzlib", lessThan: false},
		{v1: "1.513.JENKINS-14362-jzlib", v2: "1.513", lessThan: true},
		{v1: "1.516.JENKINS-14362-jzlib", v2: "1.516", lessThan: true},
		{v1: "2.4", v2: "2.4", lessThan: false},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("compare %s < %s", tc.v1, tc.v2), func(t *testing.T) {
			version1 := NewVersion(tc.v1)
			version2 := NewVersion(tc.v2)
			assert.Equal(t, tc.lessThan, version1.LessThan(version2))
		})
	}
}
