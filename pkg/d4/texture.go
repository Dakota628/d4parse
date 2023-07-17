package d4

import (
	"fmt"
	"github.com/go-gl/gl/v4.6-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"image"
	"os"
	"sync"
)

var (
	TextureFormats = map[int32]TextureFormat{
		0:  {gl.RGBA8, 64, false},
		7:  {gl.R8, 64, false},
		9:  {gl.COMPRESSED_RGB_S3TC_DXT1_EXT, 128, true},
		10: {gl.COMPRESSED_RGB_S3TC_DXT1_EXT, 128, true},
		12: {gl.COMPRESSED_RGBA_S3TC_DXT5_EXT, 64, true},
		23: {gl.R8, 128, false},
		25: {gl.RGBA16F, 64, false},
		41: {gl.COMPRESSED_RED_RGTC1, 64, true},
		42: {gl.COMPRESSED_RG_RGTC2, 64, true},
		43: {gl.COMPRESSED_RGB_BPTC_SIGNED_FLOAT_ARB, 64, true},
		44: {gl.COMPRESSED_RGBA_BPTC_UNORM_ARB, 64, true},
		45: {gl.RGBA8, 64, false},
		46: {gl.COMPRESSED_RGB_S3TC_DXT1_EXT, 64, true},
		47: {gl.COMPRESSED_RGBA_S3TC_DXT1_EXT, 128, true},
		48: {gl.COMPRESSED_RGBA_S3TC_DXT3_EXT, 64, true},
		49: {gl.COMPRESSED_RGBA_S3TC_DXT5_EXT, 64, true},
		50: {gl.COMPRESSED_RGBA_BPTC_UNORM_ARB, 64, true},
		51: {gl.COMPRESSED_RGB_BPTC_UNSIGNED_FLOAT_ARB, 64, true},
	}

	glfwOnce   sync.Once
	glfwWindow *glfw.Window
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

func initGl() *glfw.Window {
	glfwOnce.Do(func() {
		if err := glfw.Init(); err != nil {
			panic(err)
		}

		glfw.WindowHint(glfw.Resizable, glfw.False)
		glfw.WindowHint(glfw.ContextVersionMajor, 4) // OR 2
		glfw.WindowHint(glfw.ContextVersionMinor, 1)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

		window, err := glfw.CreateWindow(1, 1, "d4parse texture", nil, nil)
		if err != nil {
			panic(err)
		}
		window.MakeContextCurrent()

		if err = gl.Init(); err != nil {
			panic(err)
		}
		glfwWindow = window
	})
	return glfwWindow
}

func LoadTexture(def *TextureDefinition, payloadPath string) (image.Image, error) {
	// Init OpenGL
	_ = initGl()

	// Get texture format info
	textureFormat, ok := TextureFormats[def.ETexFormat.Value]
	if !ok {
		return nil, fmt.Errorf("unknown texture format: %d", textureFormat)
	}

	// Get texture pixels
	pixels, err := os.ReadFile(payloadPath)
	if err != nil {
		return nil, err
	}

	// Load texture in OpenGL
	var tex uint32
	gl.GenTextures(1, &tex)
	gl.ActiveTexture(tex)
	gl.BindTexture(gl.TEXTURE_2D, tex)
	defer gl.BindTexture(gl.TEXTURE_2D, 0)

	if textureFormat.Compressed {
		gl.CompressedTexImage2D(
			gl.TEXTURE_2D,
			0,
			uint32(textureFormat.GlInternalFormat),
			int32(def.DwWidth.Value),
			int32(def.DwHeight.Value),
			0,
			int32(len(pixels)),
			gl.Ptr(pixels),
		)
	} else {
		gl.TexImage2D(
			gl.TEXTURE_2D,
			0,
			textureFormat.GlInternalFormat,
			int32(def.DwWidth.Value),
			int32(def.DwHeight.Value),
			0,
			uint32(textureFormat.GlFormat()),
			uint32(textureFormat.GlType()),
			gl.Ptr(pixels),
		)
	}
	gl.GenerateMipmap(gl.TEXTURE_2D)

	rect := image.Rect(0, 0, int(def.DwWidth.Value), int(def.DwHeight.Value))
	rgba := image.NewRGBA(rect)
	gl.GetTexImage(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix),
	)

	return rgba, nil
}
