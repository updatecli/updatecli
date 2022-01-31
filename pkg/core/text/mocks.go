package text

// type MockFile struct {
// 	Content string
// 	Err     error
// 	Exists  bool
// }

// MockTextRetriever is a stub implementation of the `TextRetriever` interface to be used in our unit test suites.
// It stores the expected Content and Err
type MockTextRetriever struct {
	Content  string
	Location string
	Err      error
	Line     int
	Exists   bool
	// Files    map[string]MockFile
}

func (mtr *MockTextRetriever) ReadLine(location string, line int) (string, error) {
	// mtr.Files[location] = MockFile{
	// 	Content: mtr.Content,

	// }
	mtr.Location = location
	mtr.Line = line
	return mtr.Content, mtr.Err
}

func (mtr *MockTextRetriever) ReadAll(location string) (string, error) {
	mtr.Location = location
	return mtr.Content, mtr.Err
}

func (mtr *MockTextRetriever) WriteLineToFile(lineContent, location string, lineNumber int) error {
	mtr.Location = location
	mtr.Line = lineNumber
	mtr.Content = lineContent
	return mtr.Err
}

func (mtr *MockTextRetriever) WriteToFile(content string, location string) error {
	mtr.Location = location
	mtr.Content = content
	return mtr.Err
}

func (mtr *MockTextRetriever) FileExists(location string) bool {
	return mtr.Exists
}
