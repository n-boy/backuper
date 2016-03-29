// Code generated by go-bindata.
// sources:
// webui/templates/archived_list.html
// webui/static/styles.css
// DO NOT EDIT!

package base

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _webuiTemplatesArchived_listHtml = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xa4\x54\x4d\x8f\xd3\x30\x10\x3d\xef\xfe\x8a\xc1\xaa\x38\x6d\x13\x84\x38\x20\x70\x8c\x16\xed\x01\xa4\x6a\xb5\xa2\x12\xd7\xca\x89\xa7\x8d\x85\xeb\x44\xb6\xdb\x6e\x89\xf2\xdf\x19\x3b\x69\xd9\x54\x15\x48\x70\x9a\x78\x3e\x9e\x67\xde\x1b\x87\xd7\x61\x6b\xc4\x2d\xaf\x51\x2a\x71\x7b\xc3\x83\x0e\x06\xc5\x42\xfb\x00\xcd\x1a\xa4\xab\x6a\xbd\x47\x05\x3a\xe0\xd6\xc3\xba\x71\xd0\x1a\x69\xa1\xeb\xb2\x27\xb2\x8f\x72\x8b\x7d\xcf\xf3\xa1\xe8\x96\xca\x8d\xb6\x3f\x20\x1c\x5b\x2c\x58\xc0\xe7\x90\x57\xde\x33\x70\x68\x0a\xe6\xc3\xd1\xa0\xaf\x11\x03\x83\xda\xe1\xba\x60\xb9\x0f\x32\xe8\x2a\x1f\x22\x59\x4c\xa5\x46\xf2\xa1\x13\x5e\x36\xea\x48\x46\xe9\x3d\x68\x55\xb0\x78\x5c\x55\x8d\x0d\x52\x5b\x74\x2c\xb6\x7a\x0a\xb5\x72\x83\xab\x58\x15\xfd\xa7\xce\xf9\x4e\x4c\x9b\xe7\xf9\x4e\xfc\x1e\x80\x97\xe2\x62\x86\x52\xf0\x9c\x10\xa7\xc0\xa1\x5e\x19\x02\x8c\xd7\xdd\x74\x1d\x38\x69\x37\x08\x33\x6d\x15\x3e\xdf\xc1\xac\x85\x0f\x05\x64\x9f\xa5\xc7\x27\xca\x4c\x37\xf7\x3d\x65\x72\x39\x4e\xf8\xa9\x1c\x63\x45\xd7\xcd\xda\x2c\x7e\xf5\x3d\x13\xe9\xb0\xac\x1b\x17\x06\x0f\xcf\xa5\x20\x4a\x67\x29\x61\x89\xad\x74\x32\x34\x2e\x41\xd1\xa5\x68\x55\x82\x9d\xb4\x57\x19\xe9\x3d\x71\x2c\x4b\x83\x53\x5a\x48\xc2\xe8\x84\x0a\x8d\xf1\xad\xac\xb4\xdd\x14\xec\x0d\x4b\xe7\x56\x2a\x95\xce\xef\x18\x94\x8d\x23\xc2\x62\x68\x28\x72\xd1\xdc\xf0\x57\xf3\x79\x50\xe0\x9a\x03\xd5\xda\x82\xbd\x65\xe2\xb5\x2d\x7d\xfb\x91\x44\x56\xf3\xf9\x90\x13\x94\x88\x9d\x46\xd7\xc9\x01\x07\xad\x68\x4c\xf6\x9e\xe0\xee\x47\xde\xef\x60\x21\x89\x12\x87\xfb\xec\xef\xa9\xf7\xc6\xfc\x31\x73\xd1\x54\xd2\x98\x23\x28\x34\x18\x62\x41\x79\x04\x73\x09\x4f\xd6\x5d\x97\x4a\xdb\x46\x61\x92\xeb\x91\x3e\xfc\x0b\xad\x26\x83\x5f\x1f\x96\xf4\xec\x3a\xbd\x1e\x51\xb2\xaf\xfe\x41\x93\x3a\x57\x24\x1e\xe2\xa3\xcc\x5d\x47\xca\xf5\xbd\x38\xfb\x2f\x14\x9f\x4c\x2a\x8d\xde\x10\xdd\x4e\x6f\xea\x10\x17\x64\xad\xe9\x49\xe8\x9f\xf8\x65\xb7\x95\xf6\xbb\xc6\xc3\xe9\xf2\x48\xe9\x37\xdc\x2f\x29\x96\x5e\xde\x3f\x40\x10\xd5\xff\x89\x90\xc4\x78\x18\x94\x98\xe2\xbc\x50\xe0\xb4\xb7\xd1\x17\x37\x52\x9c\x57\xf8\x6c\xc6\x37\x9e\xa7\x7f\xd0\xaf\x00\x00\x00\xff\xff\x5c\x8e\x3d\xa8\x8a\x04\x00\x00")

func webuiTemplatesArchived_listHtmlBytes() ([]byte, error) {
	return bindataRead(
		_webuiTemplatesArchived_listHtml,
		"webui/templates/archived_list.html",
	)
}

func webuiTemplatesArchived_listHtml() (*asset, error) {
	bytes, err := webuiTemplatesArchived_listHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "webui/templates/archived_list.html", size: 1162, mode: os.FileMode(420), modTime: time.Unix(1454161424, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _webuiStaticStylesCss = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xec\x56\x5b\x6b\xdb\x30\x14\x7e\x6e\x7e\x85\x68\x28\x34\x25\x4a\x15\x7a\x5b\x6c\xf6\x30\xb6\x97\x3d\xec\x6d\x0f\x7b\x2b\xb2\x25\xdb\xa2\xb2\x65\x64\xf5\x1a\xfa\xdf\xa7\x9b\x13\x3b\x95\x9d\x6c\x94\xc2\x60\x81\x80\xed\x73\xf4\x9d\x4f\x47\xdf\x39\x47\x89\x20\xcf\x60\x3d\x39\xca\x44\xa5\x60\x86\x4b\xc6\x9f\x23\xf0\x45\x32\xcc\x63\xff\xb1\x61\x2f\x34\x02\xcb\x8b\xfa\x29\x9e\xbc\x4e\x26\x53\xb3\xe2\x36\xd5\x16\xcc\x2a\x2a\xcd\xda\x47\x46\x54\x11\x81\x4f\x08\x19\x9f\xa3\x12\xcb\x9c\x55\x11\x40\x00\xdf\x2b\xe1\x16\xd5\x58\x15\xb7\x9c\x35\xca\xf8\xd7\x98\x10\x56\xe5\x11\xb8\xaa\x9f\x00\x6a\x1d\x72\x7a\x5b\x50\x4c\x1c\x64\x37\xf2\xb5\x45\x55\xf4\x49\x41\xcc\x59\xae\x91\x53\x5a\x29\x2a\xe3\x3e\xd2\x26\x32\x4c\x84\x52\xa2\x8c\xc0\x85\xe3\x93\xe0\xf4\x2e\x97\xe2\xbe\x22\x30\x15\x5c\xc8\x08\x4c\xb3\x15\x45\xe9\x8d\x8d\xbc\x50\x38\xe1\xb4\xbf\x21\xbf\x03\xb3\xbc\x8d\xe0\xa0\xdc\x4e\x97\x08\x9d\x18\x5c\x21\x35\xdb\x68\xa9\x37\xd1\x08\xce\x08\x98\x22\xfb\xd3\xa6\xc9\x11\x2c\xc5\x0b\x74\x1e\x50\x62\xc2\xee\x1b\x4f\x8b\xd3\x4c\x79\x34\xf8\x48\x93\x3b\xa6\x5a\x37\x67\x87\xc6\xc1\x2f\xf1\x7e\xfb\xec\x23\xd1\x24\xcb\x8b\xf1\x70\xd6\x63\x2c\x5e\xc0\x21\x18\x50\x89\x7a\x24\x9a\xb6\x8e\x84\x1a\xb0\x0e\xc5\x19\xce\xa1\x01\x1a\x4c\x60\xd0\xf8\xfa\x56\x01\xf6\x7d\x3d\x01\xfa\xe7\x57\x6a\xdd\x70\x5c\x37\x5a\x8d\xed\x53\xdc\x35\x37\x35\x4e\xad\x0c\xd1\xae\x46\x86\x94\x14\x8a\x2a\x23\x8e\x1b\x05\xd3\x82\x71\x02\x14\xe9\xbe\xad\x3f\xfa\x88\x87\xb2\x62\x58\x66\x4c\x76\x69\x76\x5f\xc3\x3c\x3f\xe8\xc4\x02\xdc\xf6\xa6\xf0\xfd\x45\x7b\xc8\xc9\x76\x58\x8e\x9d\xec\xfb\xb7\x8a\x30\xb9\x42\x3c\x98\x07\xc3\x65\xc0\xa3\x52\x85\xa3\x7b\x2a\x08\x99\xad\xc1\x9b\x96\x3a\xcd\xd2\x74\xb5\x5a\xc5\x60\xdf\x7a\xfa\x40\xab\x30\x80\xfd\x85\x01\x0c\x33\x4d\x51\xb1\x14\x73\x3f\x03\x4a\x46\x08\xa7\xe3\x5d\xd8\x27\xc3\x15\xa4\xde\x3f\x58\xfa\xbf\xcb\xd5\xf9\x59\x67\xa6\x98\x54\xc5\x67\xe7\xdb\x91\x72\x63\x7d\xb6\x73\xc8\x0d\xc0\xde\x9c\xec\x8d\xc9\x47\x6a\x75\x54\x09\x59\xda\x8f\x7e\x5b\x2d\x99\x03\x64\xb1\x1e\x60\x8c\x5a\xc6\x41\x8c\xbe\xcc\x43\x18\xa8\xbb\xeb\x3f\x6b\x3c\x43\x70\xe3\x94\x76\xaa\x70\xdd\x1d\xc1\x11\x14\x90\x6b\x3f\x2c\x61\x6e\x94\xa9\x27\xf9\xa9\x53\xeb\x1c\x4c\x49\x72\x9d\xa2\x4b\x70\x75\xb2\x7d\x36\x7d\x74\x16\xf7\x00\x7c\x29\x6c\x96\x03\x87\x37\x07\xe6\x0c\x81\xae\x4a\xff\xd4\xc2\xda\x93\x80\x8d\x36\x9c\xa2\x05\xba\xda\x60\xcf\x7a\xa6\xe5\xf6\x3b\x98\xc5\x7d\xca\xa6\x40\x77\x49\xfb\x4b\x88\x8b\x37\xc4\xdc\x22\x65\x8c\x6b\xc7\xa8\x96\x22\x67\x24\xfa\xf6\xeb\x7b\xa9\xef\x3a\x3f\x25\xae\x9a\x4c\x8b\x65\xf1\x83\xa5\x52\x34\x22\x53\x8b\x0d\x76\xa3\xb0\x54\x5f\x0d\xb7\x46\xc9\xcf\xc7\x1e\xf1\x78\x0e\x68\x45\xde\x7e\xee\xa7\x07\x04\x12\x6c\x38\x7a\xef\xb9\xdf\x63\x3c\x09\x5c\x8c\xbc\xcf\xb6\xaa\x50\xa0\xaa\x3a\x45\xb3\xb9\x87\x0d\x6a\x6e\xb9\x5b\x45\x97\x07\x55\x51\x22\x38\xd9\xd6\x90\x6f\x0d\xfb\xc5\xd6\xed\x64\xff\x25\xf7\x6f\x49\xee\x90\x4e\xb2\x33\x38\xff\xa6\xd3\x0d\x5e\x10\xc6\xf0\xac\x8a\x5f\x7f\x07\x00\x00\xff\xff\x53\xb2\x95\x60\x2b\x0d\x00\x00")

func webuiStaticStylesCssBytes() ([]byte, error) {
	return bindataRead(
		_webuiStaticStylesCss,
		"webui/static/styles.css",
	)
}

func webuiStaticStylesCss() (*asset, error) {
	bytes, err := webuiStaticStylesCssBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "webui/static/styles.css", size: 3371, mode: os.FileMode(420), modTime: time.Unix(1452106107, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"webui/templates/archived_list.html": webuiTemplatesArchived_listHtml,
	"webui/static/styles.css":            webuiStaticStylesCss,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"webui": &bintree{nil, map[string]*bintree{
		"static": &bintree{nil, map[string]*bintree{
			"styles.css": &bintree{webuiStaticStylesCss, map[string]*bintree{}},
		}},
		"templates": &bintree{nil, map[string]*bintree{
			"archived_list.html": &bintree{webuiTemplatesArchived_listHtml, map[string]*bintree{}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
