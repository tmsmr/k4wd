package envfile

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/tmsmr/k4wd/internal/pkg/forwarder"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	envFilePerm = 0644
)

type Envfile struct {
	path string
}

func New() (*Envfile, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	hash := sha256.New()
	hash.Write([]byte(cwd))
	return &Envfile{
		path: filepath.Join(os.TempDir(), "k4wd_env_"+fmt.Sprintf("%x", hash.Sum(nil))),
	}, nil
}

func (ef Envfile) Path() string {
	return ef.path
}

func (ef Envfile) Exists() (bool, error) {
	if _, err := os.Stat(ef.path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func (ef Envfile) Update(forwards map[string]*forwarder.Forwarder) error {
	var content bytes.Buffer
	for _, fwd := range forwards {
		if !fwd.Active {
			continue
		}
		re := regexp.MustCompile(`\W`)
		pfName := strings.ToUpper(re.ReplaceAllString(fwd.Name, "_"))
		envName := fmt.Sprintf("%s_ADDR", pfName)
		content.WriteString(fmt.Sprintf("# %s\n%s=%s:%d\n\n", fwd.Name, envName, fwd.BindAddr, fwd.BindPort))
	}
	return os.WriteFile(ef.path, content.Bytes(), envFilePerm)
}

func (ef Envfile) Remove() error {
	if ef.path == "" {
		panic("ef.path is empty")
	}
	return os.Remove(ef.path)
}

func (ef Envfile) Load() ([]byte, error) {
	exists, err := ef.Exists()
	if err != nil {
		return []byte{}, err
	}
	if !exists {
		return []byte{}, fmt.Errorf("no env available")
	}
	return os.ReadFile(ef.path)
}

func (ef Envfile) Copy(to string) error {
	exists, err := ef.Exists()
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no env available")
	}
	_, err = os.Stat(path.Dir(to))
	if err != nil {
		return fmt.Errorf("missing parent directory %s", path.Dir(to))
	}
	if stat, err := os.Stat(to); err == nil {
		if stat.IsDir() {
			return fmt.Errorf("target is a directory %s", to)
		}
	}
	content, err := os.ReadFile(ef.path)
	if err != nil {
		return err
	}
	return os.WriteFile(to, content, envFilePerm)
}
