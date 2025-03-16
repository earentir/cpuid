/* C CPUID File Writer */
#include <stdio.h>
#include <stdint.h>

#if defined(__i386__) || defined(__x86_64__)
void cpuid(uint32_t eax_in, uint32_t ecx_in, uint32_t *a, uint32_t *b, uint32_t *c, uint32_t *d) {
    __asm__ __volatile__ (
        "cpuid"
        : "=a" (*a), "=b" (*b), "=c" (*c), "=d" (*d)
        : "a" (eax_in), "c" (ecx_in)
    );
}
#else
void cpuid(uint32_t eax_in, uint32_t ecx_in, uint32_t *a, uint32_t *b, uint32_t *c, uint32_t *d) {
    // For non-x86 systems, return zeros.
    *a = *b = *c = *d = 0;
}
#endif

int main(void) {
    uint32_t a, b, c, d;
    cpuid(0, 0, &a, &b, &c, &d);

    FILE *f = fopen("cpuid.bin", "wb");
    if (!f) {
        perror("fopen");
        return 1;
    }
    fwrite(&a, sizeof(a), 1, f);
    fwrite(&b, sizeof(b), 1, f);
    fwrite(&c, sizeof(c), 1, f);
    fwrite(&d, sizeof(d), 1, f);
    fclose(f);

    return 0;
}
