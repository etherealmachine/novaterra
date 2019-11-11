#include <material>
precision highp float;

in vec2 FragTexcoord;

out vec4 FragColor;

void main() {
    float height = texture(MatTexture[0], FragTexcoord).x;
    float waterHeight = texture(MatTexture[0], FragTexcoord).y;
    FragColor = vec4(height/10, waterHeight, 0, 1);
}