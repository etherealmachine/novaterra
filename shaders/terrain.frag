precision highp float;

// Inputs from vertex shader
in vec4 Position;      // Fragment position in camera coordinates
in vec3 Normal;        // Fragment normal in camera coordinates
in vec3 WorldPosition; // Fragment position in world coordinates
in vec3 WorldNormal;   // Fragment normal in world coordinates
noperspective in vec3 BaryCoord; // Barycentric coordinate of triangle for wireframe shading
uniform int Mode;

#include <lights>
#include <material>
#include <phong_model>

// Final fragment color
out vec4 FragColor;

vec3 triplanarBlend(vec3 worldNormal) {
	vec3 blend = abs(worldNormal);
	blend = normalize(max(blend, 0.00001));
	float b = (blend.x + blend.y + blend.z);
	blend /= vec3(b, b, b);
	return blend;
}

vec3 triplanarMapping(sampler2D tex, vec3 normal, vec3 position) {
  vec3 normalBlend = triplanarBlend(normal);
  vec3 xColor = texture(tex, position.yz).rgb;
	vec3 yColor = texture(tex, position.xz).rgb;
	vec3 zColor = texture(tex, position.xy).rgb;
    return (xColor * normalBlend.x + yColor * normalBlend.y + zColor * normalBlend.z);
}

void main()
{
  vec3 matDiffuse = triplanarMapping(MatTexture[0], WorldNormal, WorldPosition);

  vec3 matAmbient = matDiffuse;

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
  phongModel(Position, fragNormal, camDir, matAmbient, matDiffuse, Ambdiff, Spec);

  // Final fragment color
  FragColor = vec4(Ambdiff + Spec, 1.0);

  if (Mode == 1 || Mode == 2) {
    if (BaryCoord.x + BaryCoord.y > 0.99 ||
        BaryCoord.x + BaryCoord.z > 0.99 ||
        BaryCoord.y + BaryCoord.z > 0.99) {
      FragColor = vec4(1, 1, 1, 1);
    } else if (Mode == 2) {
      FragColor = vec4(0, 0, 0, 0);
    }
  }
}