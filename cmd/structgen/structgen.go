package main

import (
	"errors"
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4data"
	"github.com/dave/jennifer/jen"
	mapset "github.com/deckarep/golang-set/v2"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
	"os"
	"strconv"
	"strings"
	"unicode"
)

var (
	BasicTypes = mapset.NewThreadUnsafeSet[string](
		"DT_NULL",
		"DT_BYTE",
		"DT_WORD",
		"DT_ENUM",
		"DT_INT",
		"DT_FLOAT",
		"DT_OPTIONAL",
		"DT_SNO",
		"DT_SNO_NAME",
		"DT_GBID",
		"DT_STARTLOC_NAME",
		"DT_UINT",
		"DT_ACD_NETWORK_NAME",
		"DT_SHARED_SERVER_DATA_ID",
		"DT_INT64",
		"DT_RANGE",
		"DT_FIXEDARRAY",
		"DT_TAGMAP",
		"DT_VARIABLEARRAY",
		"DT_POLYMORPHIC_VARIABLEARRAY",
		"DT_STRING_FORMULA",
		"DT_CSTRING",
		"DT_CHARARRAY",
		"DT_RGBACOLOR",
		"DT_RGBACOLORVALUE",
		"DT_BCVEC2I",
		"DT_VECTOR2D",
		"DT_VECTOR3D",
		"DT_VECTOR4D",
	)
	ReqParamTypes = mapset.NewThreadUnsafeSet[string](
		"DT_CSTRING",
		"DT_OPTIONAL",
		"DT_RANGE",
		"DT_FIXEDARRAY",
		"DT_TAGMAP",
		"DT_VARIABLEARRAY",
		"DT_POLYMORPHIC_VARIABLEARRAY",
	)

	ErrComposeUnknownTypeHash = errors.New("cannot compose unknown type hash")
	ErrNullNotfound           = errors.New("could not find DT_NULL type hash")
)

func SortedKeys[K constraints.Ordered, V any](m map[K]V) []K {
	// TODO: maybe move this into pkg/d4data/definitions so we dont iter + sort multiple times
	ks := make([]K, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	slices.Sort(ks)
	return ks
}

func TrackCurrentPosCode() jen.Code {
	return jen.List(jen.Id("p"), jen.Err()).
		Op(":=").
		Id("r").Dot("Pos").Call().
		Line().If(
		jen.Err().Op("!=").Nil(),
	).Block(
		jen.Return(jen.Err()),
	)
}

func SeekRelativeCode(offset int) jen.Code {
	return jen.If(
		jen.List(jen.Id("_"), jen.Err()).Op(":=").
			Id("r").Dot("Seek").
			Call(
				jen.Id("p").Op("+").Lit(offset),
				jen.Qual("io", "SeekStart"),
			),
		jen.Err().Op("!=").Nil(),
	).Block(
		jen.Return(jen.Err()),
	)
}

func UnmarshalFieldCode(transformedName string, field d4data.Field) []jen.Code {
	var code []jen.Code

	// Create options code
	optionsDict := jen.Dict{}

	if field.Flags != 0 {
		optionsDict[jen.Id("Flags")] = jen.Lit(field.Flags)
	}
	if field.ArrayLength > -1 {
		optionsDict[jen.Id("ArrayLength")] = jen.Lit(field.ArrayLength)
	}
	if field.Group > -1 {
		optionsDict[jen.Id("Group")] = jen.Lit(field.Group)
	}

	code = append(
		code,
		jen.If(
			jen.Err().Op(":=").Id("UnmarshalAt").Call(
				jen.Id("p").Op("+").Lit(field.Offset),
				jen.Op("&").Id("t").Dot(transformedName),
				jen.Id("r"),
				jen.Op("&").Id("Options").Values(optionsDict),
			),
			jen.Err().Op("!=").Nil(),
		).Block(
			jen.Return(jen.Err()),
		),
	)

	return code
}

func WalkFieldCode(transformedName string) []jen.Code {
	return []jen.Code{
		jen.Id("cb").Dot("Do").Call(
			jen.Lit(transformedName),
			jen.Op("&").Id("t").Dot(transformedName),
		),
	}
}

func TransformFieldName(fieldName string) string {
	r := []rune(fieldName)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func composeTypes(defs d4data.Definitions, types []d4data.TypeHash, hasParent bool) (jen.Code, error) {
	switch len(types) {
	case 0:
		return nil, nil
	case 1: // Type param
		t, ok := defs.GetByTypeHash(types[0])
		if !ok {
			return nil, ErrComposeUnknownTypeHash
		}

		// This is a weird edge case. We think the internal UI allows to select 3 types default to NULL. Cases like this
		// are likely a mistake. So, if the final types requires a type param, use null.
		if ReqParamTypes.Contains(t.Name) {
			nullTypeHash, _, ok := defs.GetByName("DT_NULL")
			if !ok {
				return nil, ErrNullNotfound
			}
			types = append(types, nullTypeHash)
			return composeTypes(defs, types, hasParent)
		}

		if hasParent {
			return jen.Op("*").Id(t.Name), nil
		}
		return jen.Id(t.Name), nil
	default: // Composed type
		t, ok := defs.GetByTypeHash(types[0])
		if !ok {
			return nil, ErrComposeUnknownTypeHash
		}

		sub, err := composeTypes(defs, types[1:], true)
		if err != nil {
			return nil, err
		}

		if hasParent {
			return jen.Op("*").Id(t.Name).Types(sub), nil
		}
		return jen.Id(t.Name).Types(sub), nil
	}
}

func ComposeTypes(defs d4data.Definitions, types []d4data.TypeHash) (jen.Code, error) {
	// Truncate to everything before first null type
	nullTypeHash, _, ok := defs.GetByName("DT_NULL")
	if !ok {
		return nil, ErrNullNotfound
	}

	for i, typeHash := range types {
		if nullTypeHash == typeHash {
			types = types[:i]
			break
		}
	}

	return composeTypes(defs, types, false)
}

func GenerateFormatHashMapFunc(f *jen.File, defs d4data.Definitions) error {
	// TODO: could also replace this with a map from formatHash to typeHash
	var cases []jen.Code

	for _, typeHash := range SortedKeys(defs.TypeHashToDef) {
		def := defs.TypeHashToDef[typeHash]
		if def.FormatHash > 0 {
			cases = append(
				cases,
				jen.Case(jen.Lit(def.FormatHash)).Block(
					jen.Return(jen.Op("&").Id(def.Name).Block()),
				),
			)
		}
	}

	cases = append(
		cases,
		jen.Default().Block(
			jen.Return(jen.Nil()),
		),
	)

	f.Func().Id("NewByFormatHash").Params(
		jen.Id("h").Int(),
	).Id("Object").Block(
		jen.Switch(jen.Id("h")).Block(cases...),
	).Line()

	return nil
}

func GenerateTypeHashMapFunc(f *jen.File, defs d4data.Definitions) error {
	var cases []jen.Code

	for _, typeHash := range SortedKeys(defs.TypeHashToDef) {
		def := defs.TypeHashToDef[typeHash]
		var case_ jen.Code

		if ReqParamTypes.Contains(def.Name) {
			case_ = jen.Return(jen.Op("&").Id(def.Name).Types(jen.Id("T")).Block())
		} else {
			case_ = jen.Return(jen.Op("&").Id(def.Name).Block())
		}

		cases = append(
			cases,
			jen.Case(jen.Lit(typeHash)).Block(case_),
		)
	}

	cases = append(
		cases,
		jen.Default().Block(
			jen.Return(jen.Nil()),
		),
	)

	f.Func().Id("NewByTypeHash").Types(jen.Id("T").Id("Object")).Params(
		jen.Id("h").Int(),
	).Id("Object").Block(
		jen.Switch(jen.Id("h")).Block(cases...),
	).Line()

	return nil
}

func GenerateStruct(f *jen.File, defs d4data.Definitions, def d4data.Definition) error {
	// Skip basic types
	if BasicTypes.Contains(def.Name) {
		return nil
	}

	// Construct fields
	fields := make([]jen.Code, 0, len(def.Fields)+len(def.Inherits))
	unmarshalD4Body := []jen.Code{
		TrackCurrentPosCode(),
	}
	var walkBody []jen.Code

	for _, inheritTypeHash := range def.Inherits { // Add inherits comments; TODO: incorporate actual struct embedding
		var inheritTypeName string
		if inheritTypeDef, ok := defs.GetByTypeHash(inheritTypeHash); ok {
			inheritTypeName = inheritTypeDef.Name
		} else {
			inheritTypeName = strconv.Itoa(inheritTypeHash)
		}
		fields = append(fields, jen.Commentf("// Inherits %s", inheritTypeName))
	}

	for _, field := range def.Fields { // Add fields
		// Get field info
		fieldTypes := field.Type[:]
		fieldName := TransformFieldName(field.Name)
		fieldTypeCode, err := ComposeTypes(defs, fieldTypes)
		if err != nil {
			return err
		}

		// Add type fields code
		fields = append(fields, jen.Id(fieldName).Add(fieldTypeCode))

		// Add UnmarshalD4 body code
		unmarshalD4Body = append(unmarshalD4Body, UnmarshalFieldCode(fieldName, field)...)

		// Add Walk body code
		walkBody = append(walkBody, WalkFieldCode(fieldName)...)
	}

	// Construct type
	f.Type().Id(def.Name).Struct(fields...).Line()

	// Construct UnmarshalD4 function
	unmarshalD4Body = append(unmarshalD4Body, SeekRelativeCode(def.Size))
	unmarshalD4Body = append(unmarshalD4Body, jen.Return(jen.Nil()))

	f.Func().Params(
		jen.Id("t").Op("*").Id(def.Name),
	).Id("UnmarshalD4").Params(
		jen.Id("r").Op("*").Qual("github.com/Dakota628/d4parse/pkg/bin", "BinaryReader"),
		jen.Id("o").Op("*").Id("Options"),
	).Error().Block(
		unmarshalD4Body...,
	).Line()

	// Construct Walk function
	f.Func().Params(
		jen.Id("t").Op("*").Id(def.Name),
	).Id("Walk").Params(
		jen.Id("cb").Id("WalkCallback"),
	).Block(
		walkBody...,
	).Line()

	return nil
}

func GenerateStructs(defs d4data.Definitions, outputPath string) error {
	// Create file with standard generated code header
	f := jen.NewFile("d4")
	f.HeaderComment(fmt.Sprintf("Code generated by structgen %s; DO NOT EDIT.", strings.Join(os.Args[1:], " ")))

	// Generate mapping funcs
	if err := GenerateFormatHashMapFunc(f, defs); err != nil {
		return err
	}
	if err := GenerateTypeHashMapFunc(f, defs); err != nil {
		return err
	}

	// Generate structs
	for _, typeHash := range SortedKeys(defs.TypeHashToDef) {
		if err := GenerateStruct(f, defs, defs.TypeHashToDef[typeHash]); err != nil {
			return err
		}
	}

	// Write file
	return f.Save(outputPath)
}

func main() {
	if len(os.Args) < 3 {
		slog.Error("usage: structgen definitionsPath outputPath")
		os.Exit(1)
	}

	definitionsPath := os.Args[1]
	outputPath := os.Args[2]

	defs, err := d4data.LoadDefinitions(definitionsPath)
	if err != nil {
		slog.Error("Failed to load definitions", slog.Any("err", err))
		os.Exit(1)
	}

	if err := GenerateStructs(defs, outputPath); err != nil {
		slog.Error("Failed to generate structs", slog.Any("err", err))
		os.Exit(1)
	}
}
