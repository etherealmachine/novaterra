#include <attributes>

// Model uniforms
uniform mat4 ModelViewMatrix;
uniform mat3 NormalMatrix;
uniform mat4 MVP;

#include <material>

// Output variables for Fragment shader
out vec4 Position;
out vec3 Normal;
out vec2 FragTexcoordX;
out vec2 FragTexcoordY;
out vec2 FragTexcoordZ;
out vec3 TexBlend;

uniform float delta = 0.5;
uniform float m = 1;

void main() {

    FragTexcoordX.x = VertexPosition.z;
    FragTexcoordY.x = VertexPosition.x;
    FragTexcoordZ.x = -VertexPosition.x;
    if (VertexNormal.x < 0) {
        FragTexcoordX.x = -VertexPosition.z;
    }
    if (VertexNormal.y < 0) {
        FragTexcoordX.x = -VertexPosition.x;
    }
    if (VertexNormal.z < 0) {
        FragTexcoordX.x = VertexPosition.x;
    }
    FragTexcoordX.y = VertexPosition.y;
    FragTexcoordY.y = VertexPosition.z;
    FragTexcoordZ.y = VertexPosition.y;

    float MN = length(VertexNormal);
    TexBlend = normalize(vec3(
        pow(max(VertexNormal.x / MN - delta, 0), m),
        pow(max(VertexNormal.y / MN - delta, 0), m),
        pow(max(VertexNormal.z / MN - delta, 0), m)
    ));

    // Transform vertex position to camera coordinates
    Position = ModelViewMatrix * vec4(VertexPosition, 1.0);

    // Transform vertex normal to camera coordinates
    Normal = normalize(NormalMatrix * VertexNormal);

    vec3 vPosition = VertexPosition;
    mat4 finalWorld = mat4(1.0);

    // Output projected and transformed vertex position
    gl_Position = MVP * finalWorld * vec4(vPosition, 1.0);
}