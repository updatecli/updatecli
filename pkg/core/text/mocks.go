package text

// MockTextRetriever is a stub implementation of the `TextRetriever` interface to be used in our unit test suites.
// It stores the expected Content and Err
type MockTextRetriever struct {
	Errs     map[string]error
	Lines    map[string][]int
	Contents map[string]string
}

func (mtr *MockTextRetriever) ReadLine(location string, line int) (string, error) {
	mtr.Lines[location] = append(mtr.Lines[location], line)
	return mtr.Contents[location], mtr.Errs[location]
}

func (mtr *MockTextRetriever) ReadAll(location string) (string, error) {
	return mtr.Contents[location], mtr.Errs[location]
}

func (mtr *MockTextRetriever) WriteLineToFile(lineContent, location string, lineNumber int) error {
	mtr.Lines[location] = append(mtr.Lines[location], lineNumber)
	mtr.Contents[location] = lineContent
	return mtr.Errs[location]
}

func (mtr *MockTextRetriever) WriteToFile(content string, location string) error {
	mtr.Contents[location] = content
	return mtr.Errs[location]
}

func (mtr *MockTextRetriever) FileExists(location string) bool {
	_, exists := mtr.Contents[location]
	return exists
}
