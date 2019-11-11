#include <attributes>
#include <material>

out vec2 FragTexcoord;

void main() {
    FragTexcoord = VertexTexcoord;
    gl_Position = vec4(VertexPosition, 1.0);
}