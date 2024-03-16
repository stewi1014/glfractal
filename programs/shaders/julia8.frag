#version 460

const uint COLOURS = 170;
const uint SLIDERS = 5;

in vec2 frag;
out vec3 outputColor;

uniform uint max_iterations;
uniform dvec2 pos;
uniform double zoom;
uniform vec3[COLOURS] colour_pallet;
uniform double[SLIDERS] sliders;

dvec2 multiply(in dvec2 i, in dvec2 j) {
    return dvec2(i.x * j.x - i.y * j.y, i.x * j.y + i.y * j.x);
}

uint julia(in dvec2 z_const) {
    uint iterations = 0;
    dvec2 z = z_const;

    while (abs(z.x) + abs(z.y) <= 4 && iterations < max_iterations) {
        z = multiply(multiply(multiply(z, z), multiply(z, z)), multiply(multiply(z, z), multiply(z, z))) + dvec2(sliders[0]-1.08284,sliders[1]);
        iterations++;
    }

    return iterations;
}

void main() {
    uint iterations = julia(frag * zoom - pos);

    if (iterations == max_iterations) {
        outputColor = vec3(0.1,0.1,0.1);
    } else {
        outputColor = colour_pallet[iterations%COLOURS];
    }
}
