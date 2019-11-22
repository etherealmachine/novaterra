#include <material>
precision highp float;

uniform bool EnableErosion;

in vec2 FragTexcoord;

layout(location = 0) out vec4 FragColor;

void main() {
  float height = texture(MatTexture[0], FragTexcoord).x;
  float waterHeight = texture(MatTexture[0], FragTexcoord).y;
  FragColor = vec4(height, waterHeight, 0, 0);
}