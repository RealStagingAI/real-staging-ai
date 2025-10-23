package prompt

func buildLivingRoomModern() string {
	return buildPrompt("living room", "modern",
		"Add a sleek sofa or sectional with clean lines and neutral colors (gray, white, beige).",
		"Include a modern coffee table with geometric design.",
		"Add 1-2 accent chairs if space allows.",
		"Include a modern media console or TV stand along ONE wall (not in front of windows).",
		"Add contemporary artwork or a large mirror on walls.",
		"Include modern lighting like floor lamps or table lamps.",
		"Add minimal decorative elements: books, plants, geometric vases.",
		"Use a modern area rug to define the seating area.",
		"Do NOT place TV in bedrooms, closets, or bathrooms.",
	)
}

func buildLivingRoomContemporary() string {
	return buildPrompt("living room", "contemporary",
		"Add a comfortable sofa or sectional with soft, rounded edges.",
		"Include a contemporary coffee table with mixed materials (wood and metal).",
		"Add 1-2 upholstered accent chairs.",
		"Include a media console with clean design.",
		"Add layered lighting: floor lamp, table lamps.",
		"Include contemporary art pieces and decorative mirrors.",
		"Add textured throw pillows and a cozy throw blanket.",
		"Use an area rug with subtle patterns.",
		"Include organic decorative elements: plants, natural wood accents.",
	)
}

func buildLivingRoomTraditional() string {
	return buildPrompt("living room", "traditional",
		"Add a classic sofa with rolled arms and traditional upholstery.",
		"Include a wooden coffee table with traditional detailing.",
		"Add coordinating armchairs with classic silhouettes.",
		"Include a traditional media console or entertainment center.",
		"Add table lamps with traditional bases and fabric shades.",
		"Include framed artwork or mirrors with ornate frames.",
		"Add decorative elements: books, classic vases, candlesticks.",
		"Use a traditional patterned area rug (Persian, Oriental style).",
		"Include wooden side tables or console tables.",
	)
}

func buildLivingRoomIndustrial() string {
	return buildPrompt("living room", "industrial",
		"Add a leather or distressed fabric sofa.",
		"Include a coffee table with metal base and wood or concrete top.",
		"Add metal-framed chairs or stools.",
		"Include an industrial media console with exposed hardware.",
		"Add Edison bulb floor lamps or exposed filament lighting.",
		"Include metal-framed artwork or industrial mirrors.",
		"Add minimal decorative elements with industrial materials.",
		"Use a jute or sisal area rug.",
		"Include metal shelving or storage units if space allows.",
	)
}

func buildLivingRoomScandinavian() string {
	return buildPrompt("living room", "Scandinavian",
		"Add a light-colored sofa with clean lines (white, light gray, beige).",
		"Include a simple wooden coffee table in natural finish.",
		"Add 1-2 accent chairs with wooden legs and light upholstery.",
		"Include a minimalist media console in light wood or white.",
		"Add simple lighting with natural materials.",
		"Include minimal artwork with light, natural themes.",
		"Add cozy textiles: knit throws, simple pillows in neutral tones.",
		"Use a light-colored area rug (white, cream, light gray).",
		"Include houseplants for natural warmth.",
	)
}
