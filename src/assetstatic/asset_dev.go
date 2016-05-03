// +build !embed

// Automagically generated by yaber v0.2 (https://github.com/lmas/yaber),
// please avoid editting this file as it might be regenerated again.

package assetstatic

import (
	"io/ioutil"
	"path/filepath"
)

func Asset(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func AssetDir(dir string) (map[string][]byte, error) {
	list := make(map[string][]byte)
	dirs := []string{dir}

	for len(dirs) > 0 {
		d := dirs[0]
		dirs = dirs[1:]
		files, e := ioutil.ReadDir(d)
		if e != nil {
			return nil, e
		}

		for _, f := range files {
			fpath := filepath.Join(d, f.Name())

			if f.IsDir() {
				dirs = append(dirs, fpath)
				continue
			}
			if !f.Mode().IsRegular() {
				continue
			}

			fbody, e := ioutil.ReadFile(fpath)
			if e != nil {
				return nil, e
			}
			list[fpath] = fbody
		}
	}
	return list, nil
}
