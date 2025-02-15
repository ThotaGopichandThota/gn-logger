package pipes

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/ThotaGopichandThota/gn-core2/marshal"
	logger "github.com/ThotaGopichandThota/gn-logger"
	"github.com/ThotaGopichandThota/gn-logger/mock"
	"github.com/stretchr/testify/require"
)

func TestParentPart_CannotStartLoopTwice(t *testing.T) {
	part, err := NewParentPart("child-name", &marshal.JsonMarshalizer{})
	require.Nil(t, err)

	err = part.StartLoop(bytes.NewBufferString(""), bytes.NewBufferString(""))
	require.Nil(t, err)

	err = part.StartLoop(bytes.NewBufferString(""), bytes.NewBufferString(""))
	require.Equal(t, ErrInvalidOperationGivenPartLoopState, err)
}

func TestParentPart_CannotStartLoopIfStopped(t *testing.T) {
	part, err := NewParentPart("child-name", &marshal.JsonMarshalizer{})
	require.Nil(t, err)

	err = part.StartLoop(bytes.NewBufferString(""), bytes.NewBufferString(""))
	require.Nil(t, err)

	part.StopLoop()

	err = part.StartLoop(bytes.NewBufferString(""), bytes.NewBufferString(""))
	require.Equal(t, ErrInvalidOperationGivenPartLoopState, err)
}

func TestParentPart_ReceivesLogsFromChildProcess(t *testing.T) {
	mock.ClearAllDummySignals()

	// Record logs by means of a logs gatherer, so we can apply assertions afterwards
	gatherer := &mock.DummyLogsGatherer{}
	logOutputSubject := logger.GetLogOutputSubject()
	_ = logOutputSubject.AddObserver(gatherer, gatherer)

	part, err := NewParentPart("child-name", &marshal.JsonMarshalizer{})
	require.Nil(t, err)
	profileReader, logsWriter := part.GetChildPipes()

	command := exec.Command("./testchild")
	command.ExtraFiles = []*os.File{profileReader, logsWriter}

	childStdout, err := command.StdoutPipe()
	require.Nil(t, err)
	childStderr, err := command.StderrPipe()
	require.Nil(t, err)

	err = command.Start()
	require.Nil(t, err)

	_ = part.StartLoop(childStdout, childStderr)

	mock.WaitForDummySignal("done-step-1")
	require.True(t, gatherer.ContainsLogLine("foo", logger.LogInfo, "foo-info"))
	require.True(t, gatherer.ContainsLogLine("bar", logger.LogInfo, "bar-info"))
	require.False(t, gatherer.ContainsText("foo-trace-no"))
	require.False(t, gatherer.ContainsText("bar-trace-no"))
	require.True(t, gatherer.ContainsLogLine("foo", logger.LogInfo, "foo-in-go"))
	require.True(t, gatherer.ContainsLogLine("bar", logger.LogInfo, "bar-in-go"))

	// Change logs profile
	logger.ToggleLoggerName(true)
	_ = logger.SetLogLevel("*:TRACE")
	logger.NotifyProfileChange()

	mock.WaitForDummySignal("done-step-2")
	require.True(t, gatherer.ContainsLogLine("foo", logger.LogTrace, "foo-trace-yes"))
	require.True(t, gatherer.ContainsLogLine("bar", logger.LogTrace, "bar-trace-yes"))
	require.True(t, gatherer.ContainsLogLine(textOutputSinkName, logger.LogTrace, "child-name"))
	require.True(t, gatherer.ContainsLogLine(textOutputSinkName, logger.LogError, "child-name"))
	require.True(t, gatherer.ContainsText("Here's some stderr"))
	require.True(t, gatherer.ContainsText("Here's some stdout"))
}
