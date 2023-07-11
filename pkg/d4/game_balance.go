package d4

import (
	"reflect"
	"sync"
	"unicode"
)

type GbData = sync.Map // DT_GBID -> GbInfo

type GbInfo struct {
	SnoId int32
	Name  string
}

func GbidHash(name string) (hash uint32) {
	for _, c := range name {
		if c == 0 {
			return
		}
		hash = (hash << 5) + hash + uint32(unicode.ToLower(c))
	}
	return
}

func GetGbidHeaders(gbDef *GameBalanceDefinition) (headers []GBIDHeader) {
	reflectZero := reflect.Value{}

	for _, gbTable := range gbDef.PtData.Value {
		rvGbTable := reflect.ValueOf(gbTable)
		if rvGbTable == reflectZero {
			continue
		}
		if rvGbTable.Kind() == reflect.Ptr {
			rvGbTable = rvGbTable.Elem()
		}

		entries := rvGbTable.FieldByName("TEntries")
		if entries == reflectZero {
			continue
		}
		entries = entries.FieldByName("Value")
		if entries == reflectZero {
			continue
		}

		for i := 0; i < entries.Len(); i++ {
			headerField := entries.Index(i).Elem().FieldByName("THeader")
			if headerField == reflectZero {
				continue
			}

			if header, ok := headerField.Interface().(GBIDHeader); ok {
				headers = append(headers, header)
			}
		}
	}

	return
}
