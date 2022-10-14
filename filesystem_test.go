package filesystem

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	MockFilesystem *FileSystem
	projectRoot    string
)

func teardown(t *testing.T) {
	var err error
	err = os.Chdir(projectRoot)
	if err != nil {
		panic(err)
	}
	err = os.RemoveAll(t.Name())
	if err != nil {
		panic(err)
	}
}

func setup(t *testing.T) func(*testing.T) {
	var err error

	err = os.Mkdir(t.Name(), os.ModePerm)
	if err != nil {
		panic(err)
	}

	err = os.Chdir(t.Name())
	if err != nil {
		panic(err)
	}
	return teardown
}

func TestMain(m *testing.M) {
	var err error
	MockFilesystem = NewFileSystem()
	projectRoot, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestExecute(t *testing.T) {
	tdn := setup(t)
	defer tdn(t)

	var err error
	err = MockFilesystem.Execute("touch", []string{"hello.txt"})
	assert.ErrorIs(t, err, nil)
	err = MockFilesystem.Execute("notACommand", []string{})
	assert.ErrorContains(t, err, "file not found")
}

func TestCapture(t *testing.T) {
	tdn := setup(t)
	defer tdn(t)

	var (
		err        error
		errs, outs string
	)
	outs, errs, err = MockFilesystem.Capture("touch", []string{"hello.txt"})
	assert.ErrorIs(t, err, nil)
	assert.Empty(t, errs)
	assert.Empty(t, outs)

	mockStdErr := "mock_standard_error_output"
	outs, errs, err = MockFilesystem.Capture("logger", []string{"-s", mockStdErr})
	assert.Contains(t, errs, mockStdErr)
	assert.Nil(t, err)
}

func TestFindFileInAboveCurDir(t *testing.T) {
	tdn := setup(t)
	defer tdn(t)

	var (
		err     error
		fullPth string
		fl      *os.File
	)
	const (
		mockFileName = "fl.txt"
		mockDir      = "mockDir"
	)

	// Fails to find file
	fullPth, err = MockFilesystem.FindFileInAboveCurDir(mockFileName)
	assert.ErrorContains(t, err, "did not find file")
	assert.Equal(t, fullPth, "")

	// Finds file in curr dir
	fl, err = os.Create(mockFileName)
	assert.Nil(t, err)
	assert.Nil(t, fl.Close())
	fullPth, err = MockFilesystem.FindFileInAboveCurDir(mockFileName)
	assert.Nil(t, err)
	assert.Contains(t, fullPth, path.Join(t.Name(), mockFileName))

	// Finds file above cur directory
	assert.Nil(t, os.Mkdir(mockDir, os.ModePerm))
	assert.Nil(t, os.Chdir(mockDir))
	fullPth, err = MockFilesystem.FindFileInAboveCurDir(mockFileName)
	assert.Nil(t, err)
	assert.Contains(t, fullPth, path.Join(t.Name(), mockFileName))
}
