package chart

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/yaml"
	helm "helm.sh/helm/v3/pkg/chart"

	YAML "sigs.k8s.io/yaml"
)

// Target updates helm chart, it receives the default source value and a dryrun flag
// then return if it changed something or failed
func (c *Chart) Target(source string, dryRun bool) (changed bool, err error) {
	err = c.ValidateTarget()
	if err != nil {
		return false, err
	}

	YamlResource, err := yaml.New(yaml.YamlSpec{
		File: filepath.Join(c.Name, c.File),
		Key:  c.Key,
	})
	if err != nil {
		return false, err
	}

	if len(c.Value) == 0 {
		YamlResource.Spec.Value = source
		c.Value = source
	} else {
		YamlResource.Spec.Value = c.Value
	}

	changed, err = YamlResource.Target(source, dryRun)

	if err != nil {
		return false, err
	} else if err == nil && !changed {
		return false, nil
	}

	// Reset requirements.lock if we modified the file 'requirements.yaml'
	if strings.Compare(c.File, "requirements.yaml") == 0 && !dryRun {
		_, err := c.UpdateRequirements(filepath.Join(c.Name, "requirements.lock"))
		if err != nil {
			return false, err
		}
	}

	// Update Chart.yaml file new Chart Version and appVersion if needed
	err = c.UpdateMetadata(filepath.Join(c.Name, "Chart.yaml"), dryRun)
	if err != nil {
		return false, err
	}

	return true, nil
}

// TargetFromSCM updates helm chart then push changed to a scm, it receives the default source value and dryrun flag
// then return if it changed something or failed
func (c *Chart) TargetFromSCM(source string, scm scm.ScmHandler, dryRun bool) (
	changed bool, files []string, message string, err error) {

	err = c.ValidateTarget()

	if err != nil {
		return false, files, message, err
	}

	filename := filepath.Join(c.Name, c.File)

	YamlResource, err := yaml.New(yaml.YamlSpec{
		File: filename,
		Key:  c.Key,
	})
	if err != nil {
		return false, files, message, err
	}

	if len(c.Value) == 0 {
		YamlResource.Spec.Value = source
		c.Value = source
	} else {
		YamlResource.Spec.Value = c.Value
	}

	changed, files, message, err = YamlResource.TargetFromSCM(source, scm, dryRun)

	if err != nil {
		return false, files, message, err
	} else if err == nil && !changed {
		return false, files, message, nil
	}

	if strings.Compare(c.File, "requirements.yaml") == 0 {
		found, err := c.UpdateRequirements(filepath.Join(scm.GetDirectory(), c.Name, "requirements.lock"))
		if err != nil {
			return false, files, message, err
		}
		if found {
			files = append(files, filepath.Join(c.Name, "requirements.lock"))

		}
	}

	err = c.UpdateMetadata(filepath.Join(scm.GetDirectory(), c.Name, "Chart.yaml"), dryRun)
	if err != nil {
		return false, files, message, err
	}
	files = append(files, filepath.Join(c.Name, "Chart.yaml"))

	return changed, files, message, err
}

// UpdateRequirements test if we are updating the file requirements.yaml
// if it's the case then we also have to delete and recreate the file
// requirements.lock
func (c *Chart) UpdateRequirements(lockFilename string) (bool, error) {
	if strings.Compare(c.File, "requirements.yaml") != 0 {
		return false, fmt.Errorf("No need to cleanup requirements.lock")
	}

	f, err := os.Stat(lockFilename)

	if os.IsExist(err) && !f.IsDir() {
		err = os.Remove(lockFilename)
		if err != nil {
			return false, err
		}
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}

	logrus.Debugf("Something went wrong with lock file %v", lockFilename)
	return false, fmt.Errorf("Something unexpected happened")

}

// UpdateMetadata updates a metadata if necessary and it bump the ChartVersion
func (c *Chart) UpdateMetadata(metadataFilename string, dryRun bool) error {
	md := helm.Metadata{}

	file, err := os.Open(metadataFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)

	if err != nil {
		return err
	}

	err = YAML.Unmarshal(data, &md)

	if err != nil {
		return err
	}

	if len(md.AppVersion) > 0 && c.AppVersion {
		logrus.Debugf("Updating AppVersion from %s to %s\n", md.AppVersion, c.Value)
		md.AppVersion = c.Value
	}

	// Init Chart Version if not set yet
	if len(md.Version) == 0 {
		md.Version = "0.0.0"
	}

	oldVersion := md.Version

	for _, inc := range strings.Split(c.VersionIncrement, ",") {
		v, err := semver.NewVersion(md.Version)
		if err != nil {
			return err
		}

		switch inc {
		case MAJORVERSION:
			md.Version = v.IncMajor().String()
		case MINORVERSION:
			md.Version = v.IncMinor().String()
		case PATCHVERSION:
			md.Version = v.IncPatch().String()
		default:
			logrus.Errorf("Wrong increment rule %q.", inc)
		}
	}

	logrus.Debugf("Update Chart version from %q to %q\n", oldVersion, md.Version)

	if err != nil {
		return err
	}

	if !dryRun {
		data, err := YAML.Marshal(md)
		if err != nil {
			return err
		}

		file, err := os.Create(metadataFilename)
		if err != nil {
			return err
		}

		defer file.Close()

		_, err = file.Write(data)
		if err != nil {
			return err
		}
	}

	return nil
}

//ValidateTarget ensure that target required parameter are set
func (c *Chart) ValidateTarget() error {
	if len(c.File) == 0 {
		c.File = "values.yaml"
	}

	if len(c.Name) == 0 {
		return fmt.Errorf("Parameter name required")
	}

	if len(c.Key) == 0 {
		return fmt.Errorf("Parameter key required")
	}

	if len(c.VersionIncrement) == 0 {
		c.VersionIncrement = MINORVERSION
	}

	for _, inc := range strings.Split(c.VersionIncrement, ",") {

		if inc != MAJORVERSION &&
			inc != MINORVERSION &&
			inc != PATCHVERSION &&
			inc != "" {
			return fmt.Errorf("Unrecognized increment rule %q. accepted values are a comma separated list of [major,minor,patch]", inc)
		}
	}

	return nil
}
