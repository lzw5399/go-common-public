package util

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/lzw5399/go-common-public/library/log"
)

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func CreatePath(filePath string) error {
	if !IsExist(filePath) {
		err := os.MkdirAll(filePath, 0755)
		return err
	} else {
		return RemoveContents(filePath)
	}
}

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func GetFileList(root string) []string {
	allList := []string{}
	filepath.Walk(root,
		func(path string, f os.FileInfo, err error) error {
			if f == nil {
				return err
			}
			if f.IsDir() {
				return nil
			} else {
				//log.Infoln("walk file:", path)
				allList = append(allList, path)
			}
			return nil
		})
	return allList
}

type Unzip struct {
	Src  string
	Dest string
}

func NewUnzip(src string, dest string) Unzip {
	return Unzip{src, dest}
}

func (uz Unzip) Extract() error {
	log.Infof("Extraction of " + uz.Src + " started!")
	r, err := zip.OpenReader(uz.Src)
	if err != nil {
		return err
	}
	defer r.Close()
	os.MkdirAll(uz.Dest, 0755)
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		path := filepath.Join(uz.Dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}
	for _, f := range r.File {
		if strings.Contains(f.Name, "__MACOSX") {
			continue
		}
		err := extractAndWriteFile(f)
		if err != nil {
			log.Warnf("extractAndWriteFile file:%s err:%s", f.Name, err.Error())
		}
	}
	log.Infof("Extraction of " + uz.Src + " finished!")
	return nil
}

func DeCompress(tarFile, dest string) ([]string, []string, error) {
	dirList := make([]string, 0)
	fileList := make([]string, 0)
	if strings.HasSuffix(tarFile, ".zip") {
		return zipDeCompress(tarFile, dest)
	}
	return dirList, fileList, errors.New("file format error")
}

func zipDeCompress(zipFile, dest string) ([]string, []string, error) {
	dirList := make([]string, 0)
	fileList := make([]string, 0)
	or, err := zip.OpenReader(zipFile)
	defer or.Close()
	if err != nil {
		return dirList, dirList, err
	}

	for _, item := range or.File {
		if item.FileInfo().IsDir() {
			dirList = append(dirList, item.Name)
			continue
		}
		fileList = append(fileList, item.Name)
	}

	return dirList, fileList, nil
}

func ZipReadDesFile(path, dest string) ([]byte, bool, error) {
	or, err := zip.OpenReader(path)
	defer or.Close()
	if err != nil {
		return nil, false, err
	}
	for _, item := range or.File {
		if !item.FileInfo().IsDir() && item.FileInfo().Name() == dest {
			t, err := item.Open()
			if err != nil {
				return nil, false, err
			}
			b, err := ioutil.ReadAll(t)
			if err != nil {
				return nil, false, err
			}
			t.Close()
			return b, true, nil
		}
	}
	return nil, false, nil
}

func ZipList(fileList []string, destinationPath string) error {
	fmt.Println(fileList)
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	myZip := zip.NewWriter(destinationFile)
	for _, fileName := range fileList {
		f, err := os.Open(fileName)
		if err != nil {
			return err
		}
		fInfo, err := f.Stat()
		if err != nil {
			return err
		}
		if fInfo.IsDir() {
			return fmt.Errorf("file:[%s] is dir", fileName)
		}
		_, name := filepath.Split(fileName)
		zipFile, err := myZip.Create(name)
		if err != nil {
			return err
		}
		fsFile, err := os.Open(fileName)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipFile, fsFile)
		if err != nil {
			return err
		}
	}
	err = myZip.Close()
	if err != nil {
		return err
	}
	return nil
}
