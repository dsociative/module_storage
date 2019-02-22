package store

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"time"
)

type Store struct {
	path         string
	metadataFile *os.File
}

type Meta struct {
	Name        string
	Description string
	Package     string
}

type ModuleMetadata struct {
	Versions      map[int]time.Time
	VersionCount  int
	ActiveVersion int
	Meta          Meta
}

func NewModuleMetadata() ModuleMetadata {
	return ModuleMetadata{Versions: map[int]time.Time{}, VersionCount: 0, Meta: Meta{}}
}

type MapModuleMetadata map[string]ModuleMetadata

type StoreMetadata struct {
	Modules MapModuleMetadata
}

func NewStore(p string) *Store {
	metadataPath := path.Join(p, "metadata.json")
	metadataFile, err := os.OpenFile(metadataPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("can't open metadata.json %s", err)
	}
	if stat, err := metadataFile.Stat(); err == nil {
		if stat.Size() == 0 {
			json.NewEncoder(metadataFile).Encode(MapModuleMetadata{})
		}
	} else {
		log.Fatalf("can't stat metadata.json %s", err)
	}
	return &Store{path: p, metadataFile: metadataFile}
}

func (s *Store) Modules() (m MapModuleMetadata, err error) {
	var storeMetadata StoreMetadata
	s.metadataFile.Seek(0, 0)
	err = json.NewDecoder(s.metadataFile).Decode(&storeMetadata)
	return storeMetadata.Modules, err
}

func (s *Store) MustModules() MapModuleMetadata {
	m, err := s.Modules()
	if err != nil {
		log.Fatal(err)
	}
	return m
}

func (s *Store) storeModify(f func(*StoreMetadata) error) (err error) {
	var storeMetadata StoreMetadata
	s.metadataFile.Seek(0, 0)
	if err = json.NewDecoder(s.metadataFile).Decode(&storeMetadata); err != nil {
		return err
	}
	if err = f(&storeMetadata); err != nil {
		return err
	}
	s.metadataFile.Seek(0, 0)
	s.metadataFile.Truncate(0)
	return json.NewEncoder(s.metadataFile).Encode(&storeMetadata)
}

func (s *Store) NewModule(id string) error {
	return s.storeModify(func(storeMetadata *StoreMetadata) error {
		if storeMetadata.Modules == nil {
			storeMetadata.Modules = MapModuleMetadata{}
		}
		storeMetadata.Modules[id] = NewModuleMetadata()
		return nil
	})
}

func (s *Store) openVersionFile(id string, version int) (*os.File, error) {
	return os.OpenFile(path.Join(s.path, id+"_"+strconv.Itoa(version)), os.O_CREATE|os.O_RDWR, 0644)
}

func (s *Store) AddModuleVersion(id string, t time.Time, r io.Reader) error {
	return s.storeModify(func(m *StoreMetadata) error {
		module := m.Modules[id]
		module.VersionCount++
		module.Versions[module.VersionCount] = t

		m.Modules[id] = module
		versionFile, err := s.openVersionFile(id, module.VersionCount)
		defer versionFile.Close()
		if err == nil {
			_, err := io.Copy(versionFile, r)
			return err
		}
		return err
	})
}

func (s *Store) ActiveVersion(id string) (m ModuleMetadata, r io.Reader, err error) {
	var modules MapModuleMetadata
	modules, err = s.Modules()
	if err == nil {
		var ok bool
		if m, ok = modules[id]; ok {
			r, err = s.openVersionFile(id, m.ActiveVersion)
			return
		}
	}
	return m, r, err
}

func (s *Store) SetModuleVersion(id string, version int) error {
	return s.storeModify(func(m *StoreMetadata) error {
		module := m.Modules[id]
		module.ActiveVersion = version
		m.Modules[id] = module
		return nil
	})
}

func (s *Store) SetModuleMeta(id, n, p, d string) error {
	return s.storeModify(func(m *StoreMetadata) error {
		module := m.Modules[id]
		module.Meta = Meta{Name: n, Package: p, Description: d}
		m.Modules[id] = module
		return nil
	})
}
