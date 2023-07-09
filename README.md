# d4parse

Parses various D4 files.

## CLI Tools

### dumpsnometa
Dumps SNO meta files as spew output.

Example: `dumpsnometa World_SecretCellar.qst > World_SecretCellar.qst.dump`

### dumptoc
Dumps TOC files as YAML output.

Example: `dumptoc CoreTOC.dat > CoreTOC.yaml`

### structgen
Generates Go code to parse SNO meta structs. Uses definitions from d4data.

Example: `structgen d4data/definitions.json generated_types.go`
