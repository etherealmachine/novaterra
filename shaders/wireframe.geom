layout(triangles) in;
layout(triangle_strip, max_vertices=3) out;
 
in vec4 vPosition[3];
in vec3 vNormal[3];
in vec3 vWorldPosition[3];
in vec3 vWorldNormal[3];

out vec4 Position;
out vec3 Normal;
out vec3 WorldPosition;
out vec3 WorldNormal;
 
void main()
{
  vec4 p0 = gl_in[0].gl_Position;
  vec4 p1 = gl_in[1].gl_Position;
  vec4 p2 = gl_in[2].gl_Position;

  gl_Position = p0;
  Position = vPosition[0];
  Normal = vNormal[0];
  WorldPosition = vWorldPosition[0];
  WorldNormal = vWorldNormal[0];
  EmitVertex();

  gl_Position = p1;
  Position = vPosition[1];
  Normal = vNormal[1];
  WorldPosition = vWorldPosition[1];
  WorldNormal = vWorldNormal[1];
  EmitVertex();

  gl_Position = p2;
  Position = vPosition[2];
  Normal = vNormal[2];
  WorldPosition = vWorldPosition[2];
  WorldNormal = vWorldNormal[2];
  EmitVertex();
 
  EndPrimitive();
}