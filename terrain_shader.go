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
	Height = texture(MatTexture[0], VertexTexcoord).x;
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
	vec4 c1, c2;
	float stop1, stop2, lerp;
	if (Height < 0.1) {
		c1 = texture(MatTexture[1], FragTexcoord * MatTexRepeat(1) + MatTexOffset(1));
		c2 = texture(MatTexture[2], FragTexcoord * MatTexRepeat(2) + MatTexOffset(2));
		stop1 = 0;
		stop2 = 0.1;
	} else if (Height < 1) {
		c1 = texture(MatTexture[2], FragTexcoord * MatTexRepeat(2) + MatTexOffset(2));
		c2 = texture(MatTexture[3], FragTexcoord * MatTexRepeat(3) + MatTexOffset(3));
		stop1 = 0.1;
		stop2 = 1;
	} else if (Height < 10) {
		c1 = texture(MatTexture[3], FragTexcoord * MatTexRepeat(3) + MatTexOffset(3));
		c2 = texture(MatTexture[4], FragTexcoord * MatTexRepeat(4) + MatTexOffset(4));
		stop1 = 1;
		stop2 = 10;
	} else if (Height < 25) {
		c1 = texture(MatTexture[4], FragTexcoord * MatTexRepeat(4) + MatTexOffset(4));
		c2 = texture(MatTexture[5], FragTexcoord * MatTexRepeat(5) + MatTexOffset(5));
		stop1 = 10;
		stop2 = 25;
	} else if (Height < 30) {
		c1 = texture(MatTexture[5], FragTexcoord * MatTexRepeat(5) + MatTexOffset(5));
		c2 = texture(MatTexture[6], FragTexcoord * MatTexRepeat(6) + MatTexOffset(6));
		stop1 = 25;
		stop2 = 30;
	} else {
		c1 = texture(MatTexture[6], FragTexcoord * MatTexRepeat(6) + MatTexOffset(6));
		c2 = texture(MatTexture[6], FragTexcoord * MatTexRepeat(6) + MatTexOffset(6));
		stop1 = 30;
		stop2 = 1000;
	}
	lerp = (Height - stop1) / (stop2 - stop1);
	FragColor = c1 * (1 - lerp) + c2 * lerp;
}
`
