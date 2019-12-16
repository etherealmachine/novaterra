#include <material>
precision highp float;

uniform float Resolution;

in vec2 FragTexcoord;

layout(location = 0) out vec4 FragColor;

void main() {
  vec2 velocity = texture(MatTexture[0], FragTexcoord).zw;
  float e = 1.0/Resolution;
  float hL = texture(MatTexture[0], vec2(FragTexcoord.x-e, FragTexcoord.y)).x;
  float hR = texture(MatTexture[0], vec2(FragTexcoord.x+e, FragTexcoord.y)).x;
  float hU = texture(MatTexture[0], vec2(FragTexcoord.x, FragTexcoord.y-e)).x;
  float hD = texture(MatTexture[0], vec2(FragTexcoord.x, FragTexcoord.y+e)).x;
  float gx = (hL - hR) / 2;
  float gy = (hU - hD) / 2;
  float sinTheta = sqrt(gx*gx + gy*gy) / sqrt(1 + gx*gx + gy*gy);
  float capacity = sinTheta * length(velocity);
  if (isinf(capacity) || isnan(capacity)) capacity = 0;
  float sediment = texture(MatTexture[10], FragTexcoord).x;
  float erosion = 0;
  float deposition = 0;
  if (capacity > sediment) {
    erosion = capacity - sediment;
  } else if (sediment > capacity) {
    deposition = sediment - capacity;
  }
  erosion *= 0.001;
  deposition *= 0.001;
  FragColor = vec4(sediment + erosion - deposition, capacity, erosion, deposition);
}