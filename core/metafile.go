package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/n-boy/backuper/base"
)

type ArchiveMetafile struct {
	id           int64
	cdate        time.Time
	encrypted    bool
	storage_info map[string]string
	nodes        []NodeMetaInfo
}

type yamlArchiveMetafile struct {
	Encrypted      bool
	StorageInfo    map[string]string `yaml:"storage_info"`
	NodesFormatCSV string            `yaml:"files_format"`
	NodesCSV       []string          `yaml:"files"`
}

func GetArchName(metaFileName string) string {
	return regexp.MustCompile(`_meta\.yaml$`).ReplaceAllString(metaFileName, "")
}

func GetMetaFileName(archName string) string {
	return archName + "_meta.yaml"
}

func GetMetaFileNameEnc(archName string) string {
	return GetMetaFileName(archName) + ".enc"
}

func GetMetaFileGlobMask() string {
	return "archive*_meta.yaml"
}

func CleanMetaFileNameEnc(filename string) (cf string, encrypted bool) {
	cf = filename
	re := regexp.MustCompile(`.enc$`)
	if re.MatchString(filename) {
		cf = re.ReplaceAllString(filename, "")
		encrypted = true
	}
	return
}

func GetMetaFileNameRE() *regexp.Regexp {
	return regexp.MustCompile(`^archive_\d+_\d+_meta.yaml(\.enc)?$`)
}

func GetArchiveFileNameRE() *regexp.Regexp {
	return regexp.MustCompile(`^archive_\d+_\d+.zip$`)
}

func GetMetaFile(metaFilePath string) ArchiveMetafile {
	yamlMF, err := ParseMetaFile(metaFilePath)
	if err != nil {
		base.LogErr.Fatalln(err)
	}

	archMeta := ArchiveMetafile{storage_info: yamlMF.StorageInfo, encrypted: yamlMF.Encrypted}
	parts := strings.Split(filepath.Base(metaFilePath), "_")
	if archMeta.id, err = strconv.ParseInt(parts[1], 10, 64); err != nil {
		base.LogErr.Fatalln(err)
	}
	if archMeta.cdate, err = time.Parse("20060102150405", parts[2]); err != nil {
		base.LogErr.Fatalln(err)
	}

	nodes_format := strings.Split(yamlMF.NodesFormatCSV, ",")
	for _, fileinfo_str := range yamlMF.NodesCSV {
		node, err := GetNodeFromString(fileinfo_str, nodes_format)
		if err != nil {
			base.LogErr.Fatalln(err)
		}
		archMeta.nodes = append(archMeta.nodes, node)
	}
	return archMeta
}

func ParseMetaFile(metaFilePath string) (yamlArchiveMetafile, error) {
	yamlMF := yamlArchiveMetafile{}
	yamlContent, err := ioutil.ReadFile(metaFilePath)
	if err != nil {
		return yamlMF, err
	}

	err = yaml.Unmarshal(yamlContent, &yamlMF)

	return yamlMF, err
}

func NewMetaFile(nodes []NodeMetaInfo, encrypted bool) ArchiveMetafile {
	archMeta := ArchiveMetafile{nodes: nodes, encrypted: encrypted}
	return archMeta
}

func (archMeta ArchiveMetafile) SaveMetaFile(metaFilePath string) error {
	var err error
	yamlMF := yamlArchiveMetafile{}
	yamlMF.StorageInfo = archMeta.storage_info
	yamlMF.Encrypted = archMeta.encrypted
	yamlMF.NodesFormatCSV = strings.Join(GetNodeCurrentFormat(), ",")
	for _, node := range archMeta.nodes {
		yamlMF.NodesCSV = append(yamlMF.NodesCSV, node.ToString())
	}

	yamlData, err := yaml.Marshal(&yamlMF)
	if err != nil {
		return err
	}

	metaFilePathTmp := fmt.Sprint(metaFilePath, "~")
	err = ioutil.WriteFile(metaFilePathTmp, yamlData, 0666)
	if err != nil {
		return err
	}
	return os.Rename(metaFilePathTmp, metaFilePath)
}

func (archMeta ArchiveMetafile) GetStorageInfo() map[string]string {
	return archMeta.storage_info
}

func (archMeta *ArchiveMetafile) SetStorageInfo(storageInfo map[string]string) {
	archMeta.storage_info = storageInfo
}

func (archMeta ArchiveMetafile) GetNodes() []NodeMetaInfo {
	return archMeta.nodes
}

func (archMeta ArchiveMetafile) GetMetaFileId() int64 {
	return archMeta.id
}

func (archMeta ArchiveMetafile) GetMetaFileCreateDate() time.Time {
	return archMeta.cdate
}

func (archMeta ArchiveMetafile) GetMetaFileNameId() string {
	return fmt.Sprintf("%v_%v", archMeta.id, archMeta.cdate.Format("20060102150405"))
}

type MetafileList []string

func (ml MetafileList) Len() int {
	return len(ml)
}
func (ml MetafileList) Swap(i, j int) {
	ml[i], ml[j] = ml[j], ml[i]
}
func (ml MetafileList) Less(i, j int) bool {
	ind_i, err := strconv.Atoi(strings.Split(ml[i], "_")[1])
	if err != nil {
		ind_i = 0
		base.LogErr.Println(err)
	}
	ind_j, err := strconv.Atoi(strings.Split(ml[j], "_")[1])
	if err != nil {
		ind_j = 0
		base.LogErr.Println(err)
	}
	return ind_i < ind_j
}
