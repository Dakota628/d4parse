package d4data

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
)

type Definition struct {
	Type              string  `json:"type"`
	Hash              int     `json:"hash"`
	Name              string  `json:"name"`
	Flags             int     `json:"flags"`
	Inherits          []int   `json:"inherits"`
	FormatHash        int     `json:"dwFormatHash"`
	IsPolymorphicType bool    `json:"isPolymorphicType"`
	Fields            []Field `json:"fields"`
	Size              int     `json:"size"`
}

type Field struct {
	Type                        [3]int `json:"type"`
	Hash                        int    `json:"hash"`
	Name                        string `json:"name"`
	Flags                       int    `json:"flags"`
	Offset                      int    `json:"offset"`
	ArrayLength                 int    `json:"arrayLength"`
	ArrayLengthOffset           int    `json:"arrayLengthOffset"`
	SerializedBitCount          int    `json:"serializedBitCount"`
	SerializedArraySizeBitCount int    `json:"serializedArraySizeBitCount"`
	Group                       int    `json:"group"`
	TagMapType                  int    `json:"tagMapType"`
}

type TypeName = string
type TypeHash = int
type FieldName = string
type FieldHash = int
type FormatHash = int

type Definitions struct {
	TypeHashToDef    map[TypeHash]Definition
	FieldHashToName  map[FieldHash]FieldName
	formatToTypeHash map[FormatHash]TypeHash
	nameToTypeHash   map[TypeName]TypeHash
}

func (d Definitions) GetByTypeHash(typeHash TypeHash) (def Definition, ok bool) {
	def, ok = d.TypeHashToDef[typeHash]
	return
}

func (d Definitions) GetByFormatHash(formatHash FormatHash) (TypeHash, Definition, bool) {
	typeHash, ok := d.formatToTypeHash[formatHash]
	if !ok {
		return 0, Definition{}, false
	}

	def, ok := d.TypeHashToDef[typeHash]
	if !ok {
		return 0, Definition{}, false
	}

	return typeHash, def, true
}

func (d Definitions) GetByName(name TypeName) (TypeHash, Definition, bool) {
	typeHash, ok := d.nameToTypeHash[name]
	if !ok {
		return 0, Definition{}, false
	}

	def, ok := d.TypeHashToDef[typeHash]
	if !ok {
		return 0, Definition{}, false
	}

	return typeHash, def, true
}

func LoadDefinitions(path string) (defs Definitions, err error) {
	// Open file
	f, err := os.Open(path)
	if err != nil {
		return defs, err
	}
	defer f.Close()

	// Read file
	data, err := io.ReadAll(f)
	if err != nil {
		return defs, err
	}

	// Parse JSON
	var parsed map[string]Definition
	if err = json.Unmarshal(data, &parsed); err != nil {
		return defs, err
	}

	// Transform
	defs.TypeHashToDef = make(map[int]Definition, len(parsed))
	defs.FieldHashToName = make(map[int]string)
	defs.nameToTypeHash = make(map[string]int, len(parsed))
	defs.formatToTypeHash = make(map[int]int)

	for typeHashStr, def := range parsed {
		typeHash, err := strconv.Atoi(typeHashStr)
		if err != nil {
			return defs, err
		}

		defs.TypeHashToDef[typeHash] = def
		defs.nameToTypeHash[def.Name] = typeHash
		if def.FormatHash != 0 {
			defs.formatToTypeHash[def.FormatHash] = typeHash
		}

		for _, field := range def.Fields {
			defs.FieldHashToName[field.Hash] = field.Name
		}
	}

	return defs, nil
}
