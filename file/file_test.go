package file

import (
	"fmt"
	"testing"
)

func TestReadDir(t *testing.T) {

	baseDir := "D:\\GoPath\\src\\github.com\\chenxifun\\jsonrpc\\test"
	ns, err := ReadDir(baseDir, false)

	if err != nil {
		t.Fatal(err)
	}

	for _, n := range ns {

		fmt.Println(n, IsGoFile(n), IsGoTestFile(n))
	}

}
