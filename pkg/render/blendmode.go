package render

type BlendMode uint8

const (
	BlendAlpha            BlendMode = iota // Blend textures considering alpha (default)
	BlendAdditive                          // Blend textures adding colors
	BlendMultiplied                        // Blend textures multiplying colors
	BlendAddColors                         // Blend textures adding colors (alternative)
	BlendSubtractColors                    // Blend textures subtracting colors (alternative)
	BlendAlphaPremultiply                  // Blend premultiplied textures considering alpha
)
