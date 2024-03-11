#version 460

const uint MAX_ITERATIONS = 1000u;

in vec2 frag;
out vec3 outputColor;

dvec2 multiply(in dvec2 i, in dvec2 j) {
    return dvec2(i.x * j.x - i.y * j.y, i.x * j.y + i.y * j.x);
}

uint mandelbrot(in dvec2 z_const) {
    uint iterations = 0;
    dvec2 z = z_const;

    while (abs(z.x) + abs(z.y) <= 4 && iterations < MAX_ITERATIONS) {
        z = multiply(z, z) + z_const;
        iterations++;
    }

    return iterations;
}

void main() {
    uint iterations = mandelbrot(frag);

    if (iterations == MAX_ITERATIONS) {
        outputColor = vec3(0.33,0.33,0.33);
    } else {
        outputColor = vec3(0.66,0.66,0.66);
    }
}
