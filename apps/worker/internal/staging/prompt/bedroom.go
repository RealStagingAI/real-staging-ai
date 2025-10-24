package prompt

import (
	"fmt"
	"strings"
)

// buildBedroomPrompt creates prompts specifically for bedroom spaces.
// This structure emphasizes bedroom context upfront and uses explicit furniture descriptions.
func buildBedroomPrompt(style string, specifics ...string) string {
	var b strings.Builder

	// Lead with strong bedroom context
	b.WriteString("BEDROOM STAGING: You are staging a bedroom. ")
	b.WriteString(fmt.Sprintf("This is a BEDROOM requiring %s-style BEDROOM FURNITURE ONLY. ", style))

	// Emphasize bedroom furniture requirements
	b.WriteString("BEDROOM FURNITURE REQUIREMENTS: This is a sleeping space. ")
	b.WriteString("Essential bedroom furniture includes: bed with headboard (platform, upholstered, or wood frame), ")
	b.WriteString("nightstands (1-2 flanking the bed), table lamps or pendant lights, dresser or chest of drawers. ")
	b.WriteString("NEVER add living room furniture like sofas, couches, sectionals, coffee tables, or TV stands. ")
	b.WriteString("NEVER add entertainment centers, media consoles, or television units. ")
	b.WriteString("Bedrooms are for sleeping, not watching TV. ")

	// Add style-specific bedroom instructions
	for _, s := range specifics {
		b.WriteString(s)
		b.WriteString(" ")
	}

	// Critical bedroom rules
	b.WriteString("CRITICAL BEDROOM RULES: ")
	b.WriteString("The bed must be the focal point and largest piece of furniture. ")
	b.WriteString("Place bed against the main wall, typically centered. ")
	b.WriteString("Nightstands go on either side of the bed, not elsewhere. ")
	b.WriteString("Do NOT block closet doors or place furniture inside closets. ")
	b.WriteString("Do NOT add seating unless space explicitly allows (reading chair in corner only). ")
	b.WriteString("Keep walls, paint colors, and structural elements EXACTLY as they are. ")
	b.WriteString("Bedding should be neatly made with pillows arranged. ")
	b.WriteString("Area rugs should be at the foot of the bed or under the bed, not blocking pathways.")

	return b.String()
}

func buildBedroomModern() string {
	return buildBedroomPrompt("modern",
		"Add a low-profile platform bed with simple geometric headboard in dark wood, walnut, or "+
			"upholstered fabric, centered on the main wall.",
		"Include two matching modern nightstands with clean lines, floating or with simple legs, "+
			"placed symmetrically on each side of the bed.",
		"Add sleek modern table lamps with minimal shades or contemporary pendant lights "+
			"hanging above nightstands.",
		"Include a modern dresser with handleless drawers or metal pulls along a secondary wall.",
		"Add bedding in solid neutral colors (white, gray, navy) with geometric throw pillows.",
		"Use a modern area rug in geometric pattern or solid color at the foot of the bed.",
		"Add minimal wall art above headboard: abstract prints, black and white photography, or "+
			"simple line drawings in thin frames.",
		"Only add a modern accent chair (Eames-style or sculptural) in corner if room is large.",
	)
}

func buildBedroomContemporary() string {
	return buildBedroomPrompt("contemporary",
		"Add an upholstered bed with tufted or channel-stitched headboard in velvet, linen, or "+
			"performance fabric, centered on the main wall.",
		"Include two nightstands with mixed materials: wood base with metal legs or accents, "+
			"placed on each side of the bed.",
		"Add contemporary table lamps with sculptural bases and fabric drum shades in neutral tones.",
		"Include a dresser with contemporary design: mix of wood and metal, clean lines with "+
			"subtle hardware.",
		"Add layered bedding with textured fabrics: duvet, decorative pillows in varying sizes, "+
			"throw blanket in complementary colors.",
		"Use an area rug with subtle geometric or abstract pattern in neutral palette.",
		"Add framed artwork or large mirror with modern frame above headboard.",
		"Only include a comfortable upholstered reading chair in corner if space is ample.",
	)
}

func buildBedroomTraditional() string {
	return buildBedroomPrompt("traditional",
		"Add a classic bed with substantial wood headboard (cherry, mahogany, oak) or elegantly "+
			"upholstered headboard with nailhead trim, centered on the main wall.",
		"Include two matching wooden nightstands with traditional details: curved legs, brass pulls, "+
			"or carved accents.",
		"Add traditional table lamps with ceramic or wood bases and fabric shades in warm tones.",
		"Include a traditional dresser or chest of drawers in matching wood with classic hardware "+
			"(brass knobs or pulls).",
		"Add classic bedding with traditional patterns: damask, toile, floral, or paisley in "+
			"coordinating colors.",
		"Use a traditional area rug such as Oriental, Persian, or floral pattern in rich colors.",
		"Add framed artwork with ornate gold or wood frames above bed: classic paintings, "+
			"botanical prints.",
		"Only include an upholstered bench at foot of bed if space clearly allows.",
	)
}

func buildBedroomIndustrial() string {
	return buildBedroomPrompt("industrial",
		"Add a bed with industrial metal frame (iron or steel) or reclaimed wood headboard with "+
			"visible grain and natural finish.",
		"Include two industrial nightstands: metal frame with wood shelves, pipe legs, or "+
			"reclaimed wood with metal corners.",
		"Add Edison bulb lamps in metal cages or industrial pendant lights with exposed bulbs "+
			"hanging above nightstands.",
		"Include a dresser with industrial hardware: metal handles, exposed bolts, combination of "+
			"metal frame and wood drawers.",
		"Add simple bedding in solid colors: charcoal, black, white, or denim blue with "+
			"minimal patterns.",
		"Use a simple natural fiber area rug: jute, sisal, or solid dark gray.",
		"Add minimal industrial-style wall decor: metal signs, black and white photography in "+
			"metal frames, or exposed shelving.",
		"Keep furniture minimal and functional - avoid clutter or decorative excess.",
	)
}

func buildBedroomScandinavian() string {
	return buildBedroomPrompt("Scandinavian",
		"Add a simple bed frame in light wood (birch, ash, or pine) or white-painted wood with "+
			"clean minimal headboard.",
		"Include two minimalist nightstands in light wood or white with simple drawer and clean legs.",
		"Add simple white or glass pendant lights with minimal design or slender table lamps in "+
			"neutral tones.",
		"Include a simple dresser in light wood or white with minimal hardware or push-to-open drawers.",
		"Add white or light neutral bedding (cream, light gray, soft beige) with textured "+
			"weaves or knits for warmth.",
		"Use a light-colored area rug in white, cream, light gray, or natural jute.",
		"Add minimal wall art with natural calming themes: line drawings, nature prints, or "+
			"simple typography in light frames.",
		"Include one or two houseplants in simple ceramic pots for organic warmth.",
	)
}
