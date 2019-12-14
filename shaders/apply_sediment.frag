#include <material>
precision highp float;

uniform bool EnableErosion;

in vec2 FragTexcoord;

layout(location = 0) out vec4 FragColor;

void main() {
  float height = texture(MatTexture[0], FragTexcoord).x;
  float waterHeight = texture(MatTexture[0], FragTexcoord).y;
  vec2 velocity = texture(MatTexture[0], FragTexcoord).zw;
  float erosion = texture(MatTexture[10], FragTexcoord).z;
  float deposition = texture(MatTexture[10], FragTexcoord).w;
  if (EnableErosion) {
    height = max(0, height - erosion + deposition);
  }
  FragColor = vec4(height, waterHeight, velocity.x, velocity.y);
}