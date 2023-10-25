package d4

type FieldFlag int

const (
	FieldFlagX1                 FieldFlag = 0x1      // TDF_XML_SERIALIZATION
	FieldFlagX2                 FieldFlag = 0x2      // TDF_XGROWABLE_ARRAY
	FieldFlagX4                 FieldFlag = 0x4      // 50d5e6e: DT_UINT/DT_CSTRING/DT_CHARARRAY szName;
	FieldFlagIsFixedArray       FieldFlag = 0x8      // ?
	FieldFlagHasValueConstraint FieldFlag = 0x10     // mostly DT_BOOL and DT_ENUM
	FieldFlagX20                FieldFlag = 0x20     // TDF_XPTR64
	FieldFlagX40                FieldFlag = 0x40     // TDF_IGNORE_ON_SERVER
	FieldFlagX80                FieldFlag = 0x80     // TDF_XFLOAT16
	FieldFlagRequiredSNO        FieldFlag = 0x100    // TDF_REQUIRED_SNO
	FieldFlagSoftLink           FieldFlag = 0x200    // TDF_SOFT_LINK: SNOs, MarkerHandle, DT_TAGMAP (adf9a5f)
	FieldFlagX400               FieldFlag = 0x400    // TDF_SERVER_ONLY
	FieldFlagX800               FieldFlag = 0x800    // TDF_PC_ONLY
	FieldFlagX1000              FieldFlag = 0x1000   // ?
	FieldFlagX2000              FieldFlag = 0x2000   // ?
	FieldFlagX4000              FieldFlag = 0x4000   // TDF_AXE_ONLY
	FieldFlagX8000              FieldFlag = 0x8000   // TDF_XPTR128
	FieldFlagX10000             FieldFlag = 0x10000  // TDF_COMPUTE_ARRAY_COUNT
	FieldFlagX20000             FieldFlag = 0x20000  // TDF_DISABLE_TAGMAP_OPTIMIZATION
	FieldFlagX40000             FieldFlag = 0x40000  // TDF_DISABLE_DATA_MERGE
	FieldFlagIsFlag             FieldFlag = 0x80000  // TDF_BIT_FLAGS
	FieldFlagX100000            FieldFlag = 0x100000 // TDF_XBOXONE_ONLY; 510bb32: DT_CSTRING szText;
	FieldFlagPayload            FieldFlag = 0x200000 // DT_VARIABLEARRAY pointing into Payload. Exception: 4a1717c FogMask SceneDefinition::tFogMask;
	FieldFlagPayload2           FieldFlag = 0x400000
	FieldFlagX800000            FieldFlag = 0x800000  // ?
	FieldFlagX1000000           FieldFlag = 0x1000000 // ?
	FieldFlagX2000000           FieldFlag = 0x2000000 // ?
	FieldFlagX4000000           FieldFlag = 0x4000000 // ?
	FieldFlagX8000000           FieldFlag = 0x8000000 // ?
)

func (f FieldFlag) In(i int) bool {
	return (i & int(f)) > 0
}

type DefFlag int

const (
	DefFlagX1                         DefFlag = 0x1   // TD_BEING_SCANNED_FOR_FIELDS?
	DefFlagX2                         DefFlag = 0x2   // TD_SCANNED_FOR_FIELDS?
	DefFlagX4                         DefFlag = 0x4   // ?
	DefFlagContainsSNOSubField        DefFlag = 0x8   // TD_CONTAINS_SNO_SUBFIELD
	DefFlagContainsAllocationSubField DefFlag = 0x10  // TD_CONTAINS_ALLOCATION_SUBFIELD
	DefFlagX20                        DefFlag = 0x20  // TD_CONTAINS_SERVER_ONLY_SUBFIELD?
	DefFlagContainsTagMapSubField     DefFlag = 0x40  // TD_CONTAINS_TAGMAP_SUBFIELD
	DefFlagX80                        DefFlag = 0x80  // only set for some basic types
	DefFlagIsComplex                  DefFlag = 0x100 // TD_IS_COMPLEX_TYPE
	DefFlagRelatedToSNOTypeDefinition DefFlag = 0x200
	DefFlagSNOTypeDefinition          DefFlag = 0x4000
	DefFlagHasSubType                 DefFlag = 0x8000
	DefFlagReferencesFileLocation     DefFlag = 0x10000 // references another location in the same file
	DefFlagIsPolymorphic              DefFlag = 0x20000
	DefFlagX40000                     DefFlag = 0x40000
	DefFlagX80000                     DefFlag = 0x80000
)

func (f DefFlag) In(i int) bool {
	return (i & int(f)) > 0
}
