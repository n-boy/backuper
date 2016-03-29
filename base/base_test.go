package base_test

import (
	"github.com/n-boy/backuper/base"

	"path/filepath"
	"testing"
)

func TestGetFirstLevelPath(t *testing.T) {
	type GFLPTest struct {
		basePath string
		path     string
		result   string
	}
	var tests []GFLPTest

	if filepath.Separator == '/' {
		tests = []GFLPTest{
			{"", "/usr/local/bin", "/usr"},
			{"", "/usr/local/file.txt", "/usr"},
			{"", "/usr", "/usr"},
			{"/", "/usr/local/bin", "/usr"},
			{"/usr", "/usr/local/bin", "/usr/local"},
			{"/usr", "/usr", ""},
			{"/usr", "/etc", ""},
		}
	} else {
		tests = []GFLPTest{
			{"", "D:\\usr\\local\\bin", "D:\\"},
			{"", "D:\\usr\\local\\file.txt", "D:\\"},
			{"", "D:\\", "D:\\"},
			{"D:\\", "D:\\usr\\local\\bin", "D:\\usr"},
			{"D:\\usr", "D:\\usr\\local\\bin", "D:\\usr\\local"},
			{"D:\\usr", "D:\\usr", ""},
			{"D:\\usr", "D:\\etc", ""},
		}
	}

	for _, test := range tests {
		flp := base.GetFirstLevelPath(test.basePath, test.path)
		if flp == test.result {
			t.Logf("Test passed. basePath: %v, path: %v\n", test.basePath, test.path)
		} else {
			t.Errorf("Test failed. basePath: %v, path: %v, expected: %v, got: %v\n",
				test.basePath, test.path, test.result, flp)
		}
	}
}
