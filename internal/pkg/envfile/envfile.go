package envfile

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tmsmr/k4wd/internal/pkg/forwarder"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type EnvFormat int

const (
	FormatJSON EnvFormat = iota
	FormatDefault
	FormatNoExport
	FormatPS
	FormatCmd
)

const (
	filePrefix = "k4wd_env_"
	filePerm   = 0644
	envSuffix  = "ADDR"
)

type envEntry struct {
	Addr  string `json:"addr"`
	Value string `json:"value"`
}

type Envfile struct {
	mu   sync.Mutex
	path string
}

func New(ref string) (*Envfile, error) {
	abs, err := filepath.Abs(ref)
	if err != nil {
		return nil, err
	}
	hash := sha256.New()
	hash.Write([]byte(abs))
	return &Envfile{
		path: filepath.Join(os.TempDir(), fmt.Sprintf("%s%x", filePrefix, hash.Sum(nil))),
	}, nil
}

func (ef *Envfile) exists() (bool, error) {
	if _, err := os.Stat(ef.path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func (ef *Envfile) Path() string {
	return ef.path
}

func (ef *Envfile) Update(forwards map[string]*forwarder.Forwarder) error {
	ef.mu.Lock()
	defer ef.mu.Unlock()
	re := regexp.MustCompile(`\W`)
	addrs := make([]envEntry, 0)
	for _, fwd := range forwards {
		addrs = append(addrs, envEntry{
			Addr:  fmt.Sprintf("%s_%s", strings.ToUpper(re.ReplaceAllString(fwd.Name, "_")), envSuffix),
			Value: fmt.Sprintf("%s:%d", fwd.BindAddr, fwd.BindPort),
		})
	}
	data, err := json.MarshalIndent(addrs, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(ef.path, data, filePerm)
}

func (ef *Envfile) Remove() error {
	if ef.path == "" {
		panic("ef.path is empty")
	}
	return os.Remove(ef.path)
}

func (ef *Envfile) Load(f EnvFormat) ([]byte, error) {
	exists, err := ef.exists()
	if err != nil {
		return []byte{}, err
	}
	if !exists {
		return []byte{}, fmt.Errorf("no env available")
	}
	data, err := os.ReadFile(ef.path)
	if f == FormatJSON {
		return data, err
	}
	if err != nil {
		return []byte{}, err
	}
	var addrs []envEntry
	if err := json.Unmarshal(data, &addrs); err != nil {
		return []byte{}, err
	}
	var content bytes.Buffer
	for _, addr := range addrs {
		switch f {
		case FormatDefault:
			content.WriteString(fmt.Sprintf("export %s=%s\n", addr.Addr, addr.Value))
			break
		case FormatNoExport:
			content.WriteString(fmt.Sprintf("%s=%s\n", addr.Addr, addr.Value))
			break
		case FormatPS:
			content.WriteString(fmt.Sprintf("$Env:%s=\"%s\"\n", addr.Addr, addr.Value))
			break
		case FormatCmd:
			content.WriteString(fmt.Sprintf("set %s=%s\n", addr.Addr, addr.Value))
			break
		default:
			panic("unknown format")
		}
	}
	return content.Bytes(), nil
}
