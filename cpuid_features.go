// Package cpuid provides information about the CPU running the current program.
package cpuid

import "sort"

// GetAllFeatureCategories reports all categories
func GetAllFeatureCategories() []string {
	categories := make([]string, 0, len(cpuFeaturesList))
	for category := range cpuFeaturesList {
		categories = append(categories, category)
	}
	//sort categories
	sort.Strings(categories)

	return categories
}

// GetAllFeatureCategoriesDetailed returns all categories and their features with details.
func GetAllFeatureCategoriesDetailed() map[string][]map[string]string {
	details := make(map[string][]map[string]string)

	for _, fs := range cpuFeaturesList {
		categoryDetails := []map[string]string{}
		for _, feat := range fs.features {
			vendor := feat.vendor
			if vendor == "common" {
				vendor = "both"
			}

			entry := map[string]string{
				"name":        feat.name,
				"description": feat.description,
				"vendor":      vendor,
			}

			if feat.equivalentFeatureName != "" {
				entry["equivalent"] = feat.equivalentFeatureName
			}

			categoryDetails = append(categoryDetails, entry)
		}
		details[fs.name] = categoryDetails
	}

	return details
}

// GetAllKnownFeatures reports all known features
func GetAllKnownFeatures(category string) []string {
	fs, exists := cpuFeaturesList[category]
	if !exists {
		return nil
	}

	features := make([]string, 0, len(fs.features))
	for _, f := range fs.features {
		features = append(features, f.name)
	}
	return features
}

// GetSupportedFeatures reports all supported features
func GetSupportedFeatures(category string, offline bool, filename string) []string {
	fs, exists := cpuFeaturesList[category]
	if !exists {
		return nil
	}

	// If there's a condition to check (some featuresets may only be valid if condition is met)
	if fs.condition != nil && !fs.condition(0) {
		return nil
	}

	a, b, c, d := CPUIDWithMode(fs.leaf, fs.subleaf, offline, filename)
	var regValue uint32
	switch fs.register {
	case 0:
		regValue = a
	case 1:
		regValue = b
	case 2:
		regValue = c
	case 3:
		regValue = d
	}

	supported := []string{}
	for bit, f := range fs.features {
		if (regValue>>bit)&1 == 1 {
			supported = append(supported, f.name)
		}
	}
	return supported
}

// IsFeatureSupported reports if a feature is supported
func IsFeatureSupported(featureName string, offline bool, filename string) bool {
	for _, fs := range cpuFeaturesList {
		// Check condition if present
		if fs.condition != nil && !fs.condition(0) {
			continue
		}

		var bitPos *int
		for bit, f := range fs.features {
			if f.name == featureName {
				bitPos = &bit
				break
			}
		}

		if bitPos == nil {
			continue // feature not in this category
		}

		a, b, c, d := CPUIDWithMode(fs.leaf, fs.subleaf, offline, filename)
		var regValue uint32
		switch fs.register {
		case 0:
			regValue = a
		case 1:
			regValue = b
		case 2:
			regValue = c
		case 3:
			regValue = d
		}

		if (regValue>>(*bitPos))&1 == 1 {
			return true
		} else {
			return false
		}
	}
	return false
}
