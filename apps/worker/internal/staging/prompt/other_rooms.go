package prompt

import (
	"fmt"
	"strings"
)

// buildOutdoorPrompt creates prompts specifically for outdoor/exterior spaces.
// This structure emphasizes outdoor context upfront and uses explicit material descriptions.
func buildOutdoorPrompt(style string, specifics ...string) string {
	var b strings.Builder

	// Lead with strong outdoor context
	b.WriteString("EXTERIOR SPACE STAGING: You are staging an outdoor patio, deck, or exterior area. ")
	b.WriteString(fmt.Sprintf("This is an OUTDOOR space requiring %s-style OUTDOOR FURNITURE ONLY. ", style))

	// CRITICAL: Structural preservation FIRST, before any furniture instructions
	b.WriteString("STRUCTURAL PRESERVATION - ABSOLUTE PRIORITY: ")
	b.WriteString("Do NOT modify, move, or alter ANY walls, railings, columns, doors, windows, ")
	b.WriteString("or architectural features. ")
	b.WriteString("Keep ALL existing surfaces (brick, concrete, wood decking, stone pavers, tile) ")
	b.WriteString("in their EXACT condition. ")
	b.WriteString("Do NOT remove, add, resize, or relocate windows or doors. ")
	b.WriteString("Do NOT change surface colors, finishes, or add new materials. ")
	b.WriteString("Do NOT modify built-in features like pergolas, awnings, or planters. ")
	b.WriteString("ONLY add outdoor furniture, planters, rugs, and decorative items. ")
	b.WriteString("The space's structure must remain COMPLETELY UNCHANGED. ")

	// Emphasize outdoor materials and furniture types
	b.WriteString("OUTDOOR FURNITURE REQUIREMENTS: Use only weather-resistant, exterior-grade furniture. ")
	b.WriteString("Acceptable materials include: powder-coated metals (aluminum, steel), ")
	b.WriteString("teak/eucalyptus/acacia wood, all-weather wicker/rattan, marine-grade fabrics, ")
	b.WriteString("concrete, stone, weather-resistant plastics. ")
	b.WriteString("NEVER use indoor furniture like upholstered sofas, fabric chairs, ")
	b.WriteString("or wooden furniture designed for interior spaces. ")

	// Add style-specific outdoor instructions
	for _, s := range specifics {
		b.WriteString(s)
		b.WriteString(" ")
	}

	// Critical outdoor placement rules
	b.WriteString("CRITICAL PLACEMENT RULES: ")
	b.WriteString("Furniture must be weather-appropriate with visible outdoor characteristics ")
	b.WriteString("(metal frames, slatted designs, outdoor cushion styles). ")
	b.WriteString("Do NOT block doorways or pathways. ")
	b.WriteString("Maintain natural outdoor lighting - this is an exterior space with sky visible. ")
	b.WriteString("All textiles must look like outdoor performance fabrics (cushions, pillows, rugs). ")
	b.WriteString("Plants and planters should be exterior-appropriate containers (not decorative indoor pots).")

	return b.String()
}

// Kitchen prompts - kitchens typically need minimal staging
func buildKitchenModern() string {
	return buildPrompt("kitchen", "modern",
		"Add bar stools at island/counter if present.",
		"Include minimal countertop items: coffee maker, fruit bowl.",
		"Add pendant lights if no fixtures exist.",
		"Do NOT add appliances to existing kitchens.",
		"Keep counters mostly clear.",
	)
}

func buildKitchenContemporary() string {
	return buildPrompt("kitchen", "contemporary",
		"Add upholstered bar stools.",
		"Include decorative bowl or vase on counter.",
		"Add pendant or track lighting.",
		"Keep styling minimal and functional.",
	)
}

func buildKitchenTraditional() string {
	return buildPrompt("kitchen", "traditional",
		"Add traditional bar stools.",
		"Include classic countertop accessories.",
		"Add traditional pendant lights.",
		"Use warm, inviting styling.",
	)
}

func buildKitchenIndustrial() string {
	return buildPrompt("kitchen", "industrial",
		"Add metal bar stools.",
		"Include industrial pendant lights.",
		"Keep counters clean with minimal metal accessories.",
	)
}

func buildKitchenScandinavian() string {
	return buildPrompt("kitchen", "Scandinavian",
		"Add simple wooden bar stools.",
		"Include minimal white/natural accessories.",
		"Add simple pendant lighting.",
		"Keep very minimal and clean.",
	)
}

// Bathroom prompts - bathrooms need minimal staging
func buildBathroomModern() string {
	return buildPrompt("bathroom", "modern",
		"Add rolled towels on counter or shelf.",
		"Include minimal accessories: soap dispenser, small plant.",
		"Add modern wall art if appropriate.",
		"Keep very minimal.",
	)
}

func buildBathroomContemporary() string {
	return buildPrompt("bathroom", "contemporary",
		"Add coordinating towels.",
		"Include spa-like accessories.",
		"Add soft lighting elements.",
	)
}

func buildBathroomTraditional() string {
	return buildPrompt("bathroom", "traditional",
		"Add traditional towel sets.",
		"Include classic bathroom accessories.",
		"Add framed art or mirror.",
	)
}

func buildBathroomIndustrial() string {
	return buildPrompt("bathroom", "industrial",
		"Add simple towels in neutral tones.",
		"Include minimal metal accessories.",
		"Keep industrial and minimal.",
	)
}

func buildBathroomScandinavian() string {
	return buildPrompt("bathroom", "Scandinavian",
		"Add white or light-colored towels.",
		"Include minimal natural accessories.",
		"Add small plants.",
	)
}

// Dining Room prompts
func buildDiningRoomModern() string {
	return buildPrompt("dining room", "modern",
		"Add modern dining table centered in space.",
		"Include 4-8 chairs depending on table size.",
		"Add modern chandelier or pendant light above table.",
		"Include minimal centerpiece: bowl, vase with branches.",
		"Add sideboard or buffet along wall if space allows.",
	)
}

func buildDiningRoomContemporary() string {
	return buildPrompt("dining room", "contemporary",
		"Add dining table with mixed materials.",
		"Include upholstered dining chairs.",
		"Add contemporary lighting fixture.",
		"Include artistic centerpiece.",
		"Add buffet or credenza if space allows.",
	)
}

func buildDiningRoomTraditional() string {
	return buildPrompt("dining room", "traditional",
		"Add wooden dining table.",
		"Include traditional dining chairs.",
		"Add chandelier above table.",
		"Include classic centerpiece: flowers, candlesticks.",
		"Add traditional sideboard.",
	)
}

func buildDiningRoomIndustrial() string {
	return buildPrompt("dining room", "industrial",
		"Add industrial dining table (metal/wood).",
		"Include metal or wood chairs.",
		"Add industrial pendant lights.",
		"Minimal centerpiece.",
		"Add industrial storage unit.",
	)
}

func buildDiningRoomScandinavian() string {
	return buildPrompt("dining room", "Scandinavian",
		"Add light wood dining table.",
		"Include simple wooden chairs with cushions.",
		"Add minimal pendant light.",
		"Include simple vase with branches.",
		"Keep very minimal.",
	)
}

// Office prompts
func buildOfficeModern() string {
	return buildPrompt("office", "modern",
		"Add modern desk with clean lines.",
		"Include ergonomic office chair.",
		"Add desk lamp with modern design.",
		"Include bookshelf or storage unit.",
		"Add minimal desk accessories: monitor, laptop stand.",
		"Do NOT overcrowd with items.",
	)
}

func buildOfficeContemporary() string {
	return buildPrompt("office", "contemporary",
		"Add desk with mixed materials.",
		"Include comfortable office chair.",
		"Add contemporary lighting.",
		"Include open shelving.",
		"Add plants and books.",
	)
}

func buildOfficeTraditional() string {
	return buildPrompt("office", "traditional",
		"Add wooden desk with traditional design.",
		"Include leather office chair.",
		"Add traditional desk lamp.",
		"Include wooden bookcases.",
		"Add classic desk accessories.",
	)
}

func buildOfficeIndustrial() string {
	return buildPrompt("office", "industrial",
		"Add desk with metal frame and wood top.",
		"Include industrial-style chair.",
		"Add metal shelving.",
		"Include minimal industrial accessories.",
	)
}

func buildOfficeScandinavian() string {
	return buildPrompt("office", "Scandinavian",
		"Add simple light wood desk.",
		"Include minimal office chair.",
		"Add simple desk lamp.",
		"Include white or light wood shelving.",
		"Add plants for warmth.",
	)
}

// Entryway prompts
func buildEntrywayModern() string {
	return buildPrompt("entryway", "modern",
		"Add slim console table against wall.",
		"Include modern mirror above console.",
		"Add table lamp or wall sconces.",
		"Include minimal decor: bowl, vase.",
		"Do NOT block entry path.",
	)
}

func buildEntrywayContemporary() string {
	return buildPrompt("entryway", "contemporary",
		"Add console table.",
		"Include decorative mirror.",
		"Add lighting fixture.",
		"Include welcoming accessories.",
		"Keep entry clear.",
	)
}

func buildEntrywayTraditional() string {
	return buildPrompt("entryway", "traditional",
		"Add traditional console table.",
		"Include ornate mirror.",
		"Add table lamp.",
		"Include classic decor.",
		"Keep pathway clear.",
	)
}

func buildEntrywayIndustrial() string {
	return buildPrompt("entryway", "industrial",
		"Add industrial console or bench.",
		"Include metal-framed mirror.",
		"Add industrial lighting.",
		"Minimal accessories.",
	)
}

func buildEntrywayScandinavian() string {
	return buildPrompt("entryway", "Scandinavian",
		"Add light wood console or bench.",
		"Include simple mirror.",
		"Add minimal lighting.",
		"Include coat hooks if wall space exists.",
		"Keep very minimal.",
	)
}

// Outdoor/Patio prompts
func buildOutdoorModern() string {
	return buildOutdoorPrompt("modern",
		"Add sleek outdoor sectional or lounge chairs made from powder-coated aluminum, "+
			"stainless steel, or all-weather synthetic wicker in neutral tones (gray, charcoal, white).",
		"Include low-profile outdoor coffee table made from teak, concrete, or metal with glass top.",
		"Add large planters (ceramic, concrete, or fiberglass) with tropical plants, "+
			"ornamental grasses, or succulents.",
		"Include outdoor cushions and pillows in performance fabrics that are "+
			"UV-resistant and water-repellent.",
		"Add string lights, outdoor floor lamps, or wall-mounted exterior fixtures "+
			"if mounting points exist.",
	)
}

func buildOutdoorContemporary() string {
	return buildOutdoorPrompt("contemporary",
		"Add comfortable outdoor sectional with deep seating, made from resin wicker or aluminum "+
			"with thick cushions in performance outdoor fabric.",
		"Include outdoor dining set or coffee table made from weather-resistant materials "+
			"like teak, eucalyptus wood, or powder-coated metal.",
		"Add variety of planters (mix of sizes) with lush greenery: "+
			"ferns, palms, flowering plants.",
		"Include outdoor rug made from polypropylene or recycled plastic "+
			"that can handle moisture.",
		"Add decorative outdoor pillows in weather-proof fabrics with patterns or textures.",
	)
}

func buildOutdoorTraditional() string {
	return buildOutdoorPrompt("traditional",
		"Add classic outdoor furniture made from wrought iron, cast aluminum, or natural teak wood in traditional designs.",
		"Include round or rectangular outdoor dining table with matching chairs, or conversation set with cushioned seating.",
		"Add terra cotta or ceramic planters with classic plants: boxwoods, hydrangeas, geraniums, ivy.",
		"Include traditional outdoor accessories: weather-resistant cushions in classic patterns, outdoor lanterns.",
		"Add warm outdoor lighting like hanging lanterns or coach-style wall fixtures.",
	)
}

func buildOutdoorIndustrial() string {
	return buildOutdoorPrompt("industrial",
		"Add raw metal outdoor furniture: steel mesh chairs, aluminum benches, or metal frame seating with minimal cushions.",
		"Include industrial-style outdoor table with metal base and wood or metal top (rust-finish acceptable for style).",
		"Add metal planters (galvanized steel, corten steel, or powder-coated metal) with architectural plants.",
		"Include Edison-style string lights or industrial cage outdoor fixtures.",
		"Keep accessories minimal: metal candle holders, simple outdoor cushions in gray or charcoal.",
	)
}

func buildOutdoorScandinavian() string {
	return buildOutdoorPrompt("Scandinavian",
		"Add simple outdoor furniture made from light-colored wood (acacia, eucalyptus, or painted pine) with clean lines.",
		"Include wooden outdoor table with matching chairs or bench seating, minimal design.",
		"Add white or natural-colored ceramic planters with simple greenery: herbs, small trees, grasses.",
		"Include neutral outdoor cushions and textiles in natural linen-look performance fabrics (white, cream, light gray).",
		"Add simple outdoor candles or minimalist outdoor lighting.",
	)
}

// Default/fallback prompts
func buildDefaultModern() string {
	return buildPrompt("room", "modern",
		"Add appropriate modern furniture for this space.",
		"Include modern lighting and accessories.",
		"Keep design clean and minimal.",
		"Do NOT overcrowd the space.",
	)
}

func buildDefaultContemporary() string {
	return buildPrompt("room", "contemporary",
		"Add appropriate contemporary furniture.",
		"Include layered lighting.",
		"Add comfortable, stylish pieces.",
		"Create inviting atmosphere.",
	)
}

func buildDefaultTraditional() string {
	return buildPrompt("room", "traditional",
		"Add appropriate traditional furniture.",
		"Include classic lighting and accessories.",
		"Create warm, timeless space.",
	)
}

func buildDefaultIndustrial() string {
	return buildPrompt("room", "industrial",
		"Add appropriate industrial-style furniture.",
		"Include industrial materials and finishes.",
		"Keep functional and minimal.",
	)
}

func buildDefaultScandinavian() string {
	return buildPrompt("room", "Scandinavian",
		"Add appropriate Scandinavian furniture in light woods or white.",
		"Include minimal, functional pieces.",
		"Add natural elements and plants.",
		"Keep very minimal and bright.",
	)
}
