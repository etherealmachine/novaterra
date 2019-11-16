#include <material>
precision highp float;

uniform vec2 BrushPosition;
uniform float BrushSize;
uniform int BrushType;
uniform int MouseButton;

in vec2 FragTexcoord;

layout(location = 0) out vec4 FragColor;

void main() {
    float height = texture(MatTexture[0], FragTexcoord).x;
    float e = 1.0/128.0;
    float hL = texture(MatTexture[0], vec2(FragTexcoord.x-e, FragTexcoord.y)).x;
    float hR = texture(MatTexture[0], vec2(FragTexcoord.x+e, FragTexcoord.y)).x;
    float hU = texture(MatTexture[0], vec2(FragTexcoord.x, FragTexcoord.y-e)).x;
    float hD = texture(MatTexture[0], vec2(FragTexcoord.x, FragTexcoord.y+e)).x;
    float waterHeight = texture(MatTexture[0], FragTexcoord).y;
    if (MouseButton == 1 && distance(BrushPosition, FragTexcoord) < BrushSize) {
        switch (BrushType) {
            case 1: // Raise
                height += 0.5;
                break;
            case 2: // Lower
                height -= 0.5;
                break;
            case 3: // Water
                waterHeight += 0.5;
                break;
            case 4: // Smooth
                height = height * 0.5 + 0.5 * ((hL+hR+hU+hD+height) / 5.0);
                break;
            case 5: // Flatten
                height = height * 0.5 + 0.5 * max(height, max(max(hL, hR), max(hU, hD)));
                break;
        }
    }
    FragColor = vec4(height, waterHeight, 0, 0);
}