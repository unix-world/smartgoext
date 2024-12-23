package glfw

/*
#define GLFW_EXPOSE_NATIVE_COCOA
#define GLFW_EXPOSE_NATIVE_NSGL
#include "glfw/include/GLFW/glfw3.h"
#include "glfw/include/GLFW/glfw3native.h"

// workaround wrappers needed due to a cgo and/or LLVM bug.
// See: glfw/issues/136
void *workaround_glfwGetCocoaWindow(GLFWwindow *w) {
	return (void *)glfwGetCocoaWindow(w);
}
void *workaround_glfwGetNSGLContext(GLFWwindow *w) {
	return (void *)glfwGetNSGLContext(w);
}
*/
import "C"

// GetCocoaMonitor returns the CGDirectDisplayID of the monitor.
func (m *Monitor) GetCocoaMonitor() uintptr {
	ret := uintptr(C.glfwGetCocoaMonitor(m.data))
	panicError()
	return ret
}

// GetCocoaWindow returns the NSWindow of the window.
func (w *Window) GetCocoaWindow() uintptr {
	ret := uintptr(C.workaround_glfwGetCocoaWindow(w.data))
	panicError()
	return ret
}

// GetNSGLContext returns the NSOpenGLContext of the window.
func (w *Window) GetNSGLContext() uintptr {
	ret := uintptr(C.workaround_glfwGetNSGLContext(w.data))
	panicError()
	return ret
}
