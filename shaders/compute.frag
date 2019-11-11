precision highp float;

in vec2 FragTexcoord;
in float Height;
in float WaterHeight;

out vec4 FragColor;

void main() {
	FragColor = vec4(Height, WaterHeight*0.99, 0.0, 0.0);
	FragColor = vec4(FragTexcoord.x, FragTexcoord.y, 0.0, 1.0);
}