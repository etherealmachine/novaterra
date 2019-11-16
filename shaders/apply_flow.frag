#include <material>
precision highp float;

in vec2 FragTexcoord;

layout(location = 0) out vec4 FragColor;

void main() {
    float height = texture(MatTexture[0], FragTexcoord).x;
    float waterHeight = max(0, 0.99*texture(MatTexture[0], FragTexcoord).y-0.001);
    FragColor = vec4(height, waterHeight, 0, 0);
}