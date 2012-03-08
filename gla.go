// Copyright (c) 2012, James Helferty. All rights reserved.
// Use of this source code is governed by a Clear BSD License
// that can be found in the LICENSE file.

// Package gla implements bindings for OpenGL that are meant to be more
// tightly integrated with core Go packages than other bindings. In order
// to accomplish this, gla may at times clobber additional pieces of
// OpenGL state beyond just the one implied by the function name.
// Each function documents which GL state is modified in this manner.

// It is intended that this package integrate as seemlessly as possible
// with existing OpenGL bindings such as github.com/banthar/gl

// Naming of functions contained herein attempts to mimic the verbose
// nature of the D3DX functions while still suggesting the underlying
// OpenGL calls used.

package gla

// #cgo darwin LDFLAGS: -L/opt/local/lib -framework OpenGL -lGLEW
// #cgo windows LDFLAGS: -lglew32 -lopengl32
// #cgo linux LDFLAGS: -lGLEW -lGL
// #cgo darwin CFLAGS: -I/opt/local/include
//
// #include <stdlib.h>
//
// #ifdef __APPLE__
// # include "GL/glew.h"
// #else
// # include <GL/glew.h>
// #endif
//
// #undef GLEW_GET_FUN
// #define GLEW_GET_FUN(x) (*x)
import "C"
import "unsafe"
import "reflect"
import "image"
import "image/draw"
import "github.com/banthar/gl"

type GLenum C.GLenum
type GLbitfield C.GLbitfield
type GLclampf C.GLclampf
type GLclampd C.GLclampd

type Pointer unsafe.Pointer

// those types are left for compatibility reasons
type GLboolean C.GLboolean
type GLbyte C.GLbyte
type GLshort C.GLshort
type GLint C.GLint
type GLsizei C.GLsizei
type GLubyte C.GLubyte
type GLushort C.GLushort
type GLuint C.GLuint
type GLfloat C.GLfloat
type GLdouble C.GLdouble


type imageInfo struct {
	Data		[]uint8
	RowLength	int
	Format		GLenum
	Type		GLenum
	Compressed	bool
}

// Returns GL parameters for loading data from the subrect "r" of image "img"
func getImageInfo(i image.Image) imageInfo {
	var data	[]uint8
	var stride	int
	var bpp		int
	var info	imageInfo

	switch i.(type) {
	case *image.Alpha:
		img, _ := i.(*image.Alpha)
		data, stride, bpp = img.Pix, img.Stride, 1
		info.Format, info.Type, info.Compressed = gl.ALPHA, gl.UNSIGNED_BYTE, false
	case *image.Alpha16:
		img, _ := i.(*image.Alpha16)
		data, stride, bpp = img.Pix, img.Stride, 2
		info.Format, info.Type, info.Compressed = gl.ALPHA, gl.UNSIGNED_SHORT, false
	case *image.Gray:
		img, _ := i.(*image.Gray)
		data, stride, bpp = img.Pix, img.Stride, 1
		info.Format, info.Type, info.Compressed = gl.LUMINANCE, gl.UNSIGNED_BYTE, false
	case *image.Gray16:
		img, _ := i.(*image.Gray16)
		data, stride, bpp = img.Pix, img.Stride, 2
		info.Format, info.Type, info.Compressed = gl.LUMINANCE, gl.UNSIGNED_SHORT, false
	case *image.RGBA:
		img, _ := i.(*image.RGBA)
		data, stride, bpp = img.Pix, img.Stride, 4
		info.Format, info.Type, info.Compressed = gl.RGBA, gl.UNSIGNED_BYTE, false
	case *image.RGBA64:
		img, _ := i.(*image.RGBA64)
		data, stride, bpp = img.Pix, img.Stride, 8
		info.Format, info.Type, info.Compressed = gl.RGBA, gl.UNSIGNED_SHORT, false
	default:
		// for unknown types, convert to RGBA8
		r := i.Bounds()
		img := image.NewRGBA(r)
		draw.Draw(img, r.Sub(r.Min), i, r.Min, draw.Src)
		info.Format, info.Type = gl.RGBA, gl.UNSIGNED_BYTE
		info.Data, info.RowLength, info.Compressed = img.Pix, img.Stride/4, false
		return info
	}

	info.Data = data
	info.RowLength = stride / bpp

	if stride % bpp != 0 {
		panic("gla: stride isn't usable with OpenGL")
	}

	return info
}

// TexImage2DFromImage loads texture data from an image.Image into the currently
// bound GL texture using the glTexImage2D call. If you wish to load only part of
// an image, pass a subimage as the argument.

// Additional state modified: UNPACK_ALIGNMENT, UNPACK_ROW_LENGTH, PIXEL_UNPACK_BUFFER
func TexImage2DFromImage(target GLenum, level int, internalformat int, border int, pixels interface{}) {
	img, b := pixels.(image.Image)
	if b {
		bounds := img.Bounds()
		if bounds.Empty() {
			return
		}
		info := getImageInfo(img)

		C.glBindBuffer(C.GLenum(gl.PIXEL_UNPACK_BUFFER), C.GLuint(0))

		if info.Compressed {
			// TODO
		} else {
			C.glPixelStorei(C.GLenum(gl.UNPACK_ALIGNMENT), C.GLint(1))
			C.glPixelStorei(C.GLenum(gl.UNPACK_ROW_LENGTH), C.GLint(info.RowLength))
			C.glTexImage2D(C.GLenum(target), C.GLint(level), C.GLint(internalformat),
				C.GLsizei(bounds.Dx()), C.GLsizei(bounds.Dy()), C.GLint(border),
				C.GLenum(info.Format), C.GLenum(info.Type),
				unsafe.Pointer(reflect.ValueOf(info.Data).UnsafeAddr()))
		}
	} else {
		panic("gla: invalid interface type; must be an image type")
	}
}

// TexSubImage2DFromImage loads texture data from an image.Image into the currently
// bound GL texture using the glTexSubImage2D call. If you wish to load only part of
// an image, pass a subimage as the argument.

// Additional state modified: UNPACK_ALIGNMENT, UNPACK_ROW_LENGTH, PIXEL_UNPACK_BUFFER
func TexSubImage2DFromImage(target GLenum, level int, dest image.Rectangle, pixels interface{}) {
	img, b := pixels.(image.Image)
	if b {
		bounds := img.Bounds()
		if (dest.Dx() > bounds.Dx() || dest.Dy() > bounds.Dy()) {
			return
		}
		info := getImageInfo(img)

		C.glBindBuffer(C.GLenum(gl.PIXEL_UNPACK_BUFFER), C.GLuint(0))

		if info.Compressed {
			// TODO
		} else {
			C.glPixelStorei(C.GLenum(gl.UNPACK_ALIGNMENT), C.GLint(1))
			C.glPixelStorei(C.GLenum(gl.UNPACK_ROW_LENGTH), C.GLint(info.RowLength))
			C.glTexSubImage2D(C.GLenum(target), C.GLint(level),
				C.GLint(dest.Min.X), C.GLint(dest.Min.Y),
				C.GLsizei(dest.Dx()), C.GLsizei(dest.Dy()),
				C.GLenum(info.Format), C.GLenum(info.Type),
				unsafe.Pointer(reflect.ValueOf(info.Data).UnsafeAddr()))
		}
	} else {
		panic("gla: invalid interface type; must be an image type")
	}
}


