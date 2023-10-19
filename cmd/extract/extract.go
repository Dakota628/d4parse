package main

import (
	"bufio"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/Dakota628/d4parse/pkg/d4/util"
	"golang.org/x/exp/slices"
	"log"
	"log/slog"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	cascLocale  = "enUS"
	cascProduct = "fenris"
)

func cascConsole(cascConsoleExe string, args ...string) (*exec.Cmd, error) {
	// Setup path and args
	var path string
	switch runtime.GOOS {
	case "windows":
		path = cascConsoleExe
	default:
		path = "dotnet"
		slices.Insert(args, 0, cascConsoleExe)
	}

	// Create temporary working directory (to allow for async execution)
	workDir, err := os.MkdirTemp(os.TempDir(), "CASCConsole-")
	if err != nil {
		return nil, err
	}

	name := filepath.Base(workDir)
	log.New(os.Stdout, name, 0)

	cmd := exec.Command(path, args...)
	cmd.Dir = workDir
	cmd.Stdout = log.New(os.Stdout, name, 0).Writer()
	cmd.Stderr = log.New(os.Stderr, name, 0).Writer()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Start goroutine to clean temp working directory once command finishes
	go func() {
		_ = cmd.Wait()
		_ = os.RemoveAll(workDir)
	}()

	return cmd, nil
}

func downloadByPattern(cascConsoleExe string, dumpPath string, pattern string) (*exec.Cmd, error) {
	return cascConsole(
		cascConsoleExe,
		"-m", "Pattern",
		"-e", pattern,
		"-d", dumpPath,
		"-l", cascLocale,
		"-p", cascProduct,
		"-o",
	)
}

func downloadByList(cascConsoleExe string, dumpPath string, fileList []string) (*exec.Cmd, error) {
	// Create temp list file
	dir := os.TempDir()
	f, err := os.CreateTemp(dir, "cc-listfile-")
	if err != nil {
		return nil, err
	}

	path, err := filepath.Abs(filepath.Join(dir, f.Name()))
	if err != nil {
		return nil, err
	}

	// Write to list file
	w := bufio.NewWriter(f)
	for _, file := range fileList {
		if _, err := w.WriteString(file); err != nil {
			_ = os.Remove(path)
			return nil, err
		}
		if _, err := w.WriteRune('\n'); err != nil {
			_ = os.Remove(path)
			return nil, err
		}
	}

	if err := w.Flush(); err != nil {
		_ = os.Remove(path)
		return nil, err
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return nil, err
	}

	// Execute casc console
	cmd, err := cascConsole(
		cascConsoleExe,
		"-m", "Listfile",
		"-e", path,
		"-d", dumpPath,
		"-l", cascLocale,
		"-p", cascProduct,
		"-o",
	)
	if err != nil {
		_ = os.Remove(path)
		return nil, err
	}

	// Start goroutine to clean temp file once command finishes
	go func() {
		_ = cmd.Wait()
		_ = os.Remove(path)
	}()

	return cmd, err
}

func generateFileLists(dumpPath string, chunks uint) ([][]string, error) {
	// Load the TOC
	tocFilePath := filepath.Join(dumpPath, "base", "CoreTOC.dat")
	toc, err := d4.ReadTocFile(tocFilePath)
	if err != nil {
		return nil, err
	}

	// Generate a master list of all files
	fileTypes := []util.FileType{
		util.FileTypeMeta,
		util.FileTypePayload,
		util.FileTypePaylow,
	}

	var possibleFiles []string

	for snoGroup, snoNameMap := range toc.Entries {
		for _, snoName := range snoNameMap {
			snoName = strings.Trim(snoName, "\t\n\v\f\r\x00")
			if len(snoName) == 0 {
				continue
			}

			for _, fileType := range fileTypes {
				path := util.BaseFilePath("", fileType, snoGroup, snoName)
				path = strings.ReplaceAll(path, "\\", "/")
				possibleFiles = append(possibleFiles, path)
			}
		}
	}

	// Shuffle the list before chunking it
	rand.Shuffle(len(possibleFiles), func(i, j int) {
		possibleFiles[i], possibleFiles[j] = possibleFiles[j], possibleFiles[i]
	})

	// Chunk the list via bin packing
	fileLists := make([][]string, chunks)

	var chunk uint
	for _, file := range possibleFiles {
		fileLists[chunk] = append(fileLists[chunk], file)
		chunk = (chunk + 1) % chunks
	}

	return fileLists, nil
}

func main() {
	if len(os.Args) < 3 {
		println("usage: extract cascConsoleExe dumpPath")
		os.Exit(1)
	}

	cascConsoleExe := os.Args[1]
	dumpPath := os.Args[2]

	slog.Debug(
		"Program arguments",
		slog.Any("cascConsoleExe", cascConsoleExe),
		slog.Any("dumpPath", dumpPath),
	)

	// Get dat files first so we can construct list files
	//cmd, err := downloadByPattern(cascConsoleExe, dumpPath, "base/*.dat")
	//if err != nil {
	//	slog.Error("Failed to download dat files", slog.Any("error", err))
	//	os.Exit(1)
	//}
	//
	//if err := cmd.Wait(); err != nil {
	//	slog.Error(
	//		"Failed while waiting for CASCConsole to download dat files",
	//		slog.Any("error", err),
	//	)
	//	os.Exit(1)
	//}

	fileLists, err := generateFileLists(dumpPath, uint(runtime.NumCPU()))
	if err != nil {
		slog.Error("Failed to generate file lists", slog.Any("error", err))
		os.Exit(1)
	}

	// Download each file list in a different process
	var cmds []*exec.Cmd

	for i, fileList := range fileLists {
		cmd, err := downloadByList(cascConsoleExe, dumpPath, fileList)
		if err != nil {
			slog.Error("Failed to download list file", slog.Any("index", i))
			os.Exit(1)
		}
		cmds = append(cmds, cmd)
	}

	// Wait for each process to finish
	for i, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			slog.Error("Failed while for CASCConsole to download list file", slog.Any("index", i))
			os.Exit(1)
		}
	}
}
