package exfs

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	EDITOR_ENV_VAR = "EDITOR"
)

type OperatingSystemRoute struct {
	Linux   func() error
	Mac     func() error
	Windows func() error
}

func RunOn(osr *OperatingSystemRoute) error {
	switch runtime.GOOS {
	case "linux":
		return osr.Linux()
	case "darwin":
		return osr.Mac()
	case "windows":
		return osr.Windows()
	default:
		return errors.New(fmt.Sprintf("Did not recognize os: %s", runtime.GOOS))
	}
}

func UserConfigDir() {
}

type FileSystem struct {
}

func NewFileSystem() *FileSystem {
	return &FileSystem{}
}

var execute = func(command string, args []string) error {
	var (
		cmd *exec.Cmd
	)
	cmd = exec.Command(command, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Execute runs the command with the stdout, stdin & stderr as the operating systems stdout & stderr
func (fs *FileSystem) Execute(command string, args []string) error {
	return execute(command, args)
}

var capture = func(command string, args []string) (string, string, error) {
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
	return outb.String(), errb.String(), err
}

// Capture outputs the stdout & stderr string responses from a command
func (fs *FileSystem) Capture(command string, args []string) (string, string, error) {
	return capture(command, args)
}

// EditTemporaryFile give a temp file name with extension and some text for the user to edit
// the text will be opened in editor defined by EDITOR environment variable
// returns the text after user edits
func (fs *FileSystem) EditTemporaryFile(editor string, nm string, txt string) (string, error) {
	var (
		err        error
		tskBodyByt []byte
		tskBody    string
		tempFile   *os.File
	)

	tempFile, err = ioutil.TempFile("", nm)
	if err != nil {
		return "", err
	}
	_, err = tempFile.Write([]byte(txt))
	if err != nil {
		return "", err
	}

	err = execute(editor, []string{tempFile.Name()})
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
		err                 error
		wd, chkPth, hd      string
		hmFlInfo, chkFlInfo os.FileInfo
	)
	hd, err = os.UserHomeDir()
	if err != nil {
		return "", err
	}
	hmFlInfo, err = os.Stat(hd)
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

	chkPth = wd

	for {
		chkPth = filepath.Join(chkPth, flNm)
		if chkFlInfo, err = os.Stat(chkPth); !os.IsNotExist(err) {
			return chkPth, nil
		}
		chkPth = filepath.Dir(chkPth)
		if chkFlInfo, err = os.Stat(chkPth); err != nil {
			return "", err
		}
		if os.SameFile(chkFlInfo, hmFlInfo) {
			return "", errors.New(fmt.Sprintf("Reached user home dir & did not find file: %s", flNm))
		}
		chkPth = filepath.Dir(chkPth)
	}
}
