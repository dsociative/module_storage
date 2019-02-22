package store

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testAddModule(t *testing.T, s *Store) {
	for _, tt := range []struct {
		id     string
		result MapModuleMetadata
	}{
		{"", MapModuleMetadata(nil)},
		{"proxy", MapModuleMetadata{"proxy": NewModuleMetadata()}},
		{"ads", MapModuleMetadata{
			"proxy": NewModuleMetadata(),
			"ads":   NewModuleMetadata(),
		}},
	} {
		t.Run(fmt.Sprintf("Store New Module %s", tt.id), func(t *testing.T) {
			if tt.id != "" {
				s.NewModule(tt.id)
			}
			m, err := s.Modules()
			assert.NoError(t, err)
			assert.Equal(t, tt.result, m)
		})
	}
}

func testAddVersion(t *testing.T, s *Store) {
	now := time.Date(2016, time.August, 15, 0, 0, 0, 0, time.UTC)
	b := bytes.NewBufferString("someblobdata")
	require.NoError(t, s.NewModule("proxy"))
	assert.Equal(t, NewModuleMetadata(), s.MustModules()["proxy"])
	for _, tt := range []struct {
		result ModuleMetadata
	}{
		{ModuleMetadata{VersionCount: 1, Versions: map[int]time.Time{1: now}}},
		{ModuleMetadata{VersionCount: 2, Versions: map[int]time.Time{1: now, 2: now}}},
	} {
		t.Run(fmt.Sprintf("Store New Module %#v", tt.result.VersionCount), func(t *testing.T) {
			assert.NoError(t, s.AddModuleVersion("proxy", now, b), "add module version")
			m, err := s.Modules()
			assert.NoError(t, err)
			assert.Equal(t, tt.result, m["proxy"])
		})
	}

	assert.NoError(t, s.SetModuleVersion("proxy", 2))
	assert.Equal(t, 2, s.MustModules()["proxy"].ActiveVersion)
}

func TestStore(t *testing.T) {
	for _, f := range []func(*testing.T, *Store){testAddModule, testAddVersion} {
		os.MkdirAll("/tmp/module_store_test", 0755)
		dir, _ := ioutil.ReadDir("/tmp/module_store_test")
		for _, d := range dir {
			os.RemoveAll(path.Join([]string{"/tmp/module_store_test", d.Name()}...))
		}
		s := NewStore("/tmp/module_store_test")
		f(t, s)
	}
}
