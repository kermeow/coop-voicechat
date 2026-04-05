package coop

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"path"
)

var bin binary.ByteOrder = binary.LittleEndian

type ModFS struct {
	path  string
	files map[string]*ModFSFile
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
		Data:   make([]byte, 0),
		Cursor: 0,
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
			Data:   data,
			Cursor: 0,
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

type ModFSFile struct {
	Data   []byte
	Cursor int
}

// unused modfs rw functions: uint8, uint16, uint64, int8, int16, int32, int64, string, line

func (f *ModFSFile) ReadBytes(l int) ([]byte, error) {
	if f.Cursor+l > len(f.Data) {
		return nil, io.ErrShortBuffer
	}
	data := f.Data[f.Cursor : f.Cursor+l]
	f.Cursor += l
	return data, nil
}

func (f *ModFSFile) ReadUint32() (uint32, error) {
	data, err := f.ReadBytes(4)
	if err != nil {
		return 0, err
	}
	return bin.Uint32(data), nil
}

func (f *ModFSFile) WriteBytes(d []byte) error {
	l := len(d)
	data := make([]byte, max(len(f.Data), f.Cursor+l))
	copy(data, f.Data)
	copy(data[f.Cursor:f.Cursor+l], d)
	f.Data = data
	f.Cursor += l
	return nil
}

func (f *ModFSFile) WriteUint32(n uint32) error {
	data := make([]byte, 4)
	bin.PutUint32(data, n)
	return f.WriteBytes(data)
}
