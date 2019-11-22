#include <material>
precision highp float;

in vec2 FragTexcoord;

layout(location = 0) out vec4 FragColor;

void main() {
  vec2 velocity = texture(MatTexture[0], FragTexcoord).zw;
  float e = 1.0/128.0;
  float hL = texture(MatTexture[0], vec2(FragTexcoord.x-e, FragTexcoord.y)).x;
  float hR = texture(MatTexture[0], vec2(FragTexcoord.x+e, FragTexcoord.y)).x;
  float hU = texture(MatTexture[0], vec2(FragTexcoord.x, FragTexcoord.y-e)).x;
  float hD = texture(MatTexture[0], vec2(FragTexcoord.x, FragTexcoord.y+e)).x;
  float gx = (hL - hR) / 2;
  float gy = (hU - hD) / 2;
  gx *= gx;
  gy *= gy;
  float sinTheta = sqrt(gx + gy) / sqrt(1 + gx + gy);
  float C = sinTheta + length(velocity);
  float sediment = texture(MatTexture[10], FragTexcoord).x;
  FragColor = vec4(sediment, C, 0, 0);
}