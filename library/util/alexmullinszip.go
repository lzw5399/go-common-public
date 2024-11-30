package util

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexmullins/zip"
)

func UnZip(fileBytes []byte, decompressPath string) error {
	err := os.MkdirAll(decompressPath, os.ModePerm)
	if err != nil {
		return err
	}
	r, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		return err
	}

	for _, f := range r.File {
		r, err := f.Open()
		if err != nil {
			return err
		}

		buf, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		defer r.Close()
		if strings.Contains(f.Name, "DS_Store") {
			continue
		}

		if strings.Contains(f.Name, "__MACOSX") {
			continue
		}

		if f.FileInfo().IsDir() {
			// Make Folder
			err = os.MkdirAll(decompressPath+f.Name, os.ModePerm)
		} else {
			err = ioutil.WriteFile(decompressPath+f.Name, buf, os.ModePerm)
			if err != nil {
				if os.IsNotExist(err) {
					dir := filepath.Dir(decompressPath + f.Name)
					err = os.MkdirAll(dir, os.ModePerm)
					if err != nil {
						fmt.Errorf("UnZip mkdir dir:%s err:%s\n", dir, err.Error())
						continue
					}
					err = ioutil.WriteFile(decompressPath+f.Name, buf, os.ModePerm)
					if err != nil {
						fmt.Errorf("UnZip ioutil.WriteFile err:%s\n", err.Error())
					}
				} else {
					fmt.Errorf("UnZip ioutil.WriteFile err:%s\n", err.Error())
				}
			}
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func Compress(files []*os.File, password, dest, uuid string) error {
	d, _ := os.Create(dest)
	defer d.Close()
	w := zip.NewWriter(d)
	defer w.Close()
	for _, file := range files {
		err := compress(file, "", password, uuid, w)
		if err != nil {
			return err
		}
	}
	return nil
}

func compress(file *os.File, prefix, password, uuid string, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		name := strings.Split(info.Name(), uuid)[0]
		if prefix == "" {
			prefix = name
		} else {
			prefix = prefix + "/" + name
		}
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = compress(f, prefix, password, uuid, zw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		if prefix != "" {
			header.Name = prefix + "/" + header.Name
		}
		header.SetPassword(password)
		header.Method = zip.Deflate
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
