#version 460

const uint MAX_ITERATIONS = 1000u;

in vec2 fragCoord;
out vec3 outputColor;

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
    uint iterations = mandelbrot();

    if iterations == MAX_ITERATIONS {
        outputColor = vec3(0,0,0);
    } else {
        outputColor = vec3(1,1,1);
    }
}