package d4

import (
	"os"
	"path/filepath"
)

var (
	basePath = "base"
	metaPath = filepath.Join(basePath, "meta")
)

func GroupMetaDir(dataPath string, group SnoGroup) string {
	return filepath.Join(dataPath, metaPath, group.String())
}

func EachSnoMeta(dataPath string, group SnoGroup, cb func(meta SnoMeta) bool) error {
	groupMetaDir := GroupMetaDir(dataPath, group)
	entries, err := os.ReadDir(groupMetaDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		metaFilePath := filepath.Join(groupMetaDir, entry.Name())
		meta, err := ReadSnoMetaFile(metaFilePath)
		if err != nil {
			return err
		}

		if !cb(meta) {
			break
		}
	}

	return nil
}
