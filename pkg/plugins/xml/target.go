package xml

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/clbanning/mxj/v2"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html/charset"

	"github.com/tomwright/dasel"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/core/scm"
	"github.com/updatecli/updatecli/pkg/core/text"
)

func (x *XML) Target(source string, dryRun bool) (changed bool, err error) {

	changed, _, _, err = x.TargetFromSCM(source, nil, dryRun)
	if err != nil {
		return changed, err
	}

	return changed, err
}

// TargetFromSCM updates a scm repository based on the modified yaml file.
func (x *XML) TargetFromSCM(source string, repository scm.Scm, dryRun bool) (changed bool, files []string, message string, err error) {
	if len(x.spec.Value) == 0 {
		x.spec.Value = source
	}

	// https://github.com/clbanning/mxj/issues/17

	// Test if target reference a file with a prefix like https:// or file://
	// In that case we don't know how to update those files.
	if text.IsURL(x.spec.File) {
		return changed, files, message, fmt.Errorf("unsupported filename prefix")
	}

	targetFile := ""
	if repository != nil {
		targetFile = filepath.Join(repository.GetDirectory(), x.spec.File)
	} else {
		targetFile = x.spec.File
	}

	strData, err := text.ReadAll(targetFile)

	if err != nil {
		return changed, files, message, err
	}

	if len(x.spec.Value) == 0 {
		x.spec.Value = source
	}

	mxj.XMLEscapeCharsDecoder(true)
	mxj.XmlCharsetReader = charset.NewReaderLabel

	// NewMapXml drop comments, directive, process
	data, err := mxj.NewMapXml([]byte(strData), true)

	/* Due to https://github.com/TomWright/dasel/issues/175
	The following code is some experimentation that I keep as notes until I find a better solution
		// Attempt one is to use a NewMapXMLSeq
			data, err := mxj.NewMapXmlSeq([]byte(strData))

		// Attempt two is to use a NewMapXMLReader

			r := bytes.NewReader([]byte(strData))
			data := make(map[string]interface{})
			for {
				v, err := mxj.NewMapXmlSeqReader(r)
				if err != nil {
					if err == io.EOF {
						break
					}
					if err != mxj.NoRoot {
						logrus.Errorln(err)
					}
				}
				for key, val := range v {
					data[key] = val
				}
			}

	*/

	if err != nil {
		return changed, files, message, err
	}

	rootNode := dasel.New(data)

	if rootNode == nil {
		return changed, files, message, ErrDaselFailedParsingXMLByteFormat
	}

	queryResult, err := rootNode.Query(x.spec.Key)
	if err != nil {
		return changed, files, message, err
	}

	if queryResult.String() == x.spec.Value {
		logrus.Infof("%s Key %q, from file %q, already set to %q, nothing else need to do",
			result.SUCCESS,
			x.spec.Key,
			x.spec.File,
			x.spec.Value)
		return changed, files, message, nil
	}

	err = rootNode.Put(x.spec.Key, x.spec.Value)
	if err != nil {
		return changed, files, message, err
	}

	changed = true

	logrus.Infof("%s Key %q, from file %q, updated from  %q to %q",
		result.ATTENTION,
		x.spec.Key,
		x.spec.File,
		queryResult.String(),
		x.spec.Value)

	mapVal, _ := rootNode.InterfaceValue().(mxj.Map)
	//mapVal, _ := rootNode.InterfaceValue(). (mxj.MapSeq)

	xmlVal, err := mapVal.XmlIndent("", "    ")
	if err != nil {
		return changed, files, message, fmt.Errorf("unable to re-encode XML for file %s: %w", x.spec.File, err)
	}

	/* Due to https://github.com/TomWright/dasel/issues/175
	The following code is some experimentation that I keep as notes until I find a better solution

	////	/* Testing 1*/
	//xmlVal := []byte{}
	//for key, val := range data {
	//	b, err := mxj.MapSeq(val.(map[string]interface{})).Xml(key)
	//	if err != nil {
	//		fmt.Println("err:", err)
	//		return changed, err
	//	}
	//	fmt.Println(string(b))
	//	xmlVal = append(xmlVal, b...)
	//}

	//	/* Testing 2*/
	//xmlVal := []byte{}
	//buf := new(bytes.Buffer)
	//writeMap := func(val interface{}) error {
	//	if m, ok := val.(map[string]interface{}); ok {
	//		mv := mxj.New()
	//		for k, v := range m {
	//			mv[k] = v
	//		}
	//		byteData, err := mxj.MapSeq(val.(map[string]interface{})).Xml(key)
	//		if err != nil {
	//			return err
	//		}
	//		buf.Write(byteData)
	//		buf.Write([]byte("\n"))
	//		return nil
	//	}
	//	buf.Write([]byte(fmt.Sprintf("%v\n", val)))
	//	return nil
	//}

	//writeMap(data)
	//xmlVal = buf.Bytes()

	/* end of testing */

	if !dryRun {

		fileInfo, err := os.Stat(targetFile)
		if err != nil {
			return changed, files, message, fmt.Errorf("[%s] unable to get file info: %w", x.spec.File, err)
		}

		logrus.Debugf("fileInfo for %s mode=%s", targetFile, fileInfo.Mode().String())

		user, err := user.Current()
		if err != nil {
			logrus.Errorf("unable to get user info: %s", err)
		}

		logrus.Debugf("user: username=%s, uid=%s, gid=%s", user.Username, user.Uid, user.Gid)

		newFile, err := os.Create(targetFile)
		if err != nil {
			return changed, files, message, fmt.Errorf("unable to write to file %s: %w", targetFile, err)
		}

		defer newFile.Close()

		newFile.Write(xmlVal)

	}

	files = append(files, targetFile)
	message = fmt.Sprintf("Update key %q from file %q", x.spec.Key, x.spec.File)

	return changed, files, message, err

}
