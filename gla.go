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

func glBool(b bool) C.GLboolean {
	if b {
		return 1
	}
	return 0
}

//
// Texture functions
//

type imageInfo struct {
	Data      unsafe.Pointer
	RowLength int
	Length    int
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
	UNSIGNED_BYTE_3_3_2             = 0x8032
	UNSIGNED_SHORT_4_4_4_4          = 0x8033
	UNSIGNED_SHORT_5_5_5_1          = 0x8034
	UNSIGNED_INT_8_8_8_8            = 0x8035
	UNSIGNED_INT_10_10_10_2         = 0x8036
	UNSIGNED_BYTE_2_3_3_REV         = 0x8362
	UNSIGNED_SHORT_5_6_5            = 0x8363
	UNSIGNED_SHORT_5_6_5_REV        = 0x8364
	UNSIGNED_SHORT_4_4_4_4_REV      = 0x8365
	UNSIGNED_SHORT_1_5_5_5_REV      = 0x8366
	UNSIGNED_INT_8_8_8_8_REV        = 0x8367
	UNSIGNED_INT_2_10_10_10_REV     = 0x8368
)

// Returns GL parameters for loading data from the subrect "r" of image "img"
func getImageInfo(i image.Image) imageInfo {
	var data reflect.Value
	var stride int
	var epp int // elements per pixel
	var info imageInfo

	switch i.(type) {
	case *image.Alpha:
		img, _ := i.(*image.Alpha)
		data, stride, epp = reflect.ValueOf(img.Pix), img.Stride, 1
		info.Format, info.Type = gl.ALPHA, gl.UNSIGNED_BYTE
	case *image.Alpha16:
		img, _ := i.(*image.Alpha16)
		data, stride, epp = reflect.ValueOf(img.Pix), img.Stride, 2
		info.Format, info.Type = gl.ALPHA, gl.UNSIGNED_SHORT
	case *image.Gray:
		img, _ := i.(*image.Gray)
		data, stride, epp = reflect.ValueOf(img.Pix), img.Stride, 1
		info.Format, info.Type = gl.LUMINANCE, gl.UNSIGNED_BYTE
	case *image.Gray16:
		img, _ := i.(*image.Gray16)
		data, stride, epp = reflect.ValueOf(img.Pix), img.Stride, 2
		info.Format, info.Type = gl.LUMINANCE, gl.UNSIGNED_SHORT
	case *image.RGBA:
		img, _ := i.(*image.RGBA)
		data, stride, epp = reflect.ValueOf(img.Pix), img.Stride, 4
		info.Format, info.Type = gl.RGBA, gl.UNSIGNED_BYTE
	case *image.RGBA64:
		img, _ := i.(*image.RGBA64)
		data, stride, epp = reflect.ValueOf(img.Pix), img.Stride, 8
		info.Format, info.Type = gl.RGBA, gl.UNSIGNED_SHORT
	case *glimage.BGRA:
		img, _ := i.(*glimage.BGRA)
		data, stride, epp = reflect.ValueOf(img.Pix), img.Stride, 4
		info.Format, info.Type = gl.BGRA, gl.UNSIGNED_BYTE
	case *glimage.BGRA4444:
		img, _ := i.(*glimage.BGRA4444)
		data, stride, epp = reflect.ValueOf(img.Pix), img.Stride, 1
		info.Format, info.Type = gl.BGRA, UNSIGNED_SHORT_4_4_4_4_REV
	case *glimage.BGRA5551:
		img, _ := i.(*glimage.BGRA5551)
		data, stride, epp = reflect.ValueOf(img.Pix), img.Stride, 1
		info.Format, info.Type = gl.BGRA, UNSIGNED_SHORT_1_5_5_5_REV
	case *glimage.BGR565:
		img, _ := i.(*glimage.BGR565)
		data, stride, epp = reflect.ValueOf(img.Pix), img.Stride, 1
		info.Format, info.Type = gl.RGB, UNSIGNED_SHORT_5_6_5
	default:
		// for unknown types, convert to RGBA8
		r := i.Bounds()
		img := image.NewRGBA(r)
		draw.Draw(img, r.Sub(r.Min), i, r.Min, draw.Src)
		info.Format, info.Type = gl.RGBA, gl.UNSIGNED_BYTE
		info.Data = unsafe.Pointer(reflect.ValueOf(img.Pix).Index(0).UnsafeAddr())
		info.RowLength = img.Stride / 4
		return info
	}

	info.Data = unsafe.Pointer(data.Index(0).UnsafeAddr())
	info.RowLength = stride / epp

	if stride%epp != 0 {
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

	C.glPixelStorei(C.GLenum(gl.UNPACK_ALIGNMENT), C.GLint(1))
	C.glPixelStorei(C.GLenum(gl.UNPACK_ROW_LENGTH), C.GLint(info.RowLength))
	C.glTexImage2D(C.GLenum(target), C.GLint(level), C.GLint(internalformat),
		C.GLsizei(bounds.Dx()), C.GLsizei(bounds.Dy()), C.GLint(border),
		C.GLenum(info.Format), C.GLenum(info.Type),
		info.Data)
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

	C.glPixelStorei(C.GLenum(gl.UNPACK_ALIGNMENT), C.GLint(1))
	C.glPixelStorei(C.GLenum(gl.UNPACK_ROW_LENGTH), C.GLint(info.RowLength))
	C.glTexSubImage2D(C.GLenum(target), C.GLint(level),
		C.GLint(dest.Min.X), C.GLint(dest.Min.Y),
		C.GLsizei(dest.Dx()), C.GLsizei(dest.Dy()),
		C.GLenum(info.Format), C.GLenum(info.Type),
		info.Data)
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
	info.Data = unsafe.Pointer(reflect.ValueOf(data).Index(0).UnsafeAddr())
	info.Length = len(data)

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

	C.glCompressedTexImage2D(C.GLenum(target), C.GLint(level), C.GLenum(info.Format),
		C.GLsizei(bounds.Dx()), C.GLsizei(bounds.Dy()), C.GLint(border),
		C.GLsizei(info.Length), info.Data)
}

//
// Buffer Functions
//

type sliceInfo struct {
	Ptr    uintptr
	Size   uintptr
	Length uintptr
}

func sliceToUintptr(slice interface{}) (data sliceInfo, err error) {
	val := reflect.ValueOf(slice)
	switch val.Type().Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return sliceInfo{}, fmt.Errorf("gla: nil")
		}
		i := reflect.Indirect(val)
		if i.Kind() == reflect.Array {
			e := i.Index(0)
			data.Ptr = e.UnsafeAddr()
			data.Size = e.Type().Size()
			data.Length = uintptr(i.Len())
		} else {
			data.Ptr = i.UnsafeAddr()
			data.Size = i.Type().Size()
			data.Length = 1
		}
	case reflect.Slice:
		if val.IsNil() {
			return sliceInfo{}, fmt.Errorf("gla: nil")
		}
		e := val.Index(0)
		data.Ptr = e.UnsafeAddr()
		data.Size = e.Type().Size()
		data.Length = uintptr(val.Len())
	default:
		return sliceInfo{}, fmt.Errorf("gla: not addressable")
	}

	//fmt.Printf("ptr=%v size=%v length=%v\n", data.Ptr, data.Size, data.Length)
	return data, nil
}

// BufferData uses glBufferData to allocate memory for the bound GL buffer,
// and fills it with the data in slice. slice can be either a slice or pointer.
// Behaviour is undefined if the underlying data type is or contains anything
// other than sized numeric types. (Structs and arrays are okay as long as they're
// of sized numeric types. Pointers and slices are only allowed at the topmost level.)
func BufferData(target GLenum, slice interface{}, usage GLenum) error {
	data, e := sliceToUintptr(slice)
	if e != nil {
		return e
	}
	C.glBufferData(C.GLenum(target), C.GLsizeiptr(data.Size*data.Length), unsafe.Pointer(data.Ptr), C.GLenum(usage))
	return nil
}

// BufferSubData uses glBufferSubData to update a section of the buffer
// currently bound to target. The buffer's type is treated as being an array
// of the same type as slice. Trying to set data off the end of the buffer
// is not checked, and will likely result in a GL error.
//
// This function sets:
//   buffer_bound_to_target[start:start+len(slice)] = slice[:]
func BufferSubData(target GLenum, start_index int, slice interface{}) error {
	data, e := sliceToUintptr(slice)
	if e != nil {
		return e
	}
	C.glBufferSubData(C.GLenum(target), C.GLintptr(uintptr(start_index)*data.Size), C.GLsizeiptr(data.Size*data.Length), unsafe.Pointer(data.Ptr))
	return nil
}

//
// Vertex Attrib Functions
//

func sliceFieldToGL(rtype reflect.Type) (gltype GLenum, err error) {
	// In order for us to be able to bind a Go type as a vertex attribute,
	// it needs to be in this list.
	switch rtype.Kind() {
	case reflect.Int8:
		gltype = gl.BYTE
	case reflect.Int16:
		gltype = gl.SHORT
	case reflect.Int32:
		gltype = gl.INT
	case reflect.Uint8:
		gltype = gl.UNSIGNED_BYTE
	case reflect.Uint16:
		gltype = gl.UNSIGNED_SHORT
	case reflect.Uint32:
		gltype = gl.UNSIGNED_INT
	case reflect.Float32:
		gltype = gl.FLOAT
	case reflect.Float64:
		gltype = gl.DOUBLE
	default:
		return 0, fmt.Errorf("gla: invalid data type in reflection")
	}
	return gltype, nil
}

type attribInfo struct {
	Gltype   GLenum
	Elements int
	Offset   uintptr
	Stride   uintptr
}

func sliceAttrib(dummy interface{}, dummy_index int) (data attribInfo, err error) {
	t := reflect.TypeOf(dummy)
	data.Stride = t.Size()

	switch t.Kind() {
	case reflect.Struct:
		sf := t.Field(int(dummy_index))
		if sf.Type.Kind() == reflect.Array {
			data.Gltype, err = sliceFieldToGL(sf.Type.Elem())
			data.Elements, data.Offset = sf.Type.Len(), sf.Offset
		} else {
			data.Gltype, err = sliceFieldToGL(sf.Type)
			data.Elements, data.Offset = 1, sf.Offset
		}
	case reflect.Array:
		data.Gltype, err = sliceFieldToGL(t.Elem())
		data.Elements, data.Offset = t.Len(), 0
	default:
		data.Gltype, err = sliceFieldToGL(t)
		data.Elements, data.Offset = 1, 0
	}

	if err != nil {
		return attribInfo{}, fmt.Errorf("gla: invalid type")
	}

	//fmt.Printf("type: %v  el: %v  off: %v  str: %v\n", data.Gltype, data.Elements, data.Offset, data.Stride)
	return data, nil
}

// VertexAttribSlice uses glVertexAttribPointer to bind an element of the currently
// bound buffer at the given index. In order to be able to determine the layout of
// the buffer, you need to pass in an uninitialized element of the slice used to fill
// the buffer. If dummy is a struct, the value you pass as dummy_index determines
// which member of dummy to bind at attribute index.
//
// e.g.,
//
//   VertexAttribSlice(1, false, int32(0), 0)
//
// to bind an int32 slice to attrib 1, or,
//
//   VertexAttribSlice(3, true, s{}, 1)
//
// to bind the 1st element in struct s to attrib 3
//
// Precondition: VBO bound to ARRAY_BUFFER target containing data of array type dummy
func VertexAttribSlice(index uint, normalized bool, dummy interface{}, dummy_index int) error {
	data, err := sliceAttrib(dummy, dummy_index)
	if err != nil {
		return err
	}

	C.glVertexAttribPointer(C.GLuint(index), C.GLint(data.Elements), C.GLenum(data.Gltype), glBool(normalized), C.GLsizei(data.Stride), unsafe.Pointer(data.Offset))
	return nil
}

// VertexSlice uses glVertexPointer to bind an element of the currently
// bound buffer at the given index. See VertexAttribSlice for more details.
//
// Precondition: VBO bound to ARRAY_BUFFER target containing data of array type dummy
func VertexSlice(dummy interface{}, dummy_index int) error {
	data, err := sliceAttrib(dummy, dummy_index)
	if err != nil {
		return err
	}

	C.glVertexPointer(C.GLint(data.Elements), C.GLenum(data.Gltype), C.GLsizei(data.Stride), unsafe.Pointer(data.Offset))
	return nil
}

// NormalSlice uses glNormalPointer to bind an element of the currently
// bound buffer at the given index. See VertexAttribSlice for more details.
//
// Precondition: VBO bound to ARRAY_BUFFER target containing data of array type dummy
func NormalSlice(dummy interface{}, dummy_index int) error {
	data, err := sliceAttrib(dummy, dummy_index)
	if err != nil {
		return err
	}
	if data.Elements != 3 {
		return fmt.Errorf("gla: invalid number of elements")
	}

	C.glNormalPointer(C.GLenum(data.Gltype), C.GLsizei(data.Stride), unsafe.Pointer(data.Offset))
	return nil
}

// ColorSlice uses glColorPointer to bind an element of the currently
// bound buffer at the given index. See VertexAttribSlice for more details.
//
// Precondition: VBO bound to ARRAY_BUFFER target containing data of array type dummy
func ColorSlice(dummy interface{}, dummy_index int) error {
	data, err := sliceAttrib(dummy, dummy_index)
	if err != nil {
		return err
	}

	C.glColorPointer(C.GLint(data.Elements), C.GLenum(data.Gltype), C.GLsizei(data.Stride), unsafe.Pointer(data.Offset))
	return nil
}

// TexCoordSlice uses glVertexPointer to bind an element of the currently
// bound buffer at the given index. See VertexAttribSlice for more details.
//
// Precondition: VBO bound to ARRAY_BUFFER target containing data of array type dummy
func TexCoordSlice(dummy interface{}, dummy_index int) error {
	data, err := sliceAttrib(dummy, dummy_index)
	if err != nil {
		return err
	}

	C.glTexCoordPointer(C.GLint(data.Elements), C.GLenum(data.Gltype), C.GLsizei(data.Stride), unsafe.Pointer(data.Offset))
	return nil
}
