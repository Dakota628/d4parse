package main

import (
	"encoding/json"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/Dakota628/d4parse/pkg/ml"
	"github.com/mpraski/clusters"
	"golang.org/x/exp/slog"
	"os"
	"path/filepath"
	"strconv"
)

func parseSnoGroup(s string) d4.SnoGroup {
	snoGroupInt, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return d4.SnoGroup(snoGroupInt)
}

func main() {
	if len(os.Args) < 3 {
		slog.Error("usage: ml dataPath snoGroup")
		os.Exit(1)
	}

	dataPath := os.Args[1]
	snoGroupIntStr := parseSnoGroup(os.Args[2])

	// Read toc
	toc, err := d4.ReadTocFile(filepath.Join(dataPath, "base", "CoreTOC.dat"))
	if err != nil {
		panic(err)
	}

	// Determine features
	snoIds, features, err := ml.ExtractGroupFeatures(dataPath, toc, snoGroupIntStr)
	if err != nil {
		panic(err)
	}

	// K-means
	c, e := clusters.KMeans(10000, 100, clusters.EuclideanDistance)
	if e != nil {
		panic(e)
	}

	if e = c.Learn(features); e != nil {
		panic(e)
	}

	groups := make(map[int][]int32)
	for i, cluster := range c.Guesses() {
		groups[cluster] = append(groups[cluster], snoIds[i])
	}

	jsonData, err := json.Marshal(groups)
	if err != nil {
		panic(err)
	}
	print(string(jsonData))
}
