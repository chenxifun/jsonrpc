package file

import (
	"errors"
	"io/ioutil"
	"os"
	path2 "path"
	"strings"
)

func ReadFile(path string) ([]byte, error) {

	if _, err1 := os.Stat(path); os.IsNotExist(err1) {
		return nil, errors.New("file not found")
	}

	bytes, err := ioutil.ReadFile(path) // nolint: gas
	if err != nil {
		return nil, err
	}
	if bytes == nil {
		return nil, errors.New("file not found")
	}
	return bytes, nil

}

// cover true 覆盖文件，false 检查是否已存在
func WriteFile(data []byte, path string, cover bool) error {

	if !cover {
		if _, err1 := os.Stat(path); !os.IsNotExist(err1) {
			return errors.New("file is exist")
		}
	}

	err := os.MkdirAll(path2.Dir(path), 0700)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0600)

}

func ReadDir(dir string, loop bool) (names []string, err error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {

		if !f.IsDir() {
			names = append(names, path2.Join(dir, f.Name()))
		} else {
			if loop {
				ns, err := ReadDir(path2.Join(dir, f.Name()), loop)
				if err == nil {
					names = append(names, ns...)
				}
			}

		}
	}

	return
}

func IsGoFile(name string) bool {

	if len(name) > 3 {
		return strings.ToLower(name[len(name)-3:]) == ".go"
	}
	return false
}

func IsGoTestFile(name string) bool {
	if len(name) >= 8 {
		return strings.ToLower(name[len(name)-8:]) == "_test.go"
	}
	return false
}
