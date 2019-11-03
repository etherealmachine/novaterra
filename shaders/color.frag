#include <material>

precision highp float;

in vec2 FragTexcoord;

out vec4 FragColor;
void main() {
	FragColor = vec4(FragTexcoord.x, FragTexcoord.y, 0.0, 1.0);
}