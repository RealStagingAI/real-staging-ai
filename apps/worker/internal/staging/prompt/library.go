package prompt

import (
	"fmt"
	"strings"
)

// Library provides curated prompts for different room types and styles.
type Library struct {
	prompts map[string]map[string]string
}

// New creates a new prompt library with predefined prompts.
func New() *Library {
	lib := &Library{
		prompts: make(map[string]map[string]string),
	}
	lib.loadPrompts()
	return lib
}

// Get retrieves a prompt for the given room type and style.
// Returns the prompt and a boolean indicating if it was found.
func (l *Library) Get(roomType, style string) (string, bool) {
	if roomType == "" {
		roomType = "default"
	}
	if style == "" {
		style = "modern"
	}

	// Try exact match first
	if styles, ok := l.prompts[roomType]; ok {
		if prompt, ok := styles[style]; ok {
			return prompt, true
		}
		// Fall back to default style for this room type
		if prompt, ok := styles["default"]; ok {
			return prompt, true
		}
	}

	// Fall back to default room type with requested style
	if styles, ok := l.prompts["default"]; ok {
		if prompt, ok := styles[style]; ok {
			return prompt, true
		}
	}

	return "", false
}

// Build constructs a prompt for the given room type and style.
// If customPrompt is provided, it takes precedence.
// Otherwise, looks up the prompt from the library.
func (l *Library) Build(roomType, style, customPrompt string) string {
	if customPrompt != "" {
		return customPrompt
	}

	if prompt, ok := l.Get(roomType, style); ok {
		return prompt
	}

	// Ultimate fallback
	return l.buildGenericPrompt(roomType, style)
}

// buildGenericPrompt creates a basic prompt when no specific one is found.
func (l *Library) buildGenericPrompt(roomType, style string) string {
	var b strings.Builder
	b.WriteString("You are a professional real estate staging photographer. ")
	b.WriteString("Add appropriate furniture to make this space appealing to buyers")

	if style != "" {
		b.WriteString(" who love ")
		b.WriteString(style)
		b.WriteString(" style")
	}
	b.WriteString(". ")

	if roomType != "" && roomType != "default" {
		b.WriteString("This is a ")
		b.WriteString(roomType)
		b.WriteString(". ")
	}

	b.WriteString("Keep walls, doors, and structure exactly as they are. Do not block doorways or hallways.")
	return b.String()
}

// loadPrompts populates the library with curated prompts.
func (l *Library) loadPrompts() {
	// Living Room prompts
	l.prompts["living_room"] = map[string]string{
		"modern":       buildLivingRoomModern(),
		"contemporary": buildLivingRoomContemporary(),
		"traditional":  buildLivingRoomTraditional(),
		"industrial":   buildLivingRoomIndustrial(),
		"scandinavian": buildLivingRoomScandinavian(),
		"default":      buildLivingRoomModern(),
	}

	// Bedroom prompts
	l.prompts["bedroom"] = map[string]string{
		"modern":       buildBedroomModern(),
		"contemporary": buildBedroomContemporary(),
		"traditional":  buildBedroomTraditional(),
		"industrial":   buildBedroomIndustrial(),
		"scandinavian": buildBedroomScandinavian(),
		"default":      buildBedroomModern(),
	}

	// Kitchen prompts
	l.prompts["kitchen"] = map[string]string{
		"modern":       buildKitchenModern(),
		"contemporary": buildKitchenContemporary(),
		"traditional":  buildKitchenTraditional(),
		"industrial":   buildKitchenIndustrial(),
		"scandinavian": buildKitchenScandinavian(),
		"default":      buildKitchenModern(),
	}

	// Bathroom prompts
	l.prompts["bathroom"] = map[string]string{
		"modern":       buildBathroomModern(),
		"contemporary": buildBathroomContemporary(),
		"traditional":  buildBathroomTraditional(),
		"industrial":   buildBathroomIndustrial(),
		"scandinavian": buildBathroomScandinavian(),
		"default":      buildBathroomModern(),
	}

	// Dining Room prompts
	l.prompts["dining_room"] = map[string]string{
		"modern":       buildDiningRoomModern(),
		"contemporary": buildDiningRoomContemporary(),
		"traditional":  buildDiningRoomTraditional(),
		"industrial":   buildDiningRoomIndustrial(),
		"scandinavian": buildDiningRoomScandinavian(),
		"default":      buildDiningRoomModern(),
	}

	// Office prompts
	l.prompts["office"] = map[string]string{
		"modern":       buildOfficeModern(),
		"contemporary": buildOfficeContemporary(),
		"traditional":  buildOfficeTraditional(),
		"industrial":   buildOfficeIndustrial(),
		"scandinavian": buildOfficeScandinavian(),
		"default":      buildOfficeModern(),
	}

	// Entryway prompts
	l.prompts["entryway"] = map[string]string{
		"modern":       buildEntrywayModern(),
		"contemporary": buildEntrywayContemporary(),
		"traditional":  buildEntrywayTraditional(),
		"industrial":   buildEntrywayIndustrial(),
		"scandinavian": buildEntrywayScandinavian(),
		"default":      buildEntrywayModern(),
	}

	// Outdoor/Patio prompts
	l.prompts["outdoor"] = map[string]string{
		"modern":       buildOutdoorModern(),
		"contemporary": buildOutdoorContemporary(),
		"traditional":  buildOutdoorTraditional(),
		"industrial":   buildOutdoorIndustrial(),
		"scandinavian": buildOutdoorScandinavian(),
		"default":      buildOutdoorModern(),
	}

	// Default prompts (used as fallback)
	l.prompts["default"] = map[string]string{
		"modern":       buildDefaultModern(),
		"contemporary": buildDefaultContemporary(),
		"traditional":  buildDefaultTraditional(),
		"industrial":   buildDefaultIndustrial(),
		"scandinavian": buildDefaultScandinavian(),
		"default":      buildDefaultModern(),
	}
}

// Helper function to build common prompt structure
func buildPrompt(roomType, style string, specifics ...string) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Professional real estate staging for a %s with %s design. ", roomType, style))
	b.WriteString("You are a professional real estate photographer creating staged photos. ")

	// Add room-specific instructions
	for _, s := range specifics {
		b.WriteString(s)
		b.WriteString(" ")
	}

	// Common rules for all prompts
	b.WriteString("CRITICAL RULES: ")
	b.WriteString("Keep all walls, paint colors, and structural elements EXACTLY as they are. ")
	b.WriteString("Do NOT block doorways, hallways, or thresholds with furniture. ")
	b.WriteString("Do NOT change wall colors or add wall treatments. ")
	b.WriteString("Do NOT alter windows, doors, or light fixtures. ")
	b.WriteString("Furniture must be appropriately sized and placed for the room. ")
	b.WriteString("Maintain realistic lighting and shadows.")

	return b.String()
}
