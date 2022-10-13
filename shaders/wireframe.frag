noperspective in vec3 Vertex;
out vec4 FragColor;

void main()
{
  FragColor.rgb = Vertex;
  FragColor.a = 1.0;
}
