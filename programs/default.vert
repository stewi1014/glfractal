#version 460

in vec2 vert;
out vec2 frag;

uniform mat4 camera;

void main() {
    frag = vert;
    gl_Position = camera * vec4(vert, 0 , 1);
}
