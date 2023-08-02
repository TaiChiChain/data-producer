package repo

import (
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
)

const (
	DefaultDirPath  = "~/.data_producer/"
	DefaultAccounts = DefaultDirPath + "accounts"
)

func DirPath() (string, error) {
	return homedir.Expand(DefaultDirPath)
}

func AccountsPath() (string, error) {
	return homedir.Expand(DefaultAccounts)
}

func LoadAccounts(path string) ([]string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	accounts := strings.Split(string(bytes), "\n")
	return accounts[:len(accounts)-1], nil
}
