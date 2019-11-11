#include <attributes>
#include <material>

// Output data ; will be interpolated for each fragment.
out float Height;
out float WaterHeight;
out vec2 FragTexcoord;

void main(){
	FragTexcoord = VertexTexcoord;
	Height = texture(MatTexture[0], VertexTexcoord).x;
	WaterHeight = texture(MatTexture[0], VertexTexcoord).y;
	gl_Position = vec4(VertexPosition, 1);
}