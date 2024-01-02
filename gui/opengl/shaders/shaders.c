/*
 *          Copyright 2020, Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

#include <string.h>

const char *const vbsw_primitive_vertex_shader = "#version 130\n\nin vec3 positionIn;\nin vec4 colorIn;\nout vec4 fragementColor;\n\nuniform mat4 projection = mat4(1.0);\nuniform mat4 model = mat4(1.0);\n\nvoid main() {\n\tgl_Position = projection * model * vec4(positionIn, 1.0f);\n\tfragementColor = colorIn;\n}";
const char *const vbsw_primitive_fragment_shader = "#version 130\n\nin vec4 fragementColor;\nout vec4 color;\n\nvoid main() {\n\tcolor = fragementColor;\n}";
const char *const vbsw_primitive_position_attribute = "positionIn";
const char *const vbsw_primitive_color_attribute = "colorIn";
const char *const vbsw_primitive_projection_uniform = "projection";
const char *const vbsw_primitive_model_uniform = "model";

const char *const vbsw_texture_vertex_shader = "#version 130\n\nin vec3 positionIn;\nin vec2 textureCoordsIn;\nout vec2 fragmentTextureCoords;\n\nuniform mat4 projection = mat4(1.0);\nuniform mat4 model = mat4(1.0);\n\nvoid main() {\n\tgl_Position = projection * model * vec4(positionIn, 1.0f);\n\tfragmentTextureCoords = textureCoordsIn;\n}";
const char *const vbsw_texture_fragment_shader = "#version 130\n\nin vec2 fragmentTextureCoords;\nout vec4 color;\n\nuniform sampler2D imageTexture;\n\nvoid main() {\n\tcolor = texture(imageTexture, fragmentTextureCoords);\n}";
const char *const vbsw_texture_position_attribute = "positionIn";
const char *const vbsw_texture_coords_attribute = "textureCoordsIn";
const char *const vbsw_texture_projection_uniform = "projection";
const char *const vbsw_texture_model_uniform = "model";
const char *const vbsw_texture_texture_uniform = "imageTexture";
