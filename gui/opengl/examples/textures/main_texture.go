//          Copyright 2020, Vitali Baumtrok.
// Distributed under the Boost Software License, Version 1.0.
//     (See accompanying file LICENSE or copy at
//        http://www.boost.org/LICENSE_1_0.txt)

// +build texture

package main

import (
	"fmt"
	"errors"
	"runtime"

	"github.com/unix-world/smartgoext/gui/opengl/gl/v2.1/gl"
	"github.com/unix-world/smartgoext/gui/opengl/glfw/v3.2/glfw"
	"github.com/unix-world/smartgoext/gui/opengl/shaders"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	err := glfw.Init()

	if err == nil {
		var window *glfw.Window
		defer glfw.Terminate()
		window, err = glfw.CreateWindow(300, 300, "OpenGL Example", nil, nil)

		if err == nil {
			defer window.Destroy()
			window.SetKeyCallback(onKey)
			window.SetSizeCallback(onResize)
			window.MakeContextCurrent()
			err = gl.Init()

			if err == nil {
				shader := shaders.NewTextureShader()
				err = initShaderProgram(shader)

				if err == nil {
					defer gl.DeleteShader(shader.VertexShaderID)
					defer gl.DeleteShader(shader.FragmentShaderID)
					defer gl.DeleteProgram(shader.ProgramID)
					vbos := newVBOs(2)
					defer gl.DeleteBuffers(int32(len(vbos)), &vbos[0])
					vaos := newVAOs(1)
					defer gl.DeleteVertexArrays(int32(len(vaos)), &vaos[0])
					textures := newTextures(1)
					defer gl.DeleteTextures(int32(len(textures)), &textures[0])

					bindObjects(shader, vaos, vbos, textures)
					gl.UseProgram(shader.ProgramID)

					// transparency
					// gl.Enable(gl.BLEND);
					// gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);

					// wireframe mode
					// gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

					for !window.ShouldClose() {
						gl.ClearColor(0, 0, 0, 0)
						gl.Clear(gl.COLOR_BUFFER_BIT)
						gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)

						window.SwapBuffers()
						glfw.PollEvents()
					}
				}
			}
		}
	}
	if err != nil {
		fmt.Println(err.Error())
	}
}

func onKey(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		window.SetShouldClose(true)
	}
}

func onResize(w *glfw.Window, width, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func initShaderProgram(shader *shaders.Shader) error {
	var err error
	shader.VertexShaderID, err = newShader(gl.VERTEX_SHADER, shader.VertexShader)

	if err == nil {
		shader.FragmentShaderID, err = newShader(gl.FRAGMENT_SHADER, shader.FragmentShader)

		if err == nil {
			shader.ProgramID, err = newProgram(shader)

			if err == nil {
				shader.PositionLocation = gl.GetAttribLocation(shader.ProgramID, shader.PositionAttribute)
				shader.CoordsLocation = gl.GetAttribLocation(shader.ProgramID, shader.CoordsAttribute)
				shader.TextureLocation = gl.GetUniformLocation(shader.ProgramID, shader.TextureUniform)

			} else {
				gl.DeleteShader(shader.VertexShaderID)
				gl.DeleteShader(shader.FragmentShaderID)
			}
		} else {
			gl.DeleteShader(shader.VertexShaderID)
		}
	}
	return err
}

func newShader(shaderType uint32, shaderSource **uint8) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	gl.ShaderSource(shader, 1, shaderSource, nil)
	gl.CompileShader(shader)
	err := checkShader(shader, gl.COMPILE_STATUS)

	if err != nil {
		gl.DeleteShader(shader)
	}
	return shader, err
}

func newProgram(shader *shaders.Shader) (uint32, error) {
	program := gl.CreateProgram()
	gl.AttachShader(program, shader.VertexShaderID)
	gl.AttachShader(program, shader.FragmentShaderID)
	gl.LinkProgram(program)
	err := checkProgram(program, gl.LINK_STATUS)

	if err == nil {
		gl.ValidateProgram(program)
		err = checkProgram(program, gl.VALIDATE_STATUS)

		if err != nil {
			gl.DeleteProgram(program)
		}
	}
	return program, err
}

func checkShader(shader, statusType uint32) error {
	var status int32
	var err error

	gl.GetShaderiv(shader, statusType, &status)

	if status == gl.FALSE {
		var length int32
		var infoLog string

		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &length)

		if length > 0 {
			infoLogBytes := make([]byte, length)
			gl.GetShaderInfoLog(shader, length, nil, &infoLogBytes[0])
			infoLog = string(infoLogBytes)
		}
		err = errors.New("shader " + infoLog)
	}
	return err
}

func checkProgram(program, statusType uint32) error {
	var status int32
	var err error

	gl.GetProgramiv(program, statusType, &status)

	if status == gl.FALSE {
		var length int32
		var infoLog string

		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &length)

		if length > 0 {
			infoLogBytes := make([]byte, length)
			gl.GetProgramInfoLog(program, length, nil, &infoLogBytes[0])
			infoLog = string(infoLogBytes)
		}
		err = errors.New("program " + infoLog)
	}
	return err
}

func newVBOs(n int) []uint32 {
	vbos := make([]uint32, n)
	gl.GenBuffers(int32(len(vbos)), &vbos[0])
	return vbos
}

func newVAOs(n int) []uint32 {
	vaos := make([]uint32, n)
	gl.GenVertexArrays(int32(len(vaos)), &vaos[0])
	return vaos
}

func newTextures(n int) []uint32 {
	textures := make([]uint32, n)
	gl.GenTextures(int32(len(textures)), &textures[0])
	return textures
}

func bindObjects(shader *shaders.Shader, vaos, vbos, textures []uint32) {
	textureData := newTextureData()
	// x, y, z, x_tex, y_tex (two triangles)
	vertices := []float32{
		0.5, 0.5, 0.0, 1.0, 1.0,
		0.5, 0.0, 0.0, 1.0, 0.0,
		0.0, 0.5, 0.0, 0.0, 1.0,
		0.0, 0.0, 0.0, 0.0, 0.0,
	}
	gl.BindTexture(gl.TEXTURE_2D, textures[0])
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 64, 64, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(textureData))
	gl.BindBuffer(gl.ARRAY_BUFFER, vbos[0])
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.EnableVertexAttribArray(uint32(shader.PositionLocation))
	gl.EnableVertexAttribArray(uint32(shader.CoordsLocation))
	gl.VertexAttribPointer(uint32(shader.PositionLocation), 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	gl.VertexAttribPointer(uint32(shader.CoordsLocation), 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
}

func newTextureData() []uint8 {
	data := make([]uint8, 64*64*4)
	for i := 0; i < 64*64; i++ {
		offset := i * 4
		if (i/16+i/(16*64))%2 == 0 {
			data[offset] = 255
			data[offset+1] = 255
			data[offset+2] = 255
			data[offset+3] = 255
		}
	}
	return data
}
