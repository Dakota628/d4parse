package html

import (
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"golang.org/x/exp/slog"
	"reflect"
	"strings"
	"unicode"
)

const css = `* { 
	box-sizing: border-box;
}
html, body {
	min-height: 100%;
	background: #212529;
	color: #f8f9fa;
	font-family: sans-serif;
}
	
body {
	font-size: 0.75rem;
	text-rendering: optimizeLegibility;
}
	
body, ul, ol, dl {
	margin: 0;
}

article, aside, audio, 
footer, header, nav, section, video {
	display: block; 
}
	
p { 
	-ms-word-break: break-all;
	word-break: break-all;
	word-break: break-word;
	-moz-hyphens: auto;
	-webkit-hyphens: auto;
	-ms-hyphens: auto;
	hyphens: auto;
	padding-left: .5rem;
}
	
textarea { 
	resize: vertical;
}
 
table { 
  border-collapse: collapse;
}

td {
	padding: .5rem;
}

img { 
	border: none;
	max-width: 100%;
}
	
input[type="submit"]::-moz-focus-inner, input[type="button"]::-moz-focus-inner {
	border : 0px;
}
	
input[type="search"] { 
	-webkit-appearance: textfield;
} 
input[type="submit"] { 
	-webkit-appearance:none;
}
	
input:required:after {
	color: #f00;
	content: " *";
}

input[type="email"]:invalid { 
	background: #f00;
}
	
sub, sup { 
	line-height: 0;
}

a {
	color: #4c6ef5;
	text-decoration: none;
	line-height: 1;
}

a:hover {
	text-decoration: underline;
}

.type {
	width: 100%;
	clear: both;
}

.typeName {
	width: 100%;
	background: #a61e4d;
	padding: 0.5rem;
	font-size: 1rem;
	line-height: 1.25;
	overflow-wrap: anywhere;
	font-weight: bold;
	clear: both;
}

.field {
	overflow: auto;
	clear: both;
}

.fieldKey {
	width: 10rem;
	float: left;
	margin-bottom: 0.2rem;
	padding: 0.5rem;
	background: #343a40;
}

.fieldKey > .fieldName {
	float: left;
	width: 100%;
	font-size: 0.9rem;
	line-height: 1.5;
	overflow-wrap: anywhere;
}

.fieldKey > .fieldType {
	width: 100%;
	font-size: 0.7rem;
	color: #dee2e6;
	overflow-wrap: anywhere;
}

.fieldValue {
	width: calc(100% - 10.2rem);
	margin-left: 0.2rem;
	float: left;
	font-size: 0.9rem;
	overflow-wrap: anywhere;
}

ul.array {
	list-style-type: none;
	margin: 0;
	padding: 0;
	width: 100%;
}

ul.array > li {
	margin: 0;
	padding: 0;
	display: inline-block;
	clear: both;
	width: 100%;
}` // TODO: javascript to stick most recently scrolled past field value to the top

type HtmlGenerator struct {
	sb         strings.Builder
	tocEntries d4.TocEntries
}

func NewHtmlGenerator(toc d4.Toc) *HtmlGenerator {
	return &HtmlGenerator{
		tocEntries: toc.Entries,
	}
}

func (h *HtmlGenerator) genericType(rtStr string) string {
	i := strings.IndexByte(rtStr, '[')
	if i <= 0 {
		return rtStr
	}
	return rtStr[:i]
}

func (h *HtmlGenerator) genericField(rv reflect.Value, field string) any {
	return rv.Elem().FieldByName(field).Interface()
}

func (h *HtmlGenerator) prettyTypeName(typeName string) string {
	typeName = strings.Replace(typeName, "*github.com/Dakota628/d4parse/pkg/d4.", "", -1)
	typeName = strings.Replace(typeName, "*d4.", "", -1)
	typeName = strings.Replace(typeName, "d4.", "", -1)
	return typeName
}

func (h *HtmlGenerator) prettyFieldName(fieldName string) string {
	r := []rune(fieldName)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func (h *HtmlGenerator) writeFmt(format string, a ...any) {
	h.sb.WriteString(fmt.Sprintf(format, a...)) // TODO: utilize Fprintf
}

func (h *HtmlGenerator) add(x d4.UnmarshalBinary) {
	// Fast path
	switch t := x.(type) {
	case *d4.SnoMeta:
		group, name := h.tocEntries.GetName(t.Id.Value)
		h.sb.WriteString(`<div class="type snoMeta"><div class="typeName">SNO Info</div>`)
		h.sb.WriteString(`<div class="field"><div class="fieldKey"><div class="fieldName">Group</div></div>`)
		h.writeFmt(`<div class="fieldValue">%d</div></field>`, group) // TODO: group string
		h.sb.WriteString(`<div class="field"><div class="fieldKey"><div class="fieldName">ID</div></div>`)
		h.writeFmt(`<div class="fieldValue">%d</div></div>`, t.Id.Value)
		h.sb.WriteString(`<div class="field"><div class="fieldKey"><div class="fieldName">Name</div></div>`)
		h.writeFmt(`<div class="fieldValue">%s</div></div>`, name)
		h.sb.WriteString("</div>")
		h.add(t.Meta)
		return
	case *d4.DT_NULL:
		h.writeFmt("<i>null</i>")
		return
	case *d4.DT_BYTE:
		h.writeFmt("<p>0x%x</p>", t.Value)
		return
	case *d4.DT_WORD:
		h.writeFmt("<p>0x%x</p>", t.Value)
		return
	case *d4.DT_ENUM:
		h.writeFmt("<p>%d</p>", t.Value)
		return
	case *d4.DT_INT:
		h.writeFmt("<p>%d</p>", t.Value)
		return
	case *d4.DT_FLOAT:
		h.writeFmt("<p>%f</p>", t.Value)
		return
	case *d4.DT_SNO:
		_, name := h.tocEntries.GetName(t.Id) // TODO: get group string and output
		h.writeFmt(`<p><a class="snoRef" href="sno/%d">%s</a></p>`, t.Id, name)
		return
	case *d4.DT_SNO_NAME:
		_, name := h.tocEntries.GetName(t.Id, d4.SnoGroup(t.Group)) // TODO: get group string and output
		h.writeFmt(`<p><a class="snoRef" href="sno/%d">%s</a></p>`, t.Id, name)
		return
	case *d4.DT_GBID:
		// TODO: need to enrich with actual gbid name
		h.writeFmt(`<p><a class="gbidRef" href="gbid/%d">%d</a><p>`, t.Value, t.Value)
		return
	case *d4.DT_STARTLOC_NAME:
		// TODO: can it be enriched?
		h.writeFmt("<p>%d</p>", t.Value)
		return
	case *d4.DT_UINT:
		h.writeFmt("<p>%d</p>", t.Value)
		return
	case *d4.DT_ACD_NETWORK_NAME:
		// TODO: can it be enriched?
		h.writeFmt("<p>%d</p>", t.Value)
		return
	case *d4.DT_SHARED_SERVER_DATA_ID:
		// TODO: can it be enriched?
		h.writeFmt("<p>%d</p>", t.Value)
		return
	case *d4.DT_INT64:
		h.writeFmt("<p>%d</p>", t.Value)
		return
	case *d4.DT_STRING_FORMULA:
		// TODO: do we want compiled in here too? Seems useless.
		h.sb.WriteString(`<code class="formula">`)
		h.sb.WriteString(t.Value)
		h.sb.WriteString("</code>")
		return
	case *d4.DT_CHARARRAY:
		h.sb.WriteString("<pre>")
		h.sb.WriteString(string(t.Value))
		h.sb.WriteString("</pre>")
		return
	case *d4.DT_RGBACOLOR:
		h.writeFmt("<p>#%x%x%x%x</p>", t.B, t.G, t.B, t.A)
		return
	case *d4.DT_RGBACOLORVALUE:
		h.writeFmt("<p>rgba(%f, %f, %f, %f)</p>", t.B, t.G, t.B, t.A)
		return
	case *d4.DT_BCVEC2I:
		h.writeFmt("<p>(%f, %f)</p>", t.X, t.Y)
		return
	case *d4.DT_VECTOR2D:
		h.writeFmt("<p>(%f, %f)</p>", t.X, t.Y)
		return
	case *d4.DT_VECTOR3D:
		h.writeFmt("<p>(%f, %f, %f)</p>", t.X, t.Y, t.Z)
		return
	case *d4.DT_VECTOR4D:
		h.writeFmt("<p>(%f, %f, %f, %f)</p>", t.X, t.Y, t.Z, t.W)
		return
	}

	// Slow path (reflection)
	xrt := reflect.TypeOf(x)
	xrv := reflect.ValueOf(x)
	baseTypeName := h.genericType(xrt.String())

	switch baseTypeName {
	case "*d4.DT_OPTIONAL":
		if h.genericField(xrv, "Exists").(int32) > 0 {
			h.add(h.genericField(xrv, "Value").(d4.UnmarshalBinary))
		}
		return
	case "*d4.DT_RANGE":
		h.sb.WriteString(`<div class="type">`)
		h.sb.WriteString(`<div class="field"><div class="fieldKey">lowerBound</dt>`)
		h.sb.WriteString(`<div class="fieldValue">`)
		h.add(h.genericField(xrv, "LowerBound").(d4.UnmarshalBinary))
		h.sb.WriteString("</div></div>")
		h.sb.WriteString(`<div class="field"><div class="fieldKey">upperBound</dt>`)
		h.sb.WriteString(`<div class="fieldValue">`)
		h.add(h.genericField(xrv, "UpperBound").(d4.UnmarshalBinary))
		h.sb.WriteString("</div</div>>")
		h.sb.WriteString("</div>")
		return
	case "*d4.DT_FIXEDARRAY", "*d4.DT_VARIABLEARRAY", "*d4.DT_POLYMORPHIC_VARIABLEARRAY":
		h.sb.WriteString(`<ul class="array">`)
		valueRv := xrv.Elem().FieldByName("Value")
		for i := 0; i < valueRv.Len(); i++ {
			h.sb.WriteString("<li>")
			elemRv := valueRv.Index(i)
			h.add(elemRv.Interface().(d4.UnmarshalBinary))
			h.sb.WriteString("</li>")
		}
		h.sb.WriteString("</ul>")
		return
	case "*d4.DT_TAGMAP":
		h.sb.WriteString("<i>tag map parsing unsupported</i>") // TODO
		return
	case "*d4.DT_CSTRING":
		h.sb.WriteString(h.genericField(xrv, "Value").(string))
	default:
		// Non-basic types
		rv := reflect.ValueOf(x).Elem()
		rt := rv.Type()

		h.sb.WriteString(`<div class="type">`)
		h.writeFmt(`<div class="typeName">%s</div>`, h.prettyTypeName(rt.Name()))
		for i := 0; i < rv.NumField(); i++ {
			vField := rv.Field(i)
			tField := rt.Field(i)

			if vField.Kind() != reflect.Ptr {
				vField = vField.Addr()
			}

			value, ok := vField.Interface().(d4.UnmarshalBinary)
			if !ok {
				slog.Warn(
					"invalid field in type",
					slog.String("type", baseTypeName),
					slog.String("field", tField.Name),
					slog.String("fieldType", vField.Type().String()),
				)
				continue
			}

			h.writeFmt(
				`<div class="field"><div class="fieldKey"><span class="fieldName">%s</span><span class="fieldType">%s</span></div>`,
				h.prettyFieldName(tField.Name),
				h.prettyTypeName(tField.Type.String()),
			)
			h.sb.WriteString(`<div class="fieldValue">`)
			h.add(value)
			h.sb.WriteString(`</div></div>`)
		}
		h.sb.WriteString("</div>")
	}
	return
}

func (h *HtmlGenerator) Add(x d4.UnmarshalBinary) {
	h.add(x)
}

func (h *HtmlGenerator) String() string {
	return fmt.Sprintf(
		`<html><head><style type="text/css">%s</style></head><body>%s</body></html>`,
		css,
		h.sb.String(),
	)
}
