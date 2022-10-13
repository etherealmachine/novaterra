#include <attributes>

// Model uniforms
uniform mat4 ModelViewMatrix;
uniform mat3 NormalMatrix;
uniform mat4 MVP;

#include <material>

// Output variables for Geometry shader
out vec4 vPosition;
out vec3 vNormal;
out vec3 vWorldPosition;
out vec3 vWorldNormal;

void main() {
    vWorldPosition = VertexPosition;
    vWorldNormal = VertexNormal;
    // Transform vertex position to camera coordinates
    vPosition = ModelViewMatrix * vec4(VertexPosition, 1.0);
    // Transform vertex normal to camera coordinates
    vNormal = normalize(NormalMatrix * VertexNormal);
    // Output projected and transformed vertex position
    gl_Position = MVP * mat4(1.0) * vec4(VertexPosition, 1.0);
}