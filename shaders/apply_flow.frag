#include <material>
precision highp float;

in vec2 FragTexcoord;

layout(location = 0) out vec4 FragColor;

void main() {
    float e = 1.0/128.0;

    float height = texture(MatTexture[0], FragTexcoord).x;
    float waterHeight = texture(MatTexture[0], FragTexcoord).y;
    vec4 flowOut = texture(MatTexture[8], FragTexcoord);
    vec4 flowIn = vec4(
        texture(MatTexture[8], vec2(FragTexcoord.x-e, FragTexcoord.y)).y,
        texture(MatTexture[8], vec2(FragTexcoord.x+e, FragTexcoord.y)).x,
        texture(MatTexture[8], vec2(FragTexcoord.x, FragTexcoord.y-e)).w,
        texture(MatTexture[8], vec2(FragTexcoord.x, FragTexcoord.y+e)).z
    );
    float deltaV = dot(flowIn, vec4(1)) - dot(flowOut, vec4(1));
    FragColor = vec4(height, waterHeight + 0.1*deltaV, 0, 0);
}