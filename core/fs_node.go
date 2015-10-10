package core

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type NodeMetaInfo struct {
	path    string
	size    int64
	modtime time.Time
	is_dir  bool
}

type NodeList struct {
	list []NodeMetaInfo
}

func (nodes *NodeList) addNodeToList(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	meta := NodeMetaInfo{path: path}
	meta.applyFileInfo(info)
	nodes.list = append(nodes.list, meta)
	return nil
}

func (nodes NodeList) GetList() []NodeMetaInfo {
	return nodes.list
}

func (node *NodeMetaInfo) applyFileInfo(info os.FileInfo) {
	node.size = info.Size()
	node.modtime = info.ModTime()
	node.is_dir = info.IsDir()
}

func (node *NodeMetaInfo) GetNodePath() string {
	return node.path
}

func (node *NodeMetaInfo) isNodeInPath(path string) bool {
	return isPathInBasePath(path, node.path)
}

func GetNodeCurrentFormat() []string {
	return []string{"path", "size", "modtime", "is_dir"}
}

func (node *NodeMetaInfo) ToString() string {
	line := make([]string, 0)
	for _, field := range GetNodeCurrentFormat() {
		value := ""
		switch field {
		case "path":
			value = node.path
		case "size":
			value = strconv.FormatInt(node.size, 10)
		case "modtime":
			value = node.modtime.UTC().Format(time.RFC3339)
		case "is_dir":
			value = strconv.FormatBool(node.is_dir)
		}
		line = append(line, value)
	}
	return strings.Join(line, ",")
}

func GetNodeFromString(nodeString string, format []string) (NodeMetaInfo, error) {
	var err error
	line := strings.Split(nodeString, ",")
	n := len(format)
	if len(line) > n {
		border := len(line) - n + 1
		line[0] = strings.Join(line[0:border], ",")
		line = append(line[0:1], line[border:]...)
	}
	named_line := make(map[string]string)
	for i := 0; i < len(format); i++ {
		named_line[format[i]] = line[i]
	}

	var node NodeMetaInfo
	node.path = named_line["path"]
	node.size, err = strconv.ParseInt(named_line["size"], 10, 64)
	if err != nil {
		return node, err
	}
	node.modtime, err = time.Parse(time.RFC3339, named_line["modtime"])
	if err != nil {
		return node, err
	}
	node.is_dir, err = strconv.ParseBool(named_line["is_dir"])

	return node, err
}
