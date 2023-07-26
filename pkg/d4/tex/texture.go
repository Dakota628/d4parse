package tex

import (
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/go-gl/gl/v4.1-compatibility/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
	"image"
	"math"
	"os"
	"runtime"
)

var (
	TextureFormats = map[int32]TextureFormat{
		0:  {gl.RGBA8, 64, false},
		7:  {gl.R8, 64, false},
		9:  {gl.COMPRESSED_RGB_S3TC_DXT1_EXT, 128, true},
		10: {gl.COMPRESSED_RGB_S3TC_DXT1_EXT, 128, true},
		12: {gl.COMPRESSED_RGBA_S3TC_DXT5_EXT, 64, true},
		23: {gl.R8, 128, false},
		25: {gl.RGBA16F, 32, false},
		41: {gl.COMPRESSED_RED_RGTC1, 64, true},
		42: {gl.COMPRESSED_RG_RGTC2, 64, true},
		43: {gl.COMPRESSED_RGB_BPTC_SIGNED_FLOAT_ARB, 64, true},
		44: {gl.COMPRESSED_RGBA_BPTC_UNORM_ARB, 64, true},
		45: {gl.RGBA8, 64, false},
		46: {gl.COMPRESSED_RGB_S3TC_DXT1_EXT, 128, true},
		47: {gl.COMPRESSED_RGBA_S3TC_DXT1_EXT, 128, true},
		48: {gl.COMPRESSED_RGBA_S3TC_DXT3_EXT, 64, true},
		49: {gl.COMPRESSED_RGBA_S3TC_DXT5_EXT, 64, true},
		50: {gl.COMPRESSED_RGBA_BPTC_UNORM_ARB, 64, true},
		51: {gl.COMPRESSED_RGB_BPTC_UNSIGNED_FLOAT_ARB, 64, true},
	} // Compression types: BPTC=BC7, BPTC_FLOAT=BC6H, DXT1, DXT3, DXT5, RGTC1, RGTC2
)

type TextureFormat struct {
	GlInternalFormat int32
	Alignment        int
	Compressed       bool
}

func (f TextureFormat) GlFormat() (format int32) {
	gl.GetInternalformativ(gl.TEXTURE_2D, uint32(f.GlInternalFormat), gl.TEXTURE_IMAGE_FORMAT, 1, &format)
	return
}

func (f TextureFormat) GlType() (type_ int32) {
	gl.GetInternalformativ(gl.TEXTURE_2D, uint32(f.GlInternalFormat), gl.TEXTURE_IMAGE_TYPE, 1, &type_)
	return
}

func init() {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
}

func makeWindow() *glfw.Window {
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCompatProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Visible, glfw.False)
	glfw.WindowHint(glfw.OpenGLDebugContext, glfw.True)

	window, err := glfw.CreateWindow(1, 1, "d4parse texture", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	if err = gl.Init(); err != nil {
		panic(err)
	}

	return window
}

func align(n int, alignment int) int {
	if r := n % alignment; r != 0 {
		return n + alignment - r
	}
	return n
}

func LoadTexture(def *d4.TextureDefinition, payloadPath string, paylowPath string, levels ...int) (map[int]image.Image, error) {
	// OpenGL will explode if we don't do this
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// I don't know why we need a new window for each call, but this made things work on Windows...
	window := makeWindow()
	defer window.Destroy()

	// Get texture format info
	textureFormat, ok := TextureFormats[def.ETexFormat.Value]
	if !ok {
		return nil, fmt.Errorf("unknown texture format: %d", textureFormat)
	}

	// Get texture paylods
	payloadData, err := os.ReadFile(payloadPath)
	if err != nil {
		return nil, err
	}

	var paylowData []byte
	if paylowPath != "" {
		if _, err = os.Stat(paylowPath); err == nil {
			paylowData, err = os.ReadFile(paylowPath)
			if err != nil {
				return nil, err
			}
		}
	}

	// Load texture in OpenGL
	var tex uint32
	gl.GenTextures(1, &tex)
	if e := gl.GetError(); e != 0 {
		err = fmt.Errorf("error generating texture: %x", e)
	}

	gl.ActiveTexture(gl.TEXTURE0)
	if e := gl.GetError(); e != 0 {
		err = fmt.Errorf("error activating texture0: %x", e)
	}

	gl.BindTexture(gl.TEXTURE_2D, tex)
	if e := gl.GetError(); e != 0 {
		err = fmt.Errorf("error binding texture: %x", e)
	}

	// Defer unbind texture
	defer func() {
		gl.DeleteTextures(1, &tex)
		if e := gl.GetError(); e != 0 {
			err = fmt.Errorf("error unbinding texture: %x", e)
		}

		gl.BindTexture(gl.TEXTURE_2D, 0)
		if e := gl.GetError(); e != 0 {
			err = fmt.Errorf("error unbinding texture: %x", e)
		}
		gl.Flush()
	}()

	// Don't return directly so defer can update the error output if necessary
	mipMap, err := LoadMipMap(tex, def, textureFormat, payloadData, paylowData, levels...)
	return mipMap, err
}

func LoadMipMap(
	tex uint32,
	def *d4.TextureDefinition,
	texFormat TextureFormat,
	payloadData []byte,
	paylowData []byte,
	levels ...int,
) (map[int]image.Image, error) {
	mipMaps := make(map[int]image.Image, len(def.SerTex.Value))
	minLevel := int(def.DwMipMapLevelMin.Value)
	maxLevel := int(def.DwMipMapLevelMax.Value)

	width := int(def.DwWidth.Value)
	height := int(def.DwHeight.Value)

	seenZeroOffset := false
	usePaylow := false

	for l := minLevel; l <= maxLevel; l++ {
		// Get serialize data
		level := l - minLevel
		if len(levels) > 0 && !slices.Contains(levels, level) {
			continue
		}

		serData := def.SerTex.Value[level]

		// Get current data
		offset := int(serData.DwOffset.Value)
		size := int(serData.DwSizeAndFlags.Value)

		if offset == 0 {
			if !usePaylow && seenZeroOffset {
				usePaylow = true
			}
			seenZeroOffset = true
		}

		var data []byte
		if usePaylow {
			data = paylowData[offset : offset+size]
		} else {
			data = payloadData[offset : offset+size]
		}

		// Get current mipmap dimensions
		scale := 1 / math.Pow(2, float64(level))
		levelWidth := int(float64(width) * scale)
		levelWidthAligned := align(levelWidth, texFormat.Alignment)
		levelHeight := int(float64(height) * scale)

		slog.Info(
			"Loading texture",
			slog.Int("format", int(def.ETexFormat.Value)),
			slog.Int("l", l),
			slog.Int("level", level),
			slog.Int("alignment", texFormat.Alignment),
			slog.Int("levelWidth", levelWidth),
			slog.Int("levelHeight", levelHeight),
			slog.Any("dataLen", len(data)),
		)

		// Load texture into GPU memory
		if texFormat.Compressed {
			gl.CompressedTexImage2D(
				gl.TEXTURE_2D,
				int32(level),
				uint32(texFormat.GlInternalFormat),
				int32(levelWidthAligned),
				int32(levelHeight),
				0,
				int32(size),
				gl.Ptr(data),
			)
			if e := gl.GetError(); e != 0 {
				return nil, fmt.Errorf("error uploading compressed texture for mipmap %d: %x", level, e)
			}
		} else {
			gl.TexImage2D(
				gl.TEXTURE_2D,
				int32(level),
				texFormat.GlInternalFormat,
				int32(levelWidthAligned),
				int32(levelHeight),
				0,
				uint32(texFormat.GlFormat()),
				uint32(texFormat.GlType()),
				gl.Ptr(data),
			)
			if e := gl.GetError(); e != 0 {
				return nil, fmt.Errorf("error uploading texture for mipmap %d: %x", level, e)
			}

			runtime.KeepAlive(data)
		}

		// Load current level dimensions
		var texLevelWidth, texLevelHeight int32
		gl.GetTexLevelParameteriv(gl.TEXTURE_2D, int32(level), gl.TEXTURE_WIDTH, &texLevelWidth)
		if e := gl.GetError(); e != 0 {
			return nil, fmt.Errorf("error retrieving texture width for mipmap %d: %x", level, e)
		}

		gl.GetTexLevelParameteriv(gl.TEXTURE_2D, int32(level), gl.TEXTURE_HEIGHT, &texLevelHeight)
		if e := gl.GetError(); e != 0 {
			return nil, fmt.Errorf("error retrieving texture height for mipmap %d: %x", level, e)
		}

		// Load transformed texture image into CPU memory
		rgba := image.NewRGBA(image.Rect(0, 0, levelWidth, levelHeight))
		gl.GetTextureSubImage(
			tex,
			int32(level),
			0,
			0,
			0,
			int32(levelWidth),
			int32(levelHeight),
			1,
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			int32(len(rgba.Pix)),
			gl.Ptr(rgba.Pix),
		)
		if e := gl.GetError(); e != 0 {
			return nil, fmt.Errorf("error retrieving texture image for mipmap %d: %x", level, e)
		}

		mipMaps[level] = rgba
	}

	runtime.KeepAlive(paylowData)
	runtime.KeepAlive(payloadData)
	gl.Flush()

	return mipMaps, nil
}
