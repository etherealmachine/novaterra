#include <material>
precision highp float;

in vec2 FragTexcoord;

layout(location = 0) out vec4 FragColor;

void main() {
  float e = 1.0/128.0;
  vec2 velocity = texture(MatTexture[0], FragTexcoord).zw;
  float sediment = texture(MatTexture[10], vec2(FragTexcoord.x-velocity.x, FragTexcoord.y-velocity.y)).x;
  if (FragTexcoord.x - velocity.x <= 0 || FragTexcoord.x - velocity.x >= 1-e || FragTexcoord.y - velocity.y <= e || FragTexcoord.y - velocity.y >= 1-e) {
    sediment = (
      texture(MatTexture[10], vec2(FragTexcoord.x-e, FragTexcoord.y)).x+ 
      texture(MatTexture[10], vec2(FragTexcoord.x+e, FragTexcoord.y)).x+ 
      texture(MatTexture[10], vec2(FragTexcoord.x, FragTexcoord.y-e)).x+ 
      texture(MatTexture[10], vec2(FragTexcoord.x, FragTexcoord.y+e)).x
    ) / 4;
  }
  float capacity = texture(MatTexture[10], FragTexcoord).y;
  float erosion = texture(MatTexture[10], FragTexcoord).z;
  float deposition = texture(MatTexture[10], FragTexcoord).w;
  FragColor = vec4(sediment, capacity, erosion, deposition);
}