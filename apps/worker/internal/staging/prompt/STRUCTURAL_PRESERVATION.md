# Structural Preservation in Staging Prompts

## Problem
The AI staging model was modifying room structures when it should only add furniture and decor:
- Moving walls
- Removing or resizing windows
- Changing paint colors
- Altering architectural features

## Root Cause
Structural preservation rules were placed **after** furniture instructions in the prompts. This allowed the model to start generating structural changes before encountering the constraints.

## Solution
Move **STRUCTURAL PRESERVATION** rules to the **very beginning** of all prompts, immediately after establishing room context and before any furniture descriptions.

## Prompt Structure (New Pattern)

```
1. Room context (e.g., "BEDROOM STAGING: You are staging a bedroom.")
2. STRUCTURAL PRESERVATION - ABSOLUTE PRIORITY (before any furniture!)
   - Do NOT modify walls, paint, windows, doors, architectural features
   - Keep ALL existing structures in EXACT positions
   - Do NOT remove, add, resize, or relocate windows
   - Do NOT change colors, finishes, or materials
   - ONLY add furniture, rugs, artwork, decorative items
3. Room-specific furniture requirements
4. Style-specific furniture descriptions
5. Critical placement rules
```

## Files Modified

### 1. `bedroom.go` - Bedroom Prompts
**Before:** Structural rule buried at line 38
```go
b.WriteString("Keep walls, paint colors, and structural elements EXACTLY as they are. ")
```

**After:** Structural preservation block at lines 17-26 (before any furniture)
```go
b.WriteString("STRUCTURAL PRESERVATION - ABSOLUTE PRIORITY: ")
b.WriteString("Do NOT modify, move, or alter ANY walls, paint colors, windows, doors, or architectural features. ")
b.WriteString("Keep ALL existing walls in their EXACT positions. ")
// ... 8 lines of detailed structural constraints
```

### 2. `library.go` - Generic Room Prompts
Updated `buildPrompt()` function used by:
- Living rooms
- Kitchens
- Dining rooms
- Bathrooms
- Entryways
- Default rooms

**Before:** Structural rules at the end (lines 199-203)
**After:** Structural preservation block at lines 192-201 (before furniture instructions)

### 3. `other_rooms.go` - Outdoor Prompts
Updated `buildOutdoorPrompt()` function for patios, decks, exterior spaces.

**Before:** Rules scattered, surfaces mentioned mid-prompt
**After:** Comprehensive structural preservation at lines 17-27, covering:
- Walls, railings, columns
- Surface materials (decking, pavers, tile)
- Built-in features (pergolas, awnings)

## Why This Works

AI models process prompts sequentially. By establishing hard constraints **first**:
1. Model learns what it **cannot** do before what it **can** do
2. Structural preservation becomes the context for all subsequent generation
3. Furniture descriptions are framed within "additive only" constraints

This follows the same successful pattern that prevented living room furniture (sofas, TVs) from appearing in bedrooms.

## Testing

- ✅ All unit tests pass
- ✅ Linting clean (0 issues)
- ✅ Committed: `f8227bf`

## Expected Results

With these changes, bedroom staging should:
- ✅ Add beds, nightstands, dressers (furniture)
- ✅ Add lamps, artwork, rugs (decor)
- ❌ NOT move walls
- ❌ NOT remove or resize windows
- ❌ NOT change paint colors
- ❌ NOT alter architectural features

## Monitoring

After deploying, monitor staging results for:
1. Wall positions remain unchanged
2. Window count and size stays consistent
3. Paint colors preserved
4. Only furniture and decorative items added
