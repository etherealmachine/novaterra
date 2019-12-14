#include <attributes>

// Model uniforms
uniform mat4 MVP;

out vec2 FragTexcoord;

void main() {
  FragTexcoord = VertexTexcoord;
  gl_Position = MVP * vec4(VertexPosition, 1.0);
}