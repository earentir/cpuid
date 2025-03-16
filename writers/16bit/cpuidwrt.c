/* cpuidwrt.c -- Native DOS version (16-bit) */
/* Compile with Turbo C or Borland C */
#include <stdio.h>

/* Declare the external assembly routine.
   The routine is assumed to be named _cpuid16 in the object file.
   It expects a far pointer to a 16-byte buffer.
*/
void cpuid16(char far *buffer);

int main(void) {
    char buffer[16];

    /* Call the assembly routine.
       In 16-bit C, a near call is usually fine when all code is in the same segment.
    */
    cpuid16(buffer);

    /* Write the 16 bytes to a binary file. */
    {
        FILE *f = fopen("cpuid.bin", "wb");
        if (!f) {
            perror("fopen");
            return 1;
        }
        fwrite(buffer, 1, 16, f);
        fclose(f);
    }
    return 0;
}
