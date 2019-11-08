#include <attributes>
#include <material>

// Output data ; will be interpolated for each fragment.
out vec2 FragTexcoord;

void main(){
	FragTexcoord = VertexTexcoord;
	gl_Position = vec4(VertexPosition ,1);
}