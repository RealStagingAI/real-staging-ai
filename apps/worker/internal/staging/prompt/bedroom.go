package prompt

func buildBedroomModern() string {
	return buildPrompt("bedroom", "modern",
		"Add a platform bed with low profile headboard centered on main wall.",
		"Include matching nightstands on either side of bed.",
		"Add modern table lamps on nightstands.",
		"Do NOT add TV, media console, or entertainment furniture.",
		"Do NOT place furniture in closets or blocking closet doors.",
		"Include a simple dresser along a secondary wall.",
		"Add minimal bedding: solid colors, geometric patterns.",
		"Use a modern area rug at foot of bed.",
		"Add minimal wall art above headboard.",
		"Include a reading chair ONLY if ample space exists.",
	)
}

func buildBedroomContemporary() string {
	return buildPrompt("bedroom", "contemporary",
		"Add an upholstered bed with soft headboard centered on main wall.",
		"Include nightstands with mixed materials (wood/metal).",
		"Add contemporary table lamps.",
		"Do NOT add TV or media furniture in bedrooms.",
		"Do NOT block closets with furniture.",
		"Include a dresser with clean design.",
		"Add layered bedding with textured fabrics.",
		"Use an area rug with subtle pattern.",
		"Add framed artwork or mirror above headboard.",
		"Include a comfortable reading chair if space allows.",
	)
}

func buildBedroomTraditional() string {
	return buildPrompt("bedroom", "traditional",
		"Add a bed with classic headboard (wood or upholstered) centered on main wall.",
		"Include matching wooden nightstands.",
		"Add traditional table lamps with fabric shades.",
		"Do NOT add TV or media console.",
		"Do NOT place furniture in closets.",
		"Include a traditional dresser or chest of drawers.",
		"Add classic bedding with traditional patterns.",
		"Use a traditional area rug (Oriental, Persian style).",
		"Add framed artwork with ornate frames above bed.",
		"Include a upholstered bench at foot of bed if space allows.",
	)
}

func buildBedroomIndustrial() string {
	return buildPrompt("bedroom", "industrial",
		"Add a bed with metal frame or reclaimed wood headboard.",
		"Include industrial nightstands with metal and wood.",
		"Add Edison bulb lamps or metal industrial lighting.",
		"Do NOT add TV or entertainment furniture.",
		"Do NOT block closets with furniture.",
		"Include a dresser with industrial hardware.",
		"Add simple, solid-colored bedding.",
		"Use a simple jute or sisal area rug.",
		"Add minimal industrial-style wall decor.",
		"Keep furniture minimal and functional.",
	)
}

func buildBedroomScandinavian() string {
	return buildPrompt("bedroom", "Scandinavian",
		"Add a simple bed frame in light wood or white finish.",
		"Include minimalist nightstands in light wood.",
		"Add simple pendant lights or minimal table lamps.",
		"Do NOT add TV or media furniture.",
		"Do NOT block closets.",
		"Include a simple dresser in light wood or white.",
		"Add white or light neutral bedding with texture.",
		"Use a light-colored area rug (white, cream, light gray).",
		"Add minimal wall art with natural, calming themes.",
		"Include houseplants for warmth.",
	)
}
