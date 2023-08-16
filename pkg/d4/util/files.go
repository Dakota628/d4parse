package util

import (
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"os"
	"path/filepath"
)

var (
	basePath = "base"
	metaPath = filepath.Join(basePath, "meta")
)

type FileType string

const (
	FileTypeMeta         FileType = "meta"
	FileTypePayload      FileType = "payload"
	FileTypePaylow       FileType = "paylow"
	FileTypeEnUsTextMeta          = FileType(".." +
		string(filepath.Separator) +
		"enUS_Text" +
		string(filepath.Separator) +
		"meta")
)

func BaseFilePattern(
	dataPath string,
	ft FileType,
	group d4.SnoGroup,
	name string,
	prefixPattern string,
	suffixPattern string,
) string {
	return filepath.Join(
		dataPath,
		basePath,
		string(ft),
		group.String(),
		fmt.Sprintf("%s%s%s%s", prefixPattern, name, suffixPattern, group.Ext()),
	)
}

func BaseFilePath(dataPath string, ft FileType, group d4.SnoGroup, name string) string {
	return BaseFilePattern(dataPath, ft, group, name, "", "")
}

func MetaPathByName(dataPath string, group d4.SnoGroup, name string) string {
	return BaseFilePath(dataPath, FileTypeMeta, group, name)
}

func MetaPathById(dataPath string, toc d4.Toc, snoId int32) string {
	group, name := toc.Entries.GetName(snoId)
	return MetaPathByName(dataPath, group, name)
}

func GroupMetaDir(dataPath string, group d4.SnoGroup) string {
	return filepath.Join(dataPath, metaPath, group.String())
}

func ChangePathType(path string, ft FileType) string {
	var file, group string
	path, file = filepath.Split(path)
	path = filepath.Clean(path)
	path, group = filepath.Split(path)
	path = filepath.Clean(path)
	path, _ = filepath.Split(path)
	path = filepath.Clean(path)
	return filepath.Join(path, string(ft), group, file)
}

func EachSnoMeta(dataPath string, group d4.SnoGroup, cb func(meta d4.SnoMeta) bool) error {
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
		meta, err := d4.ReadSnoMetaFile(metaFilePath)
		if err != nil {
			return err
		}

		if !cb(meta) {
			break
		}
	}

	return nil
}

func EachSnoMetaAsync(workers uint, dataPath string, group d4.SnoGroup, cb func(meta d4.SnoMeta), ecb func(err error)) {
	groupMetaDir := GroupMetaDir(dataPath, group)
	entries, err := os.ReadDir(groupMetaDir)
	if err != nil {
		ecb(err)
		return
	}

	DoWorkSlice(workers, entries, func(entry os.DirEntry) {
		if entry.IsDir() {
			return
		}

		metaFilePath := filepath.Join(groupMetaDir, entry.Name())
		meta, err := d4.ReadSnoMetaFile(metaFilePath)
		if err != nil {
			ecb(err)
			return
		}

		cb(meta)
	})

	return
}
