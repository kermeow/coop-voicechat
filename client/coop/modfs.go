package coop

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path"
)

type ModFS struct {
	path  string
	files map[string]*ModFSFile
}

type ModFSFile struct {
	Data []byte
}

type ModFSProperties struct {
	Files map[string]struct {
		IsPublic bool `json:"isPublic"`
		IsText   bool `json:"isText"`
	} `json:"files"`
	IsPublic bool `json:"isPublic"`
}

func ModFSGet(modPath string) (*ModFS, error) {
	m := &ModFS{
		path:  path.Join(Sav, modPath+".modfs"),
		files: make(map[string]*ModFSFile),
	}
	return m, nil
}

func (m *ModFS) Get(fileName string) (*ModFSFile, error) {
	if _, ok := m.files[fileName]; !ok {
		return nil, os.ErrNotExist
	}
	return m.files[fileName], nil
}

func (m *ModFS) Create(fileName string) *ModFSFile {
	f := &ModFSFile{
		Data: make([]byte, 0),
	}
	m.files[fileName] = f
	return f
}

func (m *ModFS) Read(onlyCheckExists bool) (bool, error) {
	zip, err := zip.OpenReader(m.path)
	if err != nil {
		return false, err
	}
	defer zip.Close()

	if onlyCheckExists {
		return true, nil
	}

	for _, v := range zip.File {
		if path.Base(v.Name) == "properties.json" {
			// I DON'T FUCKING CARE!!!
			continue
		}

		file, err := v.Open()
		if err != nil {
			return false, err
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			return false, err
		}
		m.files[v.Name] = &ModFSFile{
			Data: data,
		}
	}

	return true, nil
}

func (m *ModFS) Write() (bool, error) {
	file, err := os.OpenFile(m.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return false, err
	}
	defer file.Close()

	properties := ModFSProperties{
		Files: make(map[string]struct {
			IsPublic bool `json:"isPublic"`
			IsText   bool `json:"isText"`
		}),
		IsPublic: true,
	}

	buf := new(bytes.Buffer)
	zip := zip.NewWriter(buf)

	for name, v := range m.files {
		properties.Files[name] = struct {
			IsPublic bool `json:"isPublic"`
			IsText   bool `json:"isText"`
		}{
			IsPublic: true,
			IsText:   false,
		}

		f, err := zip.Create(name)
		if err != nil {
			return false, err
		}
		_, err = f.Write(v.Data)
		if err != nil {
			return false, err
		}
	}

	f, err := zip.Create("properties.json")
	if err != nil {
		return false, err
	}
	err = json.NewEncoder(f).Encode(properties)
	if err != nil {
		return false, err
	}

	zip.Close()
	file.Write(buf.Bytes())

	return true, nil
}

func (f *ModFSFile) Buf() *bytes.Buffer {
	data := make([]byte, len(f.Data))
	copy(data, f.Data)
	return bytes.NewBuffer(data)
}
