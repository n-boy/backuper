package core

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/n-boy/backuper/base"
	"github.com/n-boy/backuper/crypter"
)

func ArchiveNodes(nodes []NodeMetaInfo, archFilePath string, encrypter *crypter.Encrypter) (nodesArch []NodeMetaInfo) {
	archFileWriter, err := os.Create(archFilePath)
	if err != nil {
		base.LogErr.Fatalln(err)
	}
	defer archFileWriter.Close()

	archWriter := io.Writer(archFileWriter)
	if encrypter != nil {
		encrypter.InitWriter(archFileWriter)
		archWriter = io.Writer(encrypter)
	}

	bufWriter := bufio.NewWriterSize(archWriter, 16*1024*1024)
	w := zip.NewWriter(bufWriter)

	for _, node := range nodes {
		fInfo, err := os.Stat(node.path)
		if err != nil {
			base.LogErr.Fatalln(err)
		}
		fHeader, err := zip.FileInfoHeader(fInfo)
		if err != nil {
			base.LogErr.Fatalln(err)
		}
		fHeader.Name = GetPathInArchive(node.path)

		fileWriter, err := w.CreateHeader(fHeader)
		if err != nil {
			base.LogErr.Fatalln(err)
		}

		if !node.is_dir {
			fileReader, err := os.Open(node.path)
			if err != nil {
				base.LogErr.Fatalln(err)
			}
			defer fileReader.Close()

			_, err = io.Copy(fileWriter, fileReader)
			if err != nil {
				base.LogErr.Fatalln(err)
			}

			err = fileReader.Close()
			if err != nil {
				base.LogErr.Fatalln(err)
			}
		}
		node.applyFileInfo(fInfo)
		nodesArch = append(nodesArch, node)
	}

	if err = w.Close(); err != nil {
		base.LogErr.Fatalln(err)
	} else if err = archFileWriter.Close(); err != nil {
		base.LogErr.Fatalln(err)
	}
	return nodesArch
}

func UnarchiveNodes(archFilePath string, nodes []NodeMetaInfo, targetPath string) (nodesUnarch []NodeMetaInfo, err error) {
	// if targetPath == originTargetPath {
	// 	return nodesUnarch, fmt.Errorf("Unarchiving to file origin path is not supported now")
	// }

	zipReader, err := zip.OpenReader(archFilePath)
	if err != nil {
		return nodesUnarch, err
	}
	defer zipReader.Close()

	nodesInArchiveMap := make(map[string]*zip.File)
	for _, f := range zipReader.File {
		nodesInArchiveMap[f.Name] = f
	}

	for _, node := range nodes {
		var targetFilePath string
		if node.is_dir {
			if targetPath == OriginTargetPath {
				targetFilePath = node.GetNodePath()
			} else {
				targetFilePath = filepath.Join(targetPath, GetPathInArchive(node.GetNodePath()))
			}
			err = os.MkdirAll(targetFilePath, 0775)
			if err != nil {
				return nodesUnarch, err
			}
		} else {
			f, exists := nodesInArchiveMap[GetPathInArchive(node.GetNodePath())]
			if !exists {
				return nodesUnarch, fmt.Errorf("File %v is not founded in archive %v", node.GetNodePath(), archFilePath)
			}

			if targetPath == OriginTargetPath {
				targetFilePath = node.GetNodePath()
			} else {
				targetFilePath = filepath.Join(targetPath, f.Name)
			}
			tfi, err := os.Stat(targetFilePath)
			if err == nil {
				if tfi.ModTime().Equal(node.modtime) && tfi.Size() == node.size {
					nodesUnarch = append(nodesUnarch, node)
					continue
				} else {
					return nodesUnarch, fmt.Errorf("File %v already exists and differs from that in archive %v", node.GetNodePath(), archFilePath)
				}
			} else if !os.IsNotExist(err) {
				return nodesUnarch, err
			}

			err = os.MkdirAll(filepath.Dir(targetFilePath), 0775)
			if err != nil {
				return nodesUnarch, err
			}

			fReader, err := f.Open()
			if err != nil {
				return nodesUnarch, err
			}
			defer fReader.Close()

			// os.O_WRONLY|os.O_CREATE|os.O_TRUNC
			tfWriter, err := os.OpenFile(targetFilePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, f.Mode())
			if err != nil {
				return nodesUnarch, err
			}
			defer tfWriter.Close()

			_, err = io.Copy(tfWriter, fReader)
			if err != nil {
				return nodesUnarch, err
			}

			if err = tfWriter.Close(); err != nil {
				return nodesUnarch, err
			} else if err = fReader.Close(); err != nil {
				return nodesUnarch, err
			}
			if err = os.Chtimes(targetFilePath, node.modtime, node.modtime); err != nil {
				base.LogErr.Println(err)
			}
		}
		nodesUnarch = append(nodesUnarch, node)
	}

	if err = zipReader.Close(); err != nil {
		base.LogErr.Println(err)
	}

	return nodesUnarch, nil
}

func GetPathInArchive(path string) string {
	path = regexp.MustCompile(`^([A-Za-z]):`).ReplaceAllString(path, "$1")
	path = regexp.MustCompile(`^[/\\]+`).ReplaceAllString(path, "")
	return path
}
