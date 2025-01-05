package udash

import (
	"sort"

	"github.com/sirupsen/logrus"
)

// Logout remove a token from the local updatecli configuration file
func Logout(url string) error {

	updatecliConfigPath, err := initConfigFile()
	if err != nil {
		return err
	}

	logrus.Debugf("Updatecli configuration located at %q", updatecliConfigPath)

	data, err := readConfigFile()
	if err != nil {
		return err
	}

	if len(data.Auths) == 0 {
		return nil
	}

	keys := []string{}
	for i := range data.Auths {
		if data.Auths[i].URL == url || i == sanitizeTokenID(url) || data.Auths[i].API == url {
			logrus.Debugf("logout from %q\n", i)
			delete(data.Auths, i)
			continue
		}

		keys = append(keys, i)
	}

	if len(keys) == 0 {
		data.Default = ""
		return nil
	}

	sort.Strings(keys)
	data.Default = keys[0]

	err = writeConfigFile(updatecliConfigPath, data)
	if err != nil {
		return err
	}

	return nil
}
