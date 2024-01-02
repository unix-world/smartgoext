//          Copyright 2020, Vitali Baumtrok.
// Distributed under the Boost Software License, Version 1.0.
//     (See accompanying file LICENSE or copy at
//        http://www.boost.org/LICENSE_1_0.txt)

// Package opengl-go-example opens a window showing graphics rendered in OpenGL. It is programmed using GLFW.
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
				shader := shaders.NewPrimitiveShader()
				err = initShaderProgram(shader)

				if err == nil {
					defer gl.DeleteShader(shader.VertexShaderID)
					defer gl.DeleteShader(shader.FragmentShaderID)
					defer gl.DeleteProgram(shader.ProgramID)
					vbos := newVBOs(1)
					defer gl.DeleteBuffers(int32(len(vbos)), &vbos[0])
					vaos := newVAOs(1)
					defer gl.DeleteVertexArrays(int32(len(vaos)), &vaos[0])

					bindObjects(shader, vaos, vbos)
					gl.UseProgram(shader.ProgramID)

					// transparency
					// gl.Enable(gl.BLEND);
					// gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);

					// wireframe mode
					// gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

					for !window.ShouldClose() {
						gl.ClearColor(0, 0, 0, 0)
						gl.Clear(gl.COLOR_BUFFER_BIT)

						for _, vao := range vaos {
							gl.BindVertexArray(vao)
							gl.DrawArrays(gl.TRIANGLES, 0, 3)
						}
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
				shader.ColorLocation = gl.GetAttribLocation(shader.ProgramID, shader.ColorAttribute)

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
	shaderID := gl.CreateShader(shaderType)
	gl.ShaderSource(shaderID, 1, shaderSource, nil)
	gl.CompileShader(shaderID)
	err := checkShader(shaderID, gl.COMPILE_STATUS)

	if err != nil {
		gl.DeleteShader(shaderID)
	}
	return shaderID, err
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

func checkShader(shaderID, statusType uint32) error {
	var status int32
	var err error

	gl.GetShaderiv(shaderID, statusType, &status)

	if status == gl.FALSE {
		var length int32
		var infoLog string

		gl.GetShaderiv(shaderID, gl.INFO_LOG_LENGTH, &length)

		if length > 0 {
			infoLogBytes := make([]byte, length)
			gl.GetShaderInfoLog(shaderID, length, nil, &infoLogBytes[0])
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

func bindObjects(shader *shaders.Shader, vaos, vbos []uint32) {
	// x, y, z, r, g, b (one triangle)
	vertices := []float32{
		0.0, 1.0, 0.0, 1.0, 0.0, 0.0, 1.0,
		1.0, -1.0, 0.0, 0.0, 1.0, 0.0, 1.0,
		-1.0, -1.0, 0.0, 0.0, 0.0, 1.0, 1.0,
	}
	gl.BindVertexArray(vaos[0])
	gl.EnableVertexAttribArray(uint32(shader.PositionLocation))
	gl.EnableVertexAttribArray(uint32(shader.ColorLocation))

	gl.BindBuffer(gl.ARRAY_BUFFER, vbos[0])
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)
	// position
	gl.VertexAttribPointer(uint32(shader.PositionLocation), 3, gl.FLOAT, false, 7*4, gl.PtrOffset(0))
	// color
	gl.VertexAttribPointer(uint32(shader.ColorLocation), 4, gl.FLOAT, false, 7*4, gl.PtrOffset(3*4))
}
