package prompt

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
	return buildPrompt("outdoor patio/deck", "modern",
		"Add modern outdoor seating set.",
		"Include outdoor coffee table.",
		"Add outdoor planters with greenery.",
		"Do NOT add indoor furniture outside.",
		"Include weather-appropriate cushions.",
		"Add outdoor lighting if fixture points exist.",
	)
}

func buildOutdoorContemporary() string {
	return buildPrompt("outdoor patio/deck", "contemporary",
		"Add comfortable outdoor seating.",
		"Include outdoor tables.",
		"Add planters and plants.",
		"Use weather-resistant materials only.",
	)
}

func buildOutdoorTraditional() string {
	return buildPrompt("outdoor patio/deck", "traditional",
		"Add classic outdoor furniture.",
		"Include traditional planters.",
		"Add welcoming outdoor accessories.",
		"Use traditional outdoor materials.",
	)
}

func buildOutdoorIndustrial() string {
	return buildPrompt("outdoor patio/deck", "industrial",
		"Add metal outdoor furniture.",
		"Include industrial planters.",
		"Keep minimal and functional.",
	)
}

func buildOutdoorScandinavian() string {
	return buildPrompt("outdoor patio/deck", "Scandinavian",
		"Add simple wooden outdoor furniture.",
		"Include natural planters.",
		"Keep minimal with natural elements.",
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
