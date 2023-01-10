package checksum

import (
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

type Spec struct {
	// Files specifies the list of file that Updatecli monitors to identify state change
	Files []string
}

type Checksum struct {
	output                    *string
	exitCode                  *int
	spec                      Spec
	preCommandMonitoredFiles  map[string]string
	postCommandMonitoredFiles map[string]string
}

func New(spec interface{}, exitCode *int, output *string) (*Checksum, error) {
	var s Spec
	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return nil, err
	}

	err = s.Validate()
	if err != nil {
		return nil, err
	}

	if exitCode == nil {
		return nil, errors.New("exitCode pointer is not set")
	}

	return &Checksum{
		exitCode:                  exitCode,
		output:                    output,
		preCommandMonitoredFiles:  make(map[string]string),
		postCommandMonitoredFiles: make(map[string]string),
		spec:                      s,
	}, nil
}

func (s Spec) Validate() error {
	var errs []error
	if len(s.Files) == 0 {
		errs = append(errs, fmt.Errorf("missing files for monitoring checksum changes"))
	}
	if len(errs) > 0 {
		for i := range errs {
			logrus.Errorln(errs[i])
		}
		return fmt.Errorf("wrong exit spec")
	}
	return nil
}

// PreCommand defines operations needed to be executed before the shell command
func (c *Checksum) PreCommand() error {
	for _, filename := range c.spec.Files {
		c.preCommandMonitoredFiles[filename] = getChecksum(filename)
	}
	return nil
}

// PostCommand defines operations needed to be executed after the shell command
func (c *Checksum) PostCommand() error {
	for _, filename := range c.spec.Files {
		c.postCommandMonitoredFiles[filename] = getChecksum(filename)
	}
	return nil
}

// SourceResult defines the success criteria for a source using the shell resource
func (c *Checksum) SourceResult() (string, error) {
	var missingFiles []string

	changed := false
	for _, filename := range c.spec.Files {
		if c.preCommandMonitoredFiles[filename] == "" || c.postCommandMonitoredFiles[filename] == "" {
			if c.preCommandMonitoredFiles[filename] == "" {
				logrus.Debugf("no checksum for file %q before running shell command", filename)
			}
			if c.postCommandMonitoredFiles[filename] == "" {
				missingFiles = append(missingFiles, filename)
				logrus.Debugf("no checksum for file %q after running shell command", filename)
			}

			changed = true
			continue
		}

		if c.preCommandMonitoredFiles[filename] != c.postCommandMonitoredFiles[filename] {
			logrus.Debugf("File checksum for %q changed after shell command run", filename)
			changed = true
		}
	}

	if len(missingFiles) > 0 {
		for i := range missingFiles {
			logrus.Debugf("Missing files %q", missingFiles[i])
		}
		return *c.output, fmt.Errorf("missing monitored file checksum")
	}

	if !changed {
		return *c.output, fmt.Errorf("monitored checksum changed")
	}

	return *c.output, nil
}

// ConditionResult defines the success criteria for a condition using the shell resource
func (c *Checksum) ConditionResult() (bool, error) {
	var missingFiles []string

	changed := false
	for _, filename := range c.spec.Files {
		if c.preCommandMonitoredFiles[filename] == "" || c.postCommandMonitoredFiles[filename] == "" {
			if c.preCommandMonitoredFiles[filename] == "" {
				logrus.Debugf("no checksum for file %q before running shell command", filename)
			}
			if c.postCommandMonitoredFiles[filename] == "" {
				missingFiles = append(missingFiles, filename)
				logrus.Debugf("no checksum for file %q after running shell command", filename)
			}

			changed = true
			continue
		}

		if c.preCommandMonitoredFiles[filename] != c.postCommandMonitoredFiles[filename] {
			logrus.Debugf("File checksum for %q changed after shell command run", filename)
			changed = true
		}
	}

	if len(missingFiles) > 0 {
		for i := range missingFiles {
			logrus.Debugf("Missing files %q", missingFiles[i])
		}
		return changed, fmt.Errorf("missing monitored file checksum")
	}

	return changed, nil
}

// TargetResult defines the success criteria for a target using the shell resource
func (c *Checksum) TargetResult() (bool, error) {

	var missingFiles []string

	changed := false
	for _, filename := range c.spec.Files {
		if c.preCommandMonitoredFiles[filename] == "" || c.postCommandMonitoredFiles[filename] == "" {
			if c.preCommandMonitoredFiles[filename] == "" {
				logrus.Debugf("no checksum for file %q before running shell command", filename)
			}
			if c.postCommandMonitoredFiles[filename] == "" {
				missingFiles = append(missingFiles, filename)
				logrus.Debugf("no checksum for file %q after running shell command", filename)
			}

			changed = true
			continue
		}

		if c.preCommandMonitoredFiles[filename] != c.postCommandMonitoredFiles[filename] {
			logrus.Debugf("File checksum for %q changed after shell command run", filename)
			changed = true
		}
	}

	if len(missingFiles) > 0 {
		for i := range missingFiles {
			logrus.Debugf("Missing files %q", missingFiles[i])
		}
		return changed, fmt.Errorf("missing monitored file checksum")
	}

	return changed, nil
}
