#ifndef SHADERS_H
#define SHADERS_H

#ifdef __cplusplus
extern "C" {
#endif // __cplusplus

typedef const char* vbsw_constchar_t;

extern const char *const vbsw_primitive_vertex_shader;
extern const char *const vbsw_primitive_fragment_shader;
extern const char *const vbsw_primitive_position_attribute;
extern const char *const vbsw_primitive_color_attribute;
extern const char *const vbsw_primitive_projection_uniform;
extern const char *const vbsw_primitive_model_uniform;

extern const char *const vbsw_texture_vertex_shader;
extern const char *const vbsw_texture_fragment_shader;
extern const char *const vbsw_texture_position_attribute;
extern const char *const vbsw_texture_coords_attribute;
extern const char *const vbsw_texture_projection_uniform;
extern const char *const vbsw_texture_model_uniform;
extern const char *const vbsw_texture_texture_uniform;

#ifdef __cplusplus
}
#endif // __cplusplus

#endif /* SHADERS_H */