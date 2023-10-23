package main

import (
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/Dakota628/d4parse/pkg/ml"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
	"os"
	"strconv"
)

func parseSnoGroup(s string) d4.SnoGroup {
	snoGroupInt, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return d4.SnoGroup(snoGroupInt)
}

func parseSnoId(s string) int32 {
	snoId, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return int32(snoId)
}

func main() {
	if len(os.Args) < 4 {
		slog.Error("usage: freqs dataPath snoGroup snoId")
		os.Exit(1)
	}

	dataPath := os.Args[1]
	snoGroupIntStr := parseSnoGroup(os.Args[2])
	snoId := parseSnoId(os.Args[3])

	freqs, err := ml.DetermineSnoValueFreqs(dataPath, snoGroupIntStr, snoId)
	if err != nil {
		panic(err)
	}

	output := make([]string, 0, len(freqs))
	for k, v := range freqs {
		output = append(output, fmt.Sprintf("%f %s", v, k))
	}
	slices.Sort(output)

	for _, line := range output {
		println(line)
	}

	//jsonData, err := json.Marshal(freqs)
	//if err != nil {
	//	panic(err)
	//}
	//print(string(jsonData))
}
