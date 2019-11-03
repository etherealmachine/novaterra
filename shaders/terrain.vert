#include <attributes>
#include <material>

// Built-in Model uniforms
uniform mat4 ModelViewMatrix;
uniform mat3 NormalMatrix;
uniform mat4 MVP;

out vec2 FragTexcoord;
out float Height;

void main() {
	FragTexcoord = VertexTexcoord;
	Height = texture(MatTexture[0], VertexTexcoord).x;
	vec3 pos = VertexPosition;
	pos.z += Height;
	gl_Position = MVP * vec4(pos, 1.0);
}