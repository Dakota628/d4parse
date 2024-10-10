package main

import (
	"errors"
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/Dakota628/d4parse/pkg/d4data"
	"github.com/dave/jennifer/jen"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
	"os"
	"strconv"
	"strings"
	"unicode"
)

var (
	ErrComposeUnknownTypeHash = errors.New("cannot compose unknown type hash")
	ErrNullNotFound           = errors.New("could not find DT_NULL type hash")
)

func GetBasicTypeAlignment(defs d4data.Definitions, def d4data.Definition, inTagMap bool, typeHashes ...d4data.TypeHash) int {
	switch def.Hash {
	case defs.HashByName("DT_NULL"):
		return 0
	case defs.HashByName("DT_POLYMORPHIC_VARIABLEARRAY"),
		defs.HashByName("DT_STRING_FORMULA"),
		defs.HashByName("DT_VARIABLEARRAY"),
		defs.HashByName("DT_TAGMAP"),
		defs.HashByName("DT_CSTRING"):
		if inTagMap {
			return 4
		}
		return 8
	case defs.HashByName("DT_CHARARRAY"):
		return 1
	case defs.HashByName("DT_FIXEDARRAY"):
		return GetTypeAlignment(defs, false, typeHashes[1:]...)
	case defs.HashByName("DT_OPTIONAL"):
		return GetTypeAlignment(defs, false, typeHashes[1:]...) // 2 for DT_WORD, 8 for DT_INT?
	case defs.HashByName("DT_SNO_NAME"):
		return 4
	case defs.HashByName("DT_RANGE"):
		return GetTypeAlignment(defs, false, typeHashes[1:]...)
	default:
		return def.Size
	}
}

func GetTypeAlignment(defs d4data.Definitions, inTagMap bool, typeHashes ...d4data.TypeHash) int {
	if len(typeHashes) == 0 {
		return 0
	}

	def, ok := defs.GetByTypeHash(typeHashes[0])
	if !ok {
		slog.Warn(
			"No alignment information for unknown type",
			slog.String("name", def.Name),
			slog.Int("hash", def.Hash),
		)
		return 4
	}

	if def.Fields == nil {
		if def.Type == "basic" {
			return GetBasicTypeAlignment(defs, def, inTagMap, typeHashes...)
		}

		slog.Warn(
			"No alignment information for type",
			slog.String("name", def.Name),
			slog.Int("hash", def.Hash),
		)
		return 4
	}

	if len(def.Fields) == 0 && def.Fields != nil { // Checking nil again here in case above code changes and causes hard to spot bug
		return 1
	}

	var alignment int
	for _, field := range def.Fields {
		// Since the field.Type slice is always size 3, just manually index here instead of iter and casting the slice
		alignment = max(alignment, GetTypeAlignment(defs, false, field.Type[0], field.Type[1], field.Type[2]))
	}

	return alignment
}

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

func UnmarshalFieldCode(transformedName string, field d4data.Field, defs d4data.Definitions) []jen.Code {
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
	if field.TagMapType > -1 {
		if def, ok := defs.GetByTypeHash(field.TagMapType); ok {
			optionsDict[jen.Id("TagMapType")] = jen.Op("&").Id(def.Name).Block()
		}
	}

	code = append(
		code,
		jen.If(
			jen.Err().Op(":=").Id("UnmarshalAt").Call(
				jen.Id("p").Op("+").Lit(field.Offset),
				jen.Op("&").Id("t").Dot(transformedName),
				jen.Id("r"),
				jen.Op("&").Id("FieldOptions").Values(optionsDict),
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
			jen.Id("d").Op("..."),
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
		if d4.DefFlagHasSubType.In(t.Flags) {
			nullTypeHash, _, ok := defs.GetByName("DT_NULL")
			if !ok {
				return nil, ErrNullNotFound
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
		return nil, ErrNullNotFound
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

		if d4.DefFlagHasSubType.In(def.Flags) {
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

func GenerateTypeNameMapFunc(f *jen.File, defs d4data.Definitions) error {
	var cases []jen.Code

	for _, typeHash := range SortedKeys(defs.TypeHashToDef) {
		def := defs.TypeHashToDef[typeHash]

		case_ := jen.Return(jen.Lit(typeHash))

		cases = append(
			cases,
			jen.Case(jen.Lit(def.Name)).Block(case_),
		)
	}

	cases = append(
		cases,
		jen.Default().Block(
			jen.Return(jen.Lit(0)),
		),
	)

	f.Func().Id("TypeHashByName").Params(
		jen.Id("s").String(),
	).Int().Block(
		jen.Switch(jen.Id("s")).Block(cases...),
	).Line()

	return nil
}

func GenerateTagMapFieldHashMapFunc(f *jen.File, defs d4data.Definitions) error {
	var cases []jen.Code

	for _, fieldHash := range SortedKeys(defs.FieldHashToName) {
		fieldName := defs.FieldHashToName[fieldHash]
		cases = append(
			cases,
			jen.Case(jen.Lit(fieldHash)).Block(
				jen.Return(jen.Lit(fieldName)),
			),
		)
	}

	cases = append(
		cases,
		jen.Default().Block(
			jen.Return(jen.Lit("")),
		),
	)

	f.Func().Id("NameByFieldHash").Params(
		jen.Id("h").Int(),
	).Id("string").Block(
		jen.Switch(jen.Id("h")).Block(cases...),
	).Line()

	return nil
}

func GenerateTypeOptions(options d4.TypeOptions) jen.Code {
	optionsDict := jen.Dict{}

	if options.Flags != 0 {
		optionsDict[jen.Id("Flags")] = jen.Lit(options.Flags)
	}
	if options.Alignment != 0 {
		optionsDict[jen.Id("Alignment")] = jen.Lit(options.Alignment)
	}
	if options.TagMapAlignment != 0 {
		optionsDict[jen.Id("TagMapAlignment")] = jen.Lit(options.TagMapAlignment)
	}

	return jen.Id("TypeOptions").Values(optionsDict)
}

func GenerateOptionsMapFunc(f *jen.File, defs d4data.Definitions) error {
	var cases []jen.Code

	for _, typeHash := range SortedKeys(defs.TypeHashToDef) {
		def := defs.TypeHashToDef[typeHash]

		cases = append(
			cases,
			jen.Case(jen.Lit(typeHash)).Block(
				jen.Return(GenerateTypeOptions(d4.TypeOptions{
					Flags:           def.Flags,
					Alignment:       GetTypeAlignment(defs, false, def.Hash),
					TagMapAlignment: GetTypeAlignment(defs, true, def.Hash),
				})),
			),
		)
	}

	cases = append(
		cases,
		jen.Default().Block(
			jen.Return(GenerateTypeOptions(d4.TypeOptions{})),
		),
	)

	f.Func().Id("OptionsForType").Params(
		jen.Id("h").Int(),
	).Id("TypeOptions").Block(
		jen.Switch(jen.Id("h")).Block(cases...),
	).Line()

	return nil
}

func GenerateForAllTypes(f *jen.File, _ d4data.Definitions, def d4data.Definition) error {
	hasSubType := d4.DefFlagHasSubType.In(def.Flags)

	// Generate receiver code
	var receiver jen.Code
	if hasSubType {
		receiver = jen.Id("t").Op("*").Id(def.Name).Types(jen.Id("T"))
	} else {
		receiver = jen.Id("t").Op("*").Id(def.Name)
	}

	// Construct TypHash function
	f.Func().Params(
		receiver,
	).Id("TypeHash").Params().Int().Block(
		jen.Return(jen.Lit(def.Hash)),
	).Line()

	// Construct SubTypeHash function
	if hasSubType {
		f.Func().Params(
			receiver,
		).Id("SubTypeHash").Params().Int().Block(
			jen.Return(
				jen.Id("nilObject").Types(jen.Id("T")).Call().Dot("TypeHash").Call(),
			),
		).Line()
	} else {
		f.Func().Params(
			receiver,
		).Id("SubTypeHash").Params().Int().Block(
			jen.Return(jen.Lit(0)),
		).Line()
	}

	// Construct Size function
	f.Func().Params(
		receiver,
	).Id("TypeSize").Params().Int64().Block(
		jen.Return(jen.Lit(def.Size)),
	).Line()

	return nil
}

func GenerateHashBody(f *jen.File, defs d4data.Definitions, def d4data.Definition) []jen.Code {
	var code []jen.Code
	fields := make(map[int]string, len(def.Fields)) // field hash -> field name

	// Add to map to sort
	for _, field := range def.Fields {
		fields[field.Hash] = TransformFieldName(field.Name)
	}

	// Generate the code
	// TODO: add type hash?

	for _, fieldHash := range SortedKeys(fields) {
		fieldName := fields[fieldHash]

		code = append(
			code,
			jen.If(
				jen.Err().Op(":=").Id("hashField").Call(
					jen.Id("h"),
					jen.Lit(fieldHash),
					jen.Op("&").Id("t").Dot(fieldName),
				),
				jen.Err().Op("!=").Nil(),
			).Block(
				jen.Return(jen.Err()),
			),
		)
	}

	code = append(code, jen.Return(jen.Nil()))

	return code
}

func GenerateStruct(f *jen.File, defs d4data.Definitions, def d4data.Definition) error {
	// Skip basic types
	if def.Type == "basic" {
		return GenerateForAllTypes(f, defs, def)
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
		unmarshalD4Body = append(unmarshalD4Body, UnmarshalFieldCode(fieldName, field, defs)...)

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
		jen.Id("o").Op("*").Id("FieldOptions"),
	).Error().Block(
		unmarshalD4Body...,
	).Line()

	// Construct Walk function
	f.Func().Params(
		jen.Id("t").Op("*").Id(def.Name),
	).Id("Walk").Params(
		jen.Id("cb").Id("WalkCallback"),
		jen.Id("d").Op("...").Any(),
	).Block(
		walkBody...,
	).Line()

	// Construct Hash function
	f.Func().Params(
		jen.Id("t").Op("*").Id(def.Name),
	).Id("Hash").Params(
		jen.Id("h").Qual("hash", "Hash"),
	).Error().Block(
		GenerateHashBody(f, defs, def)...,
	).Line()

	return GenerateForAllTypes(f, defs, def)
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
	if err := GenerateTypeNameMapFunc(f, defs); err != nil {
		return err
	}
	if err := GenerateTagMapFieldHashMapFunc(f, defs); err != nil {
		return err
	}
	if err := GenerateOptionsMapFunc(f, defs); err != nil {
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
