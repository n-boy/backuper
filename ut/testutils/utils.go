package testutils

import (
	"github.com/n-boy/backuper/base"

	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type TestsConfig struct {
	StorageGlacier map[string]string `yaml:"storage_glacier"`
	StorageLocalFS map[string]string `yaml:"storage_localfs"`
}

func GetTestsConfig() (TestsConfig, error) {
	configFilePath := getTestsConfigFilePath()

	config := TestsConfig{}
	if configFilePath == "" {
		return config, fmt.Errorf("config file should be defined as -args config=/path/to/myconfig.yaml")
	}

	yamlContent, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(yamlContent, &config)
	return config, err
}

func getTestsConfigFilePath() string {
	configFilePath := ""
	for _, arg := range os.Args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 && parts[0] == "config" {
			configFilePath = parts[1]
			break
		}
	}

	return configFilePath
}

func TmpDir() string {
	varNames := []string{"TMPDIR", "TMP", "TEMP"}
	val := ""
	for _, v := range varNames {
		val = os.Getenv(v)
		if val != "" {
			break
		}
	}
	if val == "" {
		panic("No tmp path determined")
	}
	return val
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func CalcFileMD5(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func InitAppForTests() {
	testAppConfig := base.DefaultAppConfig
	testAppConfig.LogToStdout = false
	testAppConfig.LogErrToStderr = false

	base.InitLogToDestination(&ioutil.Discard)
}
