package coop

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"math"
	"os"
	"path"
)

var bin binary.ByteOrder = binary.LittleEndian

// THIS IS AN INCOMPLETE IMPLEMENTATION __BY DESIGN__

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
		path:  path.Join(SavDir, modPath+".modfs"),
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

// unused modfs rw functions: int8, int16, int32, int64, string, line

func (f *ModFSFile) ReadBytes(l int) ([]byte, error) {
	if f.Cursor+l > len(f.Data) {
		return nil, io.ErrShortBuffer
	}
	data := f.Data[f.Cursor : f.Cursor+l]
	f.Cursor += l
	return data, nil
}

func (f *ModFSFile) ReadUint8() (uint8, error) {
	data, err := f.ReadBytes(1)
	if err != nil {
		return 0, err
	}
	return data[0], nil
}

func (f *ModFSFile) ReadUint16() (uint16, error) {
	data, err := f.ReadBytes(2)
	if err != nil {
		return 0, err
	}
	return bin.Uint16(data), nil
}

func (f *ModFSFile) ReadUint32() (uint32, error) {
	data, err := f.ReadBytes(4)
	if err != nil {
		return 0, err
	}
	return bin.Uint32(data), nil
}

func (f *ModFSFile) ReadUint64() (uint64, error) {
	data, err := f.ReadBytes(8)
	if err != nil {
		return 0, err
	}
	return bin.Uint64(data), nil
}

func (f *ModFSFile) ReadFloat32() (float32, error) {
	u, err := f.ReadUint32()
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(u), nil
}

func (f *ModFSFile) ReadFloat64() (float64, error) {
	u, err := f.ReadUint64()
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(u), nil
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

func (f *ModFSFile) WriteUint8(n uint8) error {
	data := make([]byte, 1)
	data[0] = n
	return f.WriteBytes(data)
}

func (f *ModFSFile) WriteUint16(n uint16) error {
	data := make([]byte, 2)
	bin.PutUint16(data, n)
	return f.WriteBytes(data)
}

func (f *ModFSFile) WriteUint32(n uint32) error {
	data := make([]byte, 4)
	bin.PutUint32(data, n)
	return f.WriteBytes(data)
}

func (f *ModFSFile) WriteUint64(n uint64) error {
	data := make([]byte, 8)
	bin.PutUint64(data, n)
	return f.WriteBytes(data)
}

func (f *ModFSFile) WriteFloat32(n float32) error {
	return f.WriteUint32(math.Float32bits(n))
}

func (f *ModFSFile) WriteFloat64(n float64) error {
	return f.WriteUint64(math.Float64bits(n))
}
