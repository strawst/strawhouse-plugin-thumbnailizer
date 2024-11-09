#include <cuda_runtime.h>
#include <stdint.h>

extern "C" {

__device__ float cubicWeight(float t) {
    t = fabsf(t);
    if (t <= 1) {
        return 1.5f * t * t * t - 2.5f * t * t + 1;
    } else if (t <= 2) {
        return -0.5f * t * t * t + 2.5f * t * t - 4.0f * t + 2;
    }
    return 0;
}

__device__ void bicubicInterpolation(uint8_t* src, int srcWidth, int srcHeight,
                                      int targetWidth, int targetHeight,
                                      uint8_t* dst) {
    float scaleX = static_cast<float>(srcWidth) / targetWidth;
    float scaleY = static_cast<float>(srcHeight) / targetHeight;

    int x = blockIdx.x * blockDim.x + threadIdx.x;
    int y = blockIdx.y * blockDim.y + threadIdx.y;

    if (x < targetWidth && y < targetHeight) {
        float srcX = x * scaleX;
        float srcY = y * scaleY;

        int x1 = static_cast<int>(srcX);
        int y1 = static_cast<int>(srcY);

        // Get surrounding pixels (4x4)
        float r = 0, g = 0, b = 0, a = 0;
        for (int i = -1; i <= 2; i++) {
            for (int j = -1; j <= 2; j++) {
                int px = min(max(x1 + i, 0), srcWidth - 1);
                int py = min(max(y1 + j, 0), srcHeight - 1);
                uint8_t* pixel = src + (py * srcWidth + px) * 4;

                float weight = cubicWeight(srcX - (x1 + i)) * cubicWeight(srcY - (y1 + j));
                r += weight * pixel[0];
                g += weight * pixel[1];
                b += weight * pixel[2];
                a += weight * pixel[3];
            }
        }

        uint8_t* outputPixel = dst + (y * targetWidth + x) * 4;
        outputPixel[0] = static_cast<uint8_t>(clamp(r));
        outputPixel[1] = static_cast<uint8_t>(clamp(g));
        outputPixel[2] = static_cast<uint8_t>(clamp(b));
        outputPixel[3] = static_cast<uint8_t>(clamp(a));
    }
}

__device__ int clamp(float value) {
    return (value < 0) ? 0 : (value > 255) ? 255 : static_cast<int>(value);
}

}
