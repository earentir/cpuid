package cpuid

import (
	"encoding/json"
	"os"
)

// Entry represents one cpuid call result.
type Entry struct {
	Leaf    uint32 `json:"leaf"`
	Subleaf uint32 `json:"subleaf"`
	EAX     uint32 `json:"eax"`
	EBX     uint32 `json:"ebx"`
	ECX     uint32 `json:"ecx"`
	EDX     uint32 `json:"edx"`
}

// Data holds a slice of Entry.
type Data struct {
	Entries []Entry `json:"entries"`
}

// CaptureData traverses the full CPUID hierarchy and writes the data to cpuid_data.json.
func CaptureData(filename string) error {
	var data Data

	// Capture Standard CPUID Leaves.
	// First, get the maximum supported standard leaf.
	maxStandard, _, _, _ := cpuid(0, 0)
	for leaf := uint32(0); leaf <= maxStandard; leaf++ {
		// For leaves that support multiple subleafs.
		if leaf == 4 || leaf == 0xB || leaf == 0xD {
			subleaf := uint32(0)
			for {
				a, b, c, d := cpuid(leaf, subleaf)
				// For leaf 4: stop if the cache type (lower 5 bits of EAX) is 0 (for subleaf > 0).
				if leaf == 4 && subleaf > 0 && (a&0x1F) == 0 {
					break
				}
				// For leaf 0xB (extended topology), stop if EAX is 0 (after the first subleaf).
				if leaf == 0xB && subleaf > 0 && a == 0 {
					break
				}
				// For leaf 0xD, stop if all registers are zero (after the first subleaf).
				if leaf == 0xD && subleaf > 0 && a == 0 && b == 0 && c == 0 && d == 0 {
					break
				}
				data.Entries = append(data.Entries, Entry{
					Leaf:    leaf,
					Subleaf: subleaf,
					EAX:     a,
					EBX:     b,
					ECX:     c,
					EDX:     d,
				})
				subleaf++
			}
		} else {
			// For leaves without subleaf iteration.
			a, b, c, d := cpuid(leaf, 0)
			data.Entries = append(data.Entries, Entry{
				Leaf:    leaf,
				Subleaf: 0,
				EAX:     a,
				EBX:     b,
				ECX:     c,
				EDX:     d,
			})
		}
	}

	// Capture Extended CPUID Leaves.
	// Get the maximum extended leaf from cpuid(0x80000000, 0).
	maxExtended, _, _, _ := cpuid(0x80000000, 0)
	for leaf := uint32(0x80000000); leaf <= maxExtended; leaf++ {
		// For extended leaf 0x8000001D (cache info on some AMD CPUs) iterate subleafs.
		if leaf == 0x8000001D {
			subleaf := uint32(0)
			for {
				a, b, c, d := cpuid(leaf, subleaf)
				if subleaf > 0 && (a&0x1F) == 0 {
					break
				}
				data.Entries = append(data.Entries, Entry{
					Leaf:    leaf,
					Subleaf: subleaf,
					EAX:     a,
					EBX:     b,
					ECX:     c,
					EDX:     d,
				})
				subleaf++
			}
		} else {
			a, b, c, d := cpuid(leaf, 0)
			data.Entries = append(data.Entries, Entry{
				Leaf:    leaf,
				Subleaf: 0,
				EAX:     a,
				EBX:     b,
				ECX:     c,
				EDX:     d,
			})
		}
	}

	// Write the collected CPUID data to a JSON file.
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return err
	}

	return nil
}

// DataFromFile reads cpuid_data.json and returns a Data struct.
func DataFromFile(filename string) (Data, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Data{}, err
	}
	defer file.Close()

	var data Data
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return Data{}, err
	}

	return data, nil
}
