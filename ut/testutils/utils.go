package testutils

import (
	"github.com/n-boy/backuper/base"
	"github.com/n-boy/backuper/core"
	"github.com/n-boy/backuper/storage"

	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
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
var randSeedDone = false

func RandString(n int) string {
	if !randSeedDone {
		rand.Seed(time.Now().UnixNano())
		randSeedDone = true
	}
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

func CreateTestPlan(tfs TestFileSystem, chunkSize int64) core.BackupPlan {
	var err error
	if _, err = os.Stat(tfs.BasePath()); os.IsNotExist(err) {
		panic("test base path is not exists: " + tfs.BasePath())
	}
	if base.GetAppDir() == "" {
		panic("AppDir is not initialized")
	}

	var plan core.BackupPlan
	plan.Name = "test_plan_" + RandString(20)
	plan.ChunkSize = chunkSize
	plan.NodesToArchive = append(plan.NodesToArchive, tfs.DataPath())

	testStorageConfig := map[string]string{
		"path": tfs.StoragePath(),
		"type": "localfs"}
	if plan.Storage, err = storage.NewStorage(testStorageConfig); err != nil {
		panic(err)
	}

	if err = plan.SavePlan(false); err != nil {
		panic(err)
	}

	plan, err = core.GetBackupPlan(plan.Name)
	if err != nil {
		panic(err)
	}

	return plan
}

type changedNode struct {
	Path   string
	Size1  int64
	MD5_1  string
	IsDir1 bool
	Size2  int64
	MD5_2  string
	IsDir2 bool
}

type CompareDirResult struct {
	Equals       bool
	AddedNodes   []string
	MissedNodes  []string
	ChangedNodes []changedNode
}

func (res *CompareDirResult) String() (s string) {
	if res.Equals {
		s = fmt.Sprintf("Compared directorie's content are equals\n")
	} else {
		if len(res.AddedNodes) > 0 {
			s += fmt.Sprintf("Some nodes are new:\n\t%v\n", strings.Join(res.AddedNodes, "\n\t"))
		}
		if len(res.MissedNodes) > 0 {
			s += fmt.Sprintf("Some nodes are missed:\n\t%v\n", strings.Join(res.MissedNodes, "\n\t"))
		}
		if len(res.ChangedNodes) > 0 {
			s += fmt.Sprintf("Some nodes are changed:\n")
			for _, cn := range res.ChangedNodes {
				s += fmt.Sprintf("\t%v\n\t\tstate1: isDir=%v, size=%v, md5=%v\n\t\tstate2: isDir=%v, size=%v, md5=%v\n",
					cn.Path, cn.IsDir1, cn.Size1, cn.MD5_1, cn.IsDir2, cn.Size2, cn.MD5_2)
			}
		}
	}
	return
}

func CompareDirs(path1, path2 string) (res CompareDirResult, err error) {
	var nodes1, nodes2 core.NodeList
	err = filepath.Walk(path1, nodes1.AddNodeToList)
	if err != nil {
		return
	}

	err = filepath.Walk(path2, nodes2.AddNodeToList)
	if err != nil {
		return
	}

	nodes1map, err1 := nodesList2Map(nodes1, path1)
	if err1 != nil {
		return res, err1
	}
	nodes2map, err2 := nodesList2Map(nodes2, path2)
	if err2 != nil {
		return res, err2
	}

	return CompareDirNodes(nodes1map, nodes2map)
}

func CompareDirNodes(nodes1map, nodes2map map[string]core.NodeMetaInfo) (res CompareDirResult, err error) {
	for p, node1 := range nodes1map {
		if node2, ok := nodes2map[p]; ok {
			if node1.IsDir() || node2.IsDir() {
				if node1.IsDir() != node2.IsDir() {
					res.ChangedNodes = append(res.ChangedNodes, changedNode{
						Path:   p,
						IsDir1: node1.IsDir(),
						IsDir2: node2.IsDir(),
					})
				}
			} else {
				if node1.Size() != node2.Size() || node1.Md5() != node2.Md5() {
					res.ChangedNodes = append(res.ChangedNodes, changedNode{
						Path:  p,
						Size1: node1.Size(),
						MD5_1: node1.Md5(),
						Size2: node2.Size(),
						MD5_2: node2.Md5(),
					})
				}
			}
		} else {
			res.MissedNodes = append(res.MissedNodes, p)
		}
	}

	for p, _ := range nodes2map {
		if _, ok := nodes1map[p]; !ok {
			res.AddedNodes = append(res.AddedNodes, p)
		}
	}

	if len(res.AddedNodes) == 0 &&
		len(res.MissedNodes) == 0 &&
		len(res.ChangedNodes) == 0 {
		res.Equals = true
	}

	return
}

func GetDirNodes(path string) (nodesMap map[string]core.NodeMetaInfo, err error) {
	var nodes core.NodeList
	err = filepath.Walk(path, nodes.AddNodeToList)
	if err != nil {
		return
	}

	return nodesList2Map(nodes, path)
}

func nodesList2Map(list core.NodeList, basePath string) (nodesMap map[string]core.NodeMetaInfo, err error) {
	var p string
	nodesMap = make(map[string]core.NodeMetaInfo)
	for _, node := range list.GetList() {
		if !node.IsDir() {
			md5, err1 := CalcFileMD5(node.GetNodePath())
			if err1 != nil {
				return nodesMap, err1
			}
			node.SetMd5(md5)
		}

		p, err = filepath.Rel(basePath, node.GetNodePath())
		if err != nil {
			return
		}
		if p != "." {
			nodesMap[p] = node
		}
	}
	return
}
