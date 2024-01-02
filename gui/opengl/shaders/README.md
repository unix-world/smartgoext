# Shaders

[![GoDoc](https://godoc.org/gitlab.com/vbsw/shaders?status.svg)](https://godoc.org/gitlab.com/vbsw/shaders) [![Go Report Card](https://goreportcard.com/badge/gitlab.com/vbsw/shaders)](https://goreportcard.com/report/gitlab.com/vbsw/shaders) [![Stability: Experimental](https://masterminds.github.io/stability/experimental.svg)](https://masterminds.github.io/stability/experimental.html)

## About
This package contains plain vertex and fragment shaders, version 1.30 (OpenGL 3.0), for Go. It is published on <https://github.com/vbsw/shaders> and <https://gitlab.com/vbsw/shaders>.

## Copyright
Copyright 2020, Vitali Baumtrok (vbsw@mailbox.org).

Shaders is distributed under the Boost Software License, version 1.0. (See accompanying file LICENSE or copy at http://www.boost.org/LICENSE_1_0.txt)

Shaders is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the Boost Software License for more details.

## Examples
With this package:

	import (
		"github.com/go-gl/gl/v3.3-core/gl"
		"gitlab.com/vbsw/shaders"
	)

	func main() {
		....
		shader := shaders.NewPrimitiveShader()
		shader.VertexShaderID = gl.CreateShader(gl.VERTEX_SHADER)
		gl.ShaderSource(shader.VertexShaderID, 1, shader.VertexShader, nil)
		....
	}

Standard way:

	import (
		"github.com/go-gl/gl/v3.3-core/gl"
	)

	func main() {
		....
		vertexShaderStr := "#version 130\nin vec3 vp;\nvoid main () {\n\tgl_Position = vec4(vp, 1.0);\n}"
		shaderID := gl.CreateShader(gl.VERTEX_SHADER)
		vertexShaderC, free := gl.Strs(vertexShaderStr + "\x00")
		gl.ShaderSource(shaderID, 1, vertexShaderC, nil)
		free()
		....
	}

Initialize the projection matrix:

	import (
		"github.com/go-gl/gl/v3.3-core/gl"
		"gitlab.com/vbsw/shaders"
	)

	func main() {
		....
		shader.ProjectionLocation = gl.GetUniformLocation(shader.ProgramID, shader.ProjectionUniform)
		gl.UseProgram(shader.ProgramID)
		setProjection(shader.ProjectionLocation, 400, 400)
		....
	}

	func setProjection(projectionLocation int32, width, height int) {
		projectionMatrix := make([]float32, 4*4)
		projectionMatrix[0] = 2.0 / float32(width)
		projectionMatrix[5] = 2.0 / float32(height)
		projectionMatrix[12] = -1.0
		projectionMatrix[13] = -1.0
		projectionMatrix[15] = 1.0
		gl.UniformMatrix4fv(projectionLocation, 1, false, &projectionMatrix[0]);
	}

## References
- https://golang.org/doc/install
- https://git-scm.com/book/en/v2/Getting-Started-Installing-Git
- https://www.khronos.org/registry/OpenGL/specs/gl/GLSLangSpec.1.30.pdf
- https://github.com/go-gl/gl/blob/master/v3.3-core/gl/conversions.go
- https://git.sr.ht/~eliasnaur/gio/tree/master/example/glfw/main.go
