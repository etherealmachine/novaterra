#include <attributes>

// Model uniforms
uniform mat4 ModelViewMatrix;
uniform mat3 NormalMatrix;
uniform mat4 MVP;

#include <material>

// Output variables for Fragment shader
out vec4 Position;
out vec3 Normal;
out vec3 WorldPosition;
out vec3 WorldNormal;

void main() {
    WorldPosition = VertexPosition;
    WorldNormal = VertexNormal;
    // Transform vertex position to camera coordinates
    Position = ModelViewMatrix * vec4(VertexPosition, 1.0);
    // Transform vertex normal to camera coordinates
    Normal = normalize(NormalMatrix * VertexNormal);
    // Output projected and transformed vertex position
    gl_Position = MVP * mat4(1.0) * vec4(VertexPosition, 1.0);
}