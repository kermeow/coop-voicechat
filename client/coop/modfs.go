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

type ModFs struct {
	path  string
	files map[string]*ModFsFile
}

type ModFsProperties struct {
	Files map[string]struct {
		IsPublic bool `json:"isPublic"`
		IsText   bool `json:"isText"`
	} `json:"files"`
	IsPublic bool `json:"isPublic"`
}

func ModFsGet(modPath string) (*ModFs, error) {
	m := &ModFs{
		path:  path.Join(SavDir, modPath+".modfs"),
		files: make(map[string]*ModFsFile),
	}
	return m, nil
}

func (m *ModFs) Get(fileName string) (*ModFsFile, error) {
	if _, ok := m.files[fileName]; !ok {
		return nil, os.ErrNotExist
	}
	return m.files[fileName], nil
}

func (m *ModFs) Create(fileName string) *ModFsFile {
	f := &ModFsFile{
		Data:   make([]byte, 0),
		Cursor: 0,
	}
	m.files[fileName] = f
	return f
}

func (m *ModFs) Read(onlyCheckExists bool) (bool, error) {
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
		m.files[v.Name] = &ModFsFile{
			Data:   data,
			Cursor: 0,
		}
	}

	return true, nil
}

func (m *ModFs) Write() (bool, error) {
	file, err := os.OpenFile(m.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return false, err
	}
	defer file.Close()

	properties := ModFsProperties{
		Files: make(map[string]struct {
			IsPublic bool `json:"isPublic"`
			IsText   bool `json:"isText"`
		}),
		IsPublic: true,
	}

	buf := new(bytes.Buffer)
	z := zip.NewWriter(buf)

	for name, v := range m.files {
		properties.Files[name] = struct {
			IsPublic bool `json:"isPublic"`
			IsText   bool `json:"isText"`
		}{
			IsPublic: true,
			IsText:   false,
		}

		f, err := z.CreateHeader(&zip.FileHeader{
			Name:   name,
			Method: zip.Store,
		})
		if err != nil {
			return false, err
		}
		_, err = f.Write(v.Data)
		if err != nil {
			return false, err
		}
	}

	f, err := z.CreateHeader(&zip.FileHeader{
		Name:   "properties.json",
		Method: zip.Store,
	})
	if err != nil {
		return false, err
	}
	err = json.NewEncoder(f).Encode(properties)
	if err != nil {
		return false, err
	}

	z.Close()
	file.Write(buf.Bytes())

	return true, nil
}

type ModFsFile struct {
	Data   []byte
	Cursor int
}

// unused modfs rw functions: int8, int16, int32, int64, string, line

func (f *ModFsFile) ReadBytes(l int) ([]byte, error) {
	if f.Cursor+l > len(f.Data) {
		return nil, io.ErrShortBuffer
	}
	data := f.Data[f.Cursor : f.Cursor+l]
	f.Cursor += l
	return data, nil
}

func (f *ModFsFile) ReadUint8() (uint8, error) {
	data, err := f.ReadBytes(1)
	if err != nil {
		return 0, err
	}
	return data[0], nil
}

func (f *ModFsFile) ReadUint16() (uint16, error) {
	data, err := f.ReadBytes(2)
	if err != nil {
		return 0, err
	}
	return bin.Uint16(data), nil
}

func (f *ModFsFile) ReadUint32() (uint32, error) {
	data, err := f.ReadBytes(4)
	if err != nil {
		return 0, err
	}
	return bin.Uint32(data), nil
}

func (f *ModFsFile) ReadUint64() (uint64, error) {
	data, err := f.ReadBytes(8)
	if err != nil {
		return 0, err
	}
	return bin.Uint64(data), nil
}

func (f *ModFsFile) ReadFloat32() (float32, error) {
	u, err := f.ReadUint32()
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(u), nil
}

func (f *ModFsFile) ReadFloat64() (float64, error) {
	u, err := f.ReadUint64()
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(u), nil
}

func (f *ModFsFile) WriteBytes(d []byte) error {
	l := len(d)
	data := make([]byte, max(len(f.Data), f.Cursor+l))
	copy(data, f.Data)
	copy(data[f.Cursor:f.Cursor+l], d)
	f.Data = data
	f.Cursor += l
	return nil
}

func (f *ModFsFile) WriteUint8(n uint8) error {
	data := make([]byte, 1)
	data[0] = n
	return f.WriteBytes(data)
}

func (f *ModFsFile) WriteUint16(n uint16) error {
	data := make([]byte, 2)
	bin.PutUint16(data, n)
	return f.WriteBytes(data)
}

func (f *ModFsFile) WriteUint32(n uint32) error {
	data := make([]byte, 4)
	bin.PutUint32(data, n)
	return f.WriteBytes(data)
}

func (f *ModFsFile) WriteUint64(n uint64) error {
	data := make([]byte, 8)
	bin.PutUint64(data, n)
	return f.WriteBytes(data)
}

func (f *ModFsFile) WriteFloat32(n float32) error {
	return f.WriteUint32(math.Float32bits(n))
}

func (f *ModFsFile) WriteFloat64(n float64) error {
	return f.WriteUint64(math.Float64bits(n))
}
