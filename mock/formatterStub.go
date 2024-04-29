package mock

import logger "github.com/ThotaGopichandThota/gn-logger1"

// FormatterStub -
type FormatterStub struct {
	OutputCalled func(line logger.LogLineHandler) []byte
}

// Output -
func (fs *FormatterStub) Output(line logger.LogLineHandler) []byte {
	return fs.OutputCalled(line)
}

// IsInterfaceNil -
func (fs *FormatterStub) IsInterfaceNil() bool {
	return fs == nil
}
