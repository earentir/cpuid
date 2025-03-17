#include <stdio.h>
#include <conio.h>

/*
 * We declare the external CPUID routine.
 * The routine is expected to be defined in an external assembly file (CPUID.ASM)
 * and linked into the final DOS executable.
 *
 * Parameters:
 *   leaf, subleaf: 32-bit input values
 *   eax, ebx, ecx, edx: pointers to 32-bit variables that receive the CPUID output.
 */
extern void CPUID(long leaf, long subleaf, long *eax, long *ebx, long *ecx, long *edx);

int main(void) {
    FILE *fp;
    unsigned long maxStandard, maxExtended;
    long eax, ebx, ecx, edx;
    int leaf, subleaf;

    fp = fopen("cpuid_data.json", "w");
    if (fp == NULL) {
        printf("Error opening output file.\n");
        return 1;
    }

    fprintf(fp, "{\n  \"entries\": [\n");

    /* --- Capture Standard CPUID Leaves --- */
    /* CPUID(0,0) returns the maximum standard leaf in EAX */
    CPUID(0, 0, (long *)&maxStandard, &ebx, &ecx, &edx);
    for (leaf = 0; leaf <= maxStandard; leaf++) {
        if ((leaf == 4) || (leaf == 0xB) || (leaf == 0xD)) {
            subleaf = 0;
            while (1) {
                CPUID(leaf, subleaf, &eax, &ebx, &ecx, &edx);
                if ((leaf == 4) && (subleaf > 0) && ((eax & 0x1F) == 0))
                    break;
                if ((leaf == 0xB) && (subleaf > 0) && (eax == 0))
                    break;
                if ((leaf == 0xD) && (subleaf > 0) && (eax == 0) && (ebx == 0) && (ecx == 0) && (edx == 0))
                    break;
                fprintf(fp, "    { \"leaf\": %d, \"subleaf\": %d, \"eax\": %ld, \"ebx\": %ld, \"ecx\": %ld, \"edx\": %ld },\n",
                        leaf, subleaf, eax, ebx, ecx, edx);
                subleaf++;
            }
        } else {
            CPUID(leaf, 0, &eax, &ebx, &ecx, &edx);
            fprintf(fp, "    { \"leaf\": %d, \"subleaf\": 0, \"eax\": %ld, \"ebx\": %ld, \"ecx\": %ld, \"edx\": %ld },\n",
                    leaf, eax, ebx, ecx, edx);
        }
    }

    /* --- Capture Extended CPUID Leaves --- */
    CPUID(0x80000000, 0, (long *)&maxExtended, &ebx, &ecx, &edx);
    for (leaf = 0x80000000; leaf <= maxExtended; leaf++) {
        if (leaf == 0x8000001D) {
            subleaf = 0;
            while (1) {
                CPUID(leaf, subleaf, &eax, &ebx, &ecx, &edx);
                if ((subleaf > 0) && ((eax & 0x1F) == 0))
                    break;
                fprintf(fp, "    { \"leaf\": %d, \"subleaf\": %d, \"eax\": %ld, \"ebx\": %ld, \"ecx\": %ld, \"edx\": %ld },\n",
                        leaf, subleaf, eax, ebx, ecx, edx);
                subleaf++;
            }
        } else {
            CPUID(leaf, 0, &eax, &ebx, &ecx, &edx);
            fprintf(fp, "    { \"leaf\": %d, \"subleaf\": 0, \"eax\": %ld, \"ebx\": %ld, \"ecx\": %ld, \"edx\": %ld },\n",
                    leaf, eax, ebx, ecx, edx);
        }
    }

    /* Finalize JSON (the last trailing comma can be cleaned up manually if needed) */
    fprintf(fp, "  ]\n}\n");
    fclose(fp);

    printf("CPUID data captured in cpuid_data.json\n");
    getch();
    return 0;
}
