#version 460

const uint COLOURS = 170;

in vec2 frag;
out vec3 outputColor;

uniform uint max_iterations;
uniform dvec2 pos;
uniform double zoom;
uniform vec3[COLOURS] colour_pallet;

dvec2 multiply(in dvec2 i, in dvec2 j) {
    return dvec2(i.x * j.x - i.y * j.y, i.x * j.y + i.y * j.x);
}

uint mandelbrot(in dvec2 z_const) {
    uint iterations = 0;
    dvec2 z = z_const;

    while (abs(z.x) + abs(z.y) <= 4 && iterations < max_iterations) {
        z = multiply(z, z) + z_const;
        iterations++;
    }

    return iterations;
}

void main() {
    uint iterations = mandelbrot(frag * zoom - pos);

    if (iterations == max_iterations) {
        outputColor = vec3(0.1,0.1,0.1);
    } else {
        outputColor = colour_pallet[iterations%COLOURS];
    }
}
