// Copyright (c) 2012, James Helferty. All rights reserved.
// Use of this source code is governed by a Clear BSD License
// that can be found in the LICENSE file.

// Package gla implements bindings for OpenGL that are meant to be more
// tightly integrated with core Go packages than other bindings. In order
// to accomplish this, gla may at times clobber additional pieces of
// OpenGL state beyond just the one implied by the function name.
// Each function documents which GL state is modified in this manner.
//
// It is intended that this package integrate as seemlessly as possible
// with existing OpenGL bindings such as github.com/banthar/gl
//
// Naming of functions contained herein attempts to mimic the verbose
// nature of the D3DX functions while still suggesting the underlying
// OpenGL calls used.
//
// (P.S. - If something isn't working, check if you forgot to call gl.Init)

package gla

// #cgo darwin LDFLAGS: -framework OpenGL
// #cgo darwin pkg-config: glew
// #cgo windows LDFLAGS: -lglew32 -lopengl32
// #cgo linux LDFLAGS: -lGLEW -lGL
//
// #include <stdlib.h>
//
// #ifdef __APPLE__
// # include "glew.h"
// #else
// # include <GL/glew.h>
// #endif
//
// #undef GLEW_GET_FUN
// #define GLEW_GET_FUN(x) (*x)
import "C"
import "unsafe"
import "reflect"
import "fmt"
import "image"
import "image/draw"
import "github.com/banthar/gl"
import "github.com/spate/glimage"

type GLenum gl.GLenum
type GLbitfield gl.GLbitfield
type GLclampf gl.GLclampf
type GLclampd gl.GLclampd

type Pointer unsafe.Pointer

// those types are left for compatibility reasons
type GLboolean gl.GLboolean
type GLbyte gl.GLbyte
type GLshort gl.GLshort
type GLint gl.GLint
type GLsizei gl.GLsizei
type GLubyte gl.GLubyte
type GLushort gl.GLushort
type GLuint gl.GLuint
type GLfloat gl.GLfloat
type GLdouble gl.GLdouble

type imageInfo struct {
	Data      []uint8
	RowLength int
	Format    GLenum
	Type      GLenum
}

// The following are missing from github.com/banthar/gl
const (
	COMPRESSED_RGB_S3TC_DXT1        = 0x83F0
	COMPRESSED_RGBA_S3TC_DXT1       = 0x83F1
	COMPRESSED_RGBA_S3TC_DXT3       = 0x83F2
	COMPRESSED_RGBA_S3TC_DXT5       = 0x83F3
	COMPRESSED_SRGB_S3TC_DXT1       = 0x8C4C
	COMPRESSED_SRGB_ALPHA_S3TC_DXT1 = 0x8C4D
	COMPRESSED_SRGB_ALPHA_S3TC_DXT3 = 0x8C4E
	COMPRESSED_SRGB_ALPHA_S3TC_DXT5 = 0x8C4F
)

// Returns GL parameters for loading data from the subrect "r" of image "img"
func getImageInfo(i image.Image) imageInfo {
	var data []uint8
	var stride int
	var bpp int
	var info imageInfo

	switch i.(type) {
	case *image.Alpha:
		img, _ := i.(*image.Alpha)
		data, stride, bpp = img.Pix, img.Stride, 1
		info.Format, info.Type = gl.ALPHA, gl.UNSIGNED_BYTE
	case *image.Alpha16:
		img, _ := i.(*image.Alpha16)
		data, stride, bpp = img.Pix, img.Stride, 2
		info.Format, info.Type = gl.ALPHA, gl.UNSIGNED_SHORT
	case *image.Gray:
		img, _ := i.(*image.Gray)
		data, stride, bpp = img.Pix, img.Stride, 1
		info.Format, info.Type = gl.LUMINANCE, gl.UNSIGNED_BYTE
	case *image.Gray16:
		img, _ := i.(*image.Gray16)
		data, stride, bpp = img.Pix, img.Stride, 2
		info.Format, info.Type = gl.LUMINANCE, gl.UNSIGNED_SHORT
	case *image.RGBA:
		img, _ := i.(*image.RGBA)
		data, stride, bpp = img.Pix, img.Stride, 4
		info.Format, info.Type = gl.RGBA, gl.UNSIGNED_BYTE
	case *image.RGBA64:
		img, _ := i.(*image.RGBA64)
		data, stride, bpp = img.Pix, img.Stride, 8
		info.Format, info.Type = gl.RGBA, gl.UNSIGNED_SHORT
	default:
		// for unknown types, convert to RGBA8
		r := i.Bounds()
		img := image.NewRGBA(r)
		draw.Draw(img, r.Sub(r.Min), i, r.Min, draw.Src)
		info.Format, info.Type = gl.RGBA, gl.UNSIGNED_BYTE
		info.Data, info.RowLength = img.Pix, img.Stride/4
		return info
	}

	info.Data = data
	info.RowLength = stride / bpp

	if stride%bpp != 0 {
		panic("gla: stride isn't usable with OpenGL")
	}

	return info
}

// TexImage2DFromImage loads texture data from an image.Image into the currently
// bound GL texture using the glTexImage2D call. If you wish to load only part of
// an image, pass a subimage as the argument.
//
// Precondition: no buffer object bound to PIXEL_UNPACK_BUFFER
//
// Additional state modified: UNPACK_ALIGNMENT, UNPACK_ROW_LENGTH
func TexImage2DFromImage(target GLenum, level int, internalformat int, border int, img image.Image) {
	bounds := img.Bounds()
	if bounds.Empty() {
		return
	}

	info := getImageInfo(img)
	ptr := unsafe.Pointer(reflect.ValueOf(info.Data).Index(0).UnsafeAddr())

	C.glPixelStorei(C.GLenum(gl.UNPACK_ALIGNMENT), C.GLint(1))
	C.glPixelStorei(C.GLenum(gl.UNPACK_ROW_LENGTH), C.GLint(info.RowLength))
	C.glTexImage2D(C.GLenum(target), C.GLint(level), C.GLint(internalformat),
		C.GLsizei(bounds.Dx()), C.GLsizei(bounds.Dy()), C.GLint(border),
		C.GLenum(info.Format), C.GLenum(info.Type),
		ptr)
}

// TexSubImage2DFromImage loads texture data from an image.Image into the currently
// bound GL texture using the glTexSubImage2D call. If you wish to load only part of
// an image, pass a subimage as the argument.
//
// Precondition: no buffer object bound to PIXEL_UNPACK_BUFFER
//
// Additional state modified: UNPACK_ALIGNMENT, UNPACK_ROW_LENGTH
func TexSubImage2DFromImage(target GLenum, level int, dest image.Rectangle, img image.Image) {
	bounds := img.Bounds()
	if dest.Dx() > bounds.Dx() || dest.Dy() > bounds.Dy() {
		return
	}

	info := getImageInfo(img)
	ptr := unsafe.Pointer(reflect.ValueOf(info.Data).Index(0).UnsafeAddr())

	C.glPixelStorei(C.GLenum(gl.UNPACK_ALIGNMENT), C.GLint(1))
	C.glPixelStorei(C.GLenum(gl.UNPACK_ROW_LENGTH), C.GLint(info.RowLength))
	C.glTexSubImage2D(C.GLenum(target), C.GLint(level),
		C.GLint(dest.Min.X), C.GLint(dest.Min.Y),
		C.GLsizei(dest.Dx()), C.GLsizei(dest.Dy()),
		C.GLenum(info.Format), C.GLenum(info.Type),
		ptr)
}

// Returns GL parameters for loading data from the subrect "r" of image "img"
func getCompressedImageInfo(i image.Image) (imageInfo, error) {
	var data []uint8
	var stride int
	var blocksize int // overall number of bytes in a block
	var blockdim int  // size of one dimension; assume square blocks
	var info imageInfo

	switch i.(type) {
	case *glimage.Dxt1:
		img, _ := i.(*glimage.Dxt1)
		data, stride, blockdim, blocksize = img.Pix, img.Stride, 4, 8
		info.Format = COMPRESSED_RGBA_S3TC_DXT1
	case *glimage.Dxt3:
		img, _ := i.(*glimage.Dxt3)
		data, stride, blockdim, blocksize = img.Pix, img.Stride, 4, 16
		info.Format = COMPRESSED_RGBA_S3TC_DXT3
	case *glimage.Dxt5:
		img, _ := i.(*glimage.Dxt5)
		data, stride, blockdim, blocksize = img.Pix, img.Stride, 4, 16
		info.Format = COMPRESSED_RGBA_S3TC_DXT5
	default:
		return imageInfo{}, fmt.Errorf("gla: unrecognized texture format")
	}

	bounds := i.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	sub_stride := (w + (blockdim - 1)) / blockdim * blocksize // stride of just the data we want
	if sub_stride != stride {
		//fmt.Printf("stride mismatch: %v != %v\n", sub_stride, stride)
		data_xoff := bounds.Min.X + (blockdim-1)/blockdim*blocksize
		data_yoff := bounds.Min.Y + (blockdim-1)/blockdim*stride
		data_offset := data_yoff + data_xoff

		// need to allocate a new chunk of memory and copy the data in
		// because compressed loads don't have pixelstore state :(
		pix := make([]uint8, ((w+(blockdim-1))/blockdim)*((h+(blockdim-1))/blockdim)*blocksize)

		var c int
		for i := 0; i < h/blockdim; i++ {
			c = copy(pix[i*sub_stride:(i+1)*sub_stride], data[data_offset+i*stride:data_offset+(i+1)*stride])
			if c != sub_stride {
				return imageInfo{}, fmt.Errorf("gla: cannot copy subimage")
			}
		}
		data = pix
	}
	info.Data = data

	return info, nil
}

// CompressedTexImage2DFromImage loads texture data from an image.Image into the
// currently bound GL texture using the glCompressedTexImage2D call.
//
// Precondition: no buffer object bound to PIXEL_UNPACK_BUFFER
func CompressedTexImage2DFromImage(target GLenum, level int, border int, img image.Image) {
	bounds := img.Bounds()
	if bounds.Empty() {
		return
	}

	info, err := getCompressedImageInfo(img)
	if err != nil {
		return
	}

	ptr := unsafe.Pointer(reflect.ValueOf(info.Data).Index(0).UnsafeAddr())

	C.glCompressedTexImage2D(C.GLenum(target), C.GLint(level), C.GLenum(info.Format),
		C.GLsizei(bounds.Dx()), C.GLsizei(bounds.Dy()), C.GLint(border),
		C.GLsizei(len(info.Data)), ptr)
}
