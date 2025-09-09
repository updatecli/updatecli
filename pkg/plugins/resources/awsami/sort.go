package awsami

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/sirupsen/logrus"
)

// Sort by CreationDate Asc
// ByAge implements sort.Interface based on the Age field.
type ByCreationDateAsc []types.Image

func (images ByCreationDateAsc) Len() int {
	return len(images)
}

func (images ByCreationDateAsc) Less(i, j int) bool {

	formatDate := time.RFC3339

	dateI, err := time.Parse(formatDate, *images[i].CreationDate)

	if err != nil {
		logrus.Errorln(err)
	}

	dateJ, err := time.Parse(formatDate, *images[j].CreationDate)

	if err != nil {
		logrus.Errorln(err)
	}

	return dateI.Before(dateJ)
}

func (images ByCreationDateAsc) Swap(i, j int) {
	images[i], images[j] = images[j], images[i]
}

// Sort by CreationDate Descendant
// ByAge implements sort.Interface based on the Age field.
type ByCreationDateDesc []types.Image

func (images ByCreationDateDesc) Len() int {
	return len(images)
}

func (images ByCreationDateDesc) Less(i, j int) bool {

	formatDate := time.RFC3339

	dateI, err := time.Parse(formatDate, *images[i].CreationDate)

	if err != nil {
		logrus.Errorln(err)
	}

	dateJ, err := time.Parse(formatDate, *images[j].CreationDate)

	if err != nil {
		logrus.Errorln(err)
	}

	return dateI.After(dateJ)
}

func (images ByCreationDateDesc) Swap(i, j int) {
	images[i], images[j] = images[j], images[i]
}
