#version 460

in vec2 vert;
out vec2 frag;

void main() {
    frag = vert;
    gl_Position = vec4(vert, 0 , 1);
}
