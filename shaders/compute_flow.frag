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
    float waterHeight = 0.99 * texture(MatTexture[0], FragTexcoord).y;
    if (waterHeight < 0.1) {
        waterHeight = 0;
    }
    if (MouseButton == 1 && distance(BrushPosition, FragTexcoord) < BrushSize) {
        switch (BrushType) {
            case 1:
                height += 0.5;
                break;
            case 2:
                height -= 0.5;
                break;
            case 3:
                waterHeight += 0.5;
                break;
        }
    }
    FragColor = vec4(height, waterHeight, 0, 1.0);
}