#include <attributes>
#include <material>

uniform float Resolution;

// Built-in Model uniforms
uniform mat4 ModelViewMatrix;
uniform mat3 NormalMatrix;
uniform mat4 MVP;

uniform vec3 CameraPosition;

out vec4 Position;
out vec3 Normal;
out vec3 CamDir;
out vec2 FragTexcoord;
out float Height;
out float WaterHeight;

float height(vec2 pos) {
	Height = texture(MatTexture[0], pos).x;
	WaterHeight = texture(MatTexture[0], pos).y;
	return Height + WaterHeight;
}

void main() {
  float hL = height(vec2(VertexTexcoord.x - 1.0/Resolution, VertexTexcoord.y));
  float hR = height(vec2(VertexTexcoord.x + 1.0/Resolution, VertexTexcoord.y));
  float hU = height(vec2(VertexTexcoord.x, VertexTexcoord.y - 1.0/Resolution));
  float hD = height(vec2(VertexTexcoord.x, VertexTexcoord.y + 1.0/Resolution));
  Normal.x = hL - hR;
  Normal.y = hD - hU;
  Normal.z = 2.0;
  Normal = normalize(NormalMatrix * Normal);

	FragTexcoord = VertexTexcoord;
	Height = texture(MatTexture[0], VertexTexcoord).x;
	WaterHeight = texture(MatTexture[0], VertexTexcoord).y;
	vec3 pos = VertexPosition;
	pos.z += Height + WaterHeight;
	Position = ModelViewMatrix * vec4(pos, 1.0);
	gl_Position = MVP * vec4(pos, 1.0);
	CamDir = normalize(-(ModelViewMatrix * vec4(CameraPosition, 1.0)).xyz);
}