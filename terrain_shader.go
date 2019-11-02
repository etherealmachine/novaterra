package main

const terrainVertexShader = `
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
		Height = texture(MatTexture[0], VertexTexcoord).w;
		vec3 pos = VertexPosition;
		pos.z += Height;
    gl_Position = MVP * vec4(pos, 1.0);
}
`

const terrainFragmentShader = `
#include <material>

precision highp float;

in vec2 FragTexcoord;
in float Height;

out vec4 FragColor;
void main() {
		FragColor = vec4(Height, 0, 0, 1.0);
		//FragColor = texture(MatTexture[0], FragTexcoord);
}
`
