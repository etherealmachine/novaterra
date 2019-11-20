#include <material>
precision highp float;

in vec2 FragTexcoord;

layout(location = 0) out vec4 FragColor;

void main() {
    float e = 1.0/128.0;

    float h = texture(MatTexture[0], FragTexcoord).x;
    float hL = texture(MatTexture[0], vec2(FragTexcoord.x-e, FragTexcoord.y)).x;
    float hR = texture(MatTexture[0], vec2(FragTexcoord.x+e, FragTexcoord.y)).x;
    float hU = texture(MatTexture[0], vec2(FragTexcoord.x, FragTexcoord.y-e)).x;
    float hD = texture(MatTexture[0], vec2(FragTexcoord.x, FragTexcoord.y+e)).x;

    float w = texture(MatTexture[0], FragTexcoord).y;
    float wL = texture(MatTexture[0], vec2(FragTexcoord.x-e, FragTexcoord.y)).y;
    float wR = texture(MatTexture[0], vec2(FragTexcoord.x+e, FragTexcoord.y)).y;
    float wU = texture(MatTexture[0], vec2(FragTexcoord.x, FragTexcoord.y-e)).y;
    float wD = texture(MatTexture[0], vec2(FragTexcoord.x, FragTexcoord.y+e)).y;

    vec4 flow = texture(MatTexture[8], FragTexcoord);

    flow += vec4(
        max(0, h + w - hL - wL),
        max(0, h + w - hR - wR),
        max(0, h + w - hU - wU),
        max(0, h + w - hD - wD)
    );

    flow *= min(1, w/dot(flow, vec4(1)));

    FragColor = flow;
}