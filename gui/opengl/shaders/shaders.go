/*
 *          Copyright 2020, Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

// Package shaders contains plain vertex and fragment shaders, version 1.30 (OpenGL 3.0).
package shaders

// #include <string.h>
// #include "shaders.h"
import "C"
import (
	"reflect"
	"unsafe"
)

// Shader is for storing shaders and their attributes.
type Shader struct {
	VertexShader       **uint8
	FragmentShader     **uint8
	VertexShaderID     uint32
	FragmentShaderID   uint32
	ProgramID          uint32
	PositionAttribute  *uint8
	ColorAttribute     *uint8
	CoordsAttribute    *uint8
	ProjectionUniform  *uint8
	ModelUniform       *uint8
	TextureUniform     *uint8
	PositionLocation   int32
	ColorLocation      int32
	CoordsLocation     int32
	ProjectionLocation int32
	ModelLocation      int32
	TextureLocation    int32
}

// VertexShaderStr returns vertex shader as string.
func (shader *Shader) VertexShaderStr() string {
	return NewString(*shader.VertexShader)
}

// FragmentShaderStr returns fragment shader as string.
func (shader *Shader) FragmentShaderStr() string {
	return NewString(*shader.FragmentShader)
}

// NewPrimitiveShader returns an instance of Shader for primitive types.
//
// Vertex shader:
//
//	#version 130
//
//	in vec3 positionIn;
//	in vec4 colorIn;
//	out vec4 fragementColor;
//
//	uniform mat4 projection = mat4(1.0);
//	uniform mat4 model = mat4(1.0);
//
//	void main() {
//		gl_Position = projection * model * vec4(positionIn, 1.0f);
//		fragementColor = colorIn;
//	}
//
// Fragment shader:
//
//	#version 130
//
//	in vec4 fragementColor;
//	out vec4 color;
//
//	void main () {
//		color = fragementColor;
//	}
func NewPrimitiveShader() *Shader {
	shader := new(Shader)
	shader.VertexShader = (**uint8)(unsafe.Pointer(&C.vbsw_primitive_vertex_shader))
	shader.FragmentShader = (**uint8)(unsafe.Pointer(&C.vbsw_primitive_fragment_shader))
	shader.PositionAttribute = (*uint8)(unsafe.Pointer(C.vbsw_primitive_position_attribute))
	shader.ColorAttribute = (*uint8)(unsafe.Pointer(C.vbsw_primitive_color_attribute))
	shader.ProjectionUniform = (*uint8)(unsafe.Pointer(C.vbsw_primitive_projection_uniform))
	shader.ModelUniform = (*uint8)(unsafe.Pointer(C.vbsw_primitive_model_uniform))
	return shader
}

// NewTextureShader returns an instance of Shader for textures.
//
// Vertex shader:
//
//	#version 130
//
//	in vec3 positionIn;
//	in vec2 textureCoordsIn;
//	out vec2 fragmentTextureCoords;
//
//	uniform mat4 projection = mat4(1.0);
//	uniform mat4 model = mat4(1.0);
//
//	void main() {
//		gl_Position = projection * model * vec4(positionIn, 1.0f);
//		fragmentTextureCoords = textureCoordsIn;
//	}
//
// Fragment shader:
//
//	#version 130
//
//	in vec2 fragmentTextureCoords;
//	out vec4 color;
//
//	uniform sampler2D imageTexture;
//
//	void main () {
//		color = texture(imageTexture, fragmentTextureCoords);
//	}
func NewTextureShader() *Shader {
	shader := new(Shader)
	shader.VertexShader = (**uint8)(unsafe.Pointer(&C.vbsw_texture_vertex_shader))
	shader.FragmentShader = (**uint8)(unsafe.Pointer(&C.vbsw_texture_fragment_shader))
	shader.PositionAttribute = (*uint8)(unsafe.Pointer(C.vbsw_texture_position_attribute))
	shader.CoordsAttribute = (*uint8)(unsafe.Pointer(C.vbsw_texture_coords_attribute))
	shader.ProjectionUniform = (*uint8)(unsafe.Pointer(C.vbsw_texture_projection_uniform))
	shader.ModelUniform = (*uint8)(unsafe.Pointer(C.vbsw_texture_model_uniform))
	shader.TextureUniform = (*uint8)(unsafe.Pointer(C.vbsw_texture_texture_uniform))
	return shader
}

// NewString converts c-string to string and returns it. Data of c-string is not reallocated,
// but used in string.
func NewString(cstr *uint8) string {
	ccstr := C.vbsw_constchar_t(unsafe.Pointer(cstr))
	n := int(C.strlen(ccstr))
	dataSlice := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(ccstr)),
		Len:  n,
		Cap:  n,
	}))
	str := string(dataSlice)
	return str
}
