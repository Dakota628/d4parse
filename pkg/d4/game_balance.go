package d4

import (
	"github.com/alphadose/haxmap"
	"reflect"
)

type GbId = uint32

type GbData = *haxmap.Map[GbId, GbInfo]

type GbInfo struct {
	SnoId int32
	Name  string
}

func GetGbHeader(gbDef *GameBalanceDefinition) (headers []GBIDHeader) {
	reflectZero := reflect.Value{}

	for _, gbTable := range gbDef.PtData.Value {
		rvGbTable := reflect.ValueOf(gbTable).Elem()
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
