precision highp float;

// Inputs from vertex shader
in vec4 Position;     // Fragment position in camera coordinates
in vec3 Normal;       // Fragment normal in camera coordinates
in vec2 FragTexcoordX; // Fragment texture coordinates
in vec2 FragTexcoordY; // Fragment texture coordinates
in vec2 FragTexcoordZ; // Fragment texture coordinates
in vec3 TexBlend;

#include <lights>
#include <material>
#include <phong_model>

// Final fragment color
out vec4 FragColor;

void main() {

    // Combine material with texture colors
    vec4 matDiffuse1 = texture(MatTexture[0], FragTexcoordX * MatTexRepeat(0) + MatTexOffset(0));
    vec4 matDiffuse2 = texture(MatTexture[1], FragTexcoordY * MatTexRepeat(0) + MatTexOffset(0));
    vec4 matDiffuse3 = texture(MatTexture[2], FragTexcoordZ * MatTexRepeat(0) + MatTexOffset(0));

    vec4 matDiffuse = TexBlend.x * matDiffuse1 + TexBlend.y * matDiffuse2 + TexBlend.z * matDiffuse3;

    vec4 matAmbient = matDiffuse;

    // Normalize interpolated normal as it may have shrinked
    vec3 fragNormal = normalize(Normal);

    // Calculate the direction vector from the fragment to the camera (origin)
    vec3 camDir = normalize(-Position.xyz);

    // Workaround for gl_FrontFacing
    vec3 fdx = dFdx(Position.xyz);
    vec3 fdy = dFdy(Position.xyz);
    vec3 faceNormal = normalize(cross(fdx,fdy));
    if (dot(fragNormal, faceNormal) < 0.0) { // Back-facing
        fragNormal = -fragNormal;
    }

    // Calculates the Ambient+Diffuse and Specular colors for this fragment using the Phong model.
    vec3 Ambdiff, Spec;
    phongModel(Position, fragNormal, camDir, vec3(matAmbient), vec3(matDiffuse), Ambdiff, Spec);

    // Final fragment color
    FragColor = min(vec4(Ambdiff + Spec, matDiffuse.a), vec4(1.0));
}