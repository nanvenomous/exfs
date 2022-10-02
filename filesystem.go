package filesystem

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const (
	EDITOR_ENV_VAR = "EDITOR"
)

type FileSystem struct {
}

func NewFileSystem() *FileSystem {
	return &FileSystem{}
}

func (fs *FileSystem) Execute(command string, args []string) error {
	var (
		err error
		cmd *exec.Cmd
	)
	cmd = exec.Command(command, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// Capture outputs the stdout & stderr string responses from a command
func (fs *FileSystem) Capture(command string, args []string) (string, string, error) {
	var (
		err        error
		cmd        *exec.Cmd
		outb, errb bytes.Buffer
	)
	cmd = exec.Command(command, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err = cmd.Run()
	if err != nil {
		return "", errb.String(), err
	}
	return outb.String(), "", nil
}

func (fs *FileSystem) EditTemporaryFile(nm string, txt string) (string, error) {
	var (
		err        error
		editor     string
		tskBodyByt []byte
		tskBody    string
		tempFile   *os.File
	)

	editor = os.Getenv(EDITOR_ENV_VAR)
	if editor == "" {
		return "", errors.New(fmt.Sprintf("Must set %s environment variable.", EDITOR_ENV_VAR))
	}

	tempFile, err = ioutil.TempFile("", nm)
	if err != nil {
		return "", err
	}
	_, err = tempFile.Write([]byte(txt))
	if err != nil {
		return "", err
	}

	err = fs.Execute(editor, []string{tempFile.Name()})
	if err != nil {
		return "", err
	}

	tskBodyByt, err = os.ReadFile(tempFile.Name())
	if err != nil {
		return "", err
	}
	tskBody = string(tskBodyByt)

	if tskBody == "" {
		return "", errors.New("Aborting with empty task body")
	}

	err = tempFile.Close()
	if err != nil {
		return "", err
	}
	err = os.Remove(tempFile.Name())
	if err != nil {
		return "", err
	}
	return tskBody, nil
}

// FindFileInAboveCurDir
// checks each directory to contain the filename argument
// will search current dir & up directory tree until it reaches user home dir
func (fs *FileSystem) FindFileInAboveCurDir(flNm string) (string, error) {
	var (
		err            error
		wd, chkPth, hd string
		wdArr          []string
	)
	hd, err = os.UserHomeDir()
	hd = strings.TrimSuffix(hd, "/")
	if err != nil {
		return "", err
	}
	wd, err = os.Getwd()
	if err != nil {
		return "", err
	}

	if !strings.Contains(wd, hd) {
		return "", errors.New("Cannot search above user home directory.")
	}

	wdArr = strings.Split(strings.TrimSuffix(wd, "/"), "/")
	wdArr = append(wdArr, flNm)

	for {
		chkPth = strings.Join(wdArr, "/")
		if _, err := os.Stat(chkPth); !os.IsNotExist(err) {
			return chkPth, nil
		}
		chkPth = strings.Join(wdArr[:len(wdArr)-1], "/")
		if chkPth == hd {
			return "", errors.New(fmt.Sprintf("Reached user home dir & did not find file: %s", flNm))
		}
		wdArr = wdArr[:len(wdArr)-1]
		wdArr[len(wdArr)-1] = flNm
	}
}
