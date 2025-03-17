#include <stdio.h>
#include <stdint.h>
#include <cpuid.h>
#include <stdlib.h>

typedef struct {
    uint32_t leaf;
    uint32_t subleaf;
    uint32_t eax;
    uint32_t ebx;
    uint32_t ecx;
    uint32_t edx;
} CPUIDEntry;

#define INITIAL_ENTRIES 64

int main(void) {
    CPUIDEntry *entries = NULL;
    size_t count = 0, capacity = INITIAL_ENTRIES;
    FILE *fp;
    uint32_t maxStandard, maxExtended;
    uint32_t eax, ebx, ecx, edx;
    uint32_t leaf, subleaf;

    entries = malloc(capacity * sizeof(CPUIDEntry));
    if (!entries) {
        fprintf(stderr, "Memory allocation error.\n");
        return 1;
    }

    #define APPEND_ENTRY(LF, SL, A, B, C, D) do {         \
        if(count >= capacity) {                           \
            capacity *= 2;                                \
            entries = realloc(entries, capacity * sizeof(CPUIDEntry)); \
            if(!entries) {                                \
                fprintf(stderr, "Memory reallocation error.\n");       \
                exit(1);                                  \
            }                                             \
        }                                                 \
        entries[count].leaf = (LF);                       \
        entries[count].subleaf = (SL);                    \
        entries[count].eax = (A);                         \
        entries[count].ebx = (B);                         \
        entries[count].ecx = (C);                         \
        entries[count].edx = (D);                         \
        count++;                                          \
    } while(0)

    /* --- Capture Standard CPUID Leaves --- */
    if (!__get_cpuid(0, &maxStandard, &ebx, &ecx, &edx)) {
        fprintf(stderr, "CPUID not supported.\n");
        free(entries);
        return 1;
    }
    for (leaf = 0; leaf <= maxStandard; leaf++) {
        if (leaf == 4 || leaf == 0xB || leaf == 0xD) {
            subleaf = 0;
            while (1) {
                __get_cpuid_count(leaf, subleaf, &eax, &ebx, &ecx, &edx);
                if (leaf == 4 && subleaf > 0 && ((eax & 0x1F) == 0))
                    break;
                if (leaf == 0xB && subleaf > 0 && (eax == 0))
                    break;
                if (leaf == 0xD && subleaf > 0 && (eax==0 && ebx==0 && ecx==0 && edx==0))
                    break;
                APPEND_ENTRY(leaf, subleaf, eax, ebx, ecx, edx);
                subleaf++;
            }
        } else {
            __get_cpuid(leaf, &eax, &ebx, &ecx, &edx);
            APPEND_ENTRY(leaf, 0, eax, ebx, ecx, edx);
        }
    }

    /* --- Capture Extended CPUID Leaves --- */
    __get_cpuid(0x80000000, &maxExtended, &ebx, &ecx, &edx);
    for (leaf = 0x80000000; leaf <= maxExtended; leaf++) {
        if (leaf == 0x8000001D) {
            subleaf = 0;
            while (1) {
                __get_cpuid_count(leaf, subleaf, &eax, &ebx, &ecx, &edx);
                if (subleaf > 0 && ((eax & 0x1F) == 0))
                    break;
                APPEND_ENTRY(leaf, subleaf, eax, ebx, ecx, edx);
                subleaf++;
            }
        } else {
            __get_cpuid(leaf, &eax, &ebx, &ecx, &edx);
            APPEND_ENTRY(leaf, 0, eax, ebx, ecx, edx);
        }
    }

    /* --- Write JSON Output --- */
    fp = fopen("cpuid_data.json", "w");
    if (!fp) {
        fprintf(stderr, "Error opening output file.\n");
        free(entries);
        return 1;
    }

    fprintf(fp, "{\n  \"entries\": [\n");
    for (size_t i = 0; i < count; i++) {
        fprintf(fp,
            "    { \"leaf\": %u, \"subleaf\": %u, \"eax\": %u, \"ebx\": %u, \"ecx\": %u, \"edx\": %u }%s\n",
            entries[i].leaf,
            entries[i].subleaf,
            entries[i].eax,
            entries[i].ebx,
            entries[i].ecx,
            entries[i].edx,
            (i + 1 == count) ? "" : ","
        );
    }
    fprintf(fp, "  ]\n}\n");
    fclose(fp);

    printf("CPUID data captured in cpuid_data.json\n");
    free(entries);
    return 0;
}
