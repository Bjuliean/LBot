package files

import (
	"bot/LBot/src/internal/storage"
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type Storage struct {
	basePath	string
}

const defaultPerm = 0774

var ErrNoSavedPages = errors.New("no saved pages")

func New(basePath string) Storage {
	return Storage{basePath: basePath}
}

func (s *Storage)Save(page *storage.Page) error {
	const ferr = "storage.files.Save"

	filePath := filepath.Join(s.basePath, page.UserName)

	if err := os.MkdirAll(filePath, defaultPerm); err != nil {
		return fmt.Errorf("%s: %w", ferr, err)
	}

	filename, err := fileName(page)
	if err != nil {
		return fmt.Errorf("%s: %w", ferr, err)
	}

	filePath = filepath.Join(filePath, filename)

	file, err := os.Create(filePath)
	defer func(){_=file.Close()}()

	if err := gob.NewEncoder(file).Encode(page); err != nil {
		return fmt.Errorf("%s: %w", ferr, err)
	}
	return nil
}

func (s *Storage)PickRandom(userName string) (*storage.Page, error) {
	const ferr = "storage.files.PickRandom"

	path := filepath.Join(s.basePath, userName)

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ferr, err)
	}
	if len(files) == 0 {
		return nil, ErrNoSavedPages
	}

	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(len(files))

	file := files[n]

	return s.decodePage(filepath.Join(path, file.Name()))
}

func (s *Storage)decodePage(filePath string) (*storage.Page, error) {
	const ferr = "storage.files.decodePage"

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ferr, err)
	}
	defer func(){_=f.Close()}()
	
	var p storage.Page

	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		return nil, fmt.Errorf("%s: %w", ferr, err)
	}
	return &p, nil
}

func (s *Storage)Remove(p *storage.Page) error {
	const ferr = "storage.files.Remove"

	fName, err := fileName(p)
	if err != nil {
		return fmt.Errorf("%s: %w", ferr, err)
	}

	path := filepath.Join(s.basePath, p.UserName, fName)

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("%s: %w", ferr, err)
	}
	return nil
}

func (s *Storage)IsExists(p *storage.Page) (bool, error) {
	const ferr = "storage.files.IsExists"
	
	fName, err := fileName(p)
	if err != nil {
		return false, fmt.Errorf("%s: %w", ferr, err)
	}

	path := filepath.Join(s.basePath, p.UserName, fName)

	if _, err := os.Stat(path); err != nil {
		return false, fmt.Errorf("%s: %w", ferr, err)
	}
	return true, nil
}

func fileName(p *storage.Page) (string, error) {
	return p.Hash()
}