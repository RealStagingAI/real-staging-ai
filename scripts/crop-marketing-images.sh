#!/bin/bash

# Crop and normalize marketing images to match perfectly
# This script ensures before/after images have identical dimensions and framing
# Usage: ./scripts/crop-marketing-images.sh

set -e

MARKETING_DIR="apps/web/public/images/marketing"
TEMP_DIR="/tmp/real-staging-crop"

echo "🔧 Setting up image cropping workspace..."
mkdir -p "$TEMP_DIR"
mkdir -p "$MARKETING_DIR"

# Check if ImageMagick is available
if ! command -v magick &> /dev/null; then
    echo "❌ ImageMagick is required but not installed."
    echo ""
    echo "📥 Install ImageMagick:"
    echo "   brew install imagemagick  # macOS"
    echo "   sudo apt install imagemagick  # Ubuntu"
    exit 1
fi

echo "📏 Analyzing current images..."

# Function to get image dimensions
get_dimensions() {
    local file="$1"
    magick identify -format "%wx%h" "$file" 2>/dev/null || echo "unknown"
}

# Function to normalize both images to a common size
normalize_to_common_size() {
    local before="$1"
    local after="$2"
    local output_before="$3"
    local output_after="$4"
    
    echo "  📐 Normalizing images to common dimensions..."
    
    # Use a standard 4:3 aspect ratio that works well for real estate
    # Target: 1200x900 (good balance of quality and performance)
    local target_width="1200"
    local target_height="900"
    
    echo "    Target dimensions: ${target_width}x${target_height}"
    
    # Process before image
    echo "    Processing before image..."
    magick "$before" \
        -gravity center \
        -resize "${target_width}x${target_height}^" \
        -extent "${target_width}x${target_height}" \
        -quality 85 \
        -strip \
        "$output_before"
    
    # Process after image  
    echo "    Processing after image..."
    magick "$after" \
        -gravity center \
        -resize "${target_width}x${target_height}^" \
        -extent "${target_width}x${target_height}" \
        -quality 85 \
        -strip \
        "$output_after"
}

# Process bedroom images
echo "🛠️ Processing bedroom transformation..."

if [[ -f "$MARKETING_DIR/bedroom-before.jpg" && -f "$MARKETING_DIR/bedroom-after.jpg" ]]; then
    # Create backup originals
    cp "$MARKETING_DIR/bedroom-before.jpg" "$TEMP_DIR/bedroom-before-original.jpg"
    cp "$MARKETING_DIR/bedroom-after.jpg" "$TEMP_DIR/bedroom-after-original.jpg"
    
    # Convert after image to JPEG if it's WebP
    if file "$MARKETING_DIR/bedroom-after.jpg" | grep -q "Web/P"; then
        echo "  🔄 Converting after image from WebP to JPEG..."
        magick "$MARKETING_DIR/bedroom-after.jpg" "$TEMP_DIR/bedroom-after-temp.jpg"
        AFTER_TEMP="$TEMP_DIR/bedroom-after-temp.jpg"
    else
        cp "$MARKETING_DIR/bedroom-after.jpg" "$TEMP_DIR/bedroom-after-temp.jpg"
        AFTER_TEMP="$TEMP_DIR/bedroom-after-temp.jpg"
    fi
    
    # Normalize both images to identical dimensions
    echo "  🎯 Normalizing both images to standard dimensions..."
    normalize_to_common_size "$MARKETING_DIR/bedroom-before.jpg" "$AFTER_TEMP" "$TEMP_DIR/bedroom-before-cropped.jpg" "$TEMP_DIR/bedroom-after-cropped.jpg"
    
    # Move cropped images to marketing directory
    mv "$TEMP_DIR/bedroom-before-cropped.jpg" "$MARKETING_DIR/bedroom-before.jpg"
    mv "$TEMP_DIR/bedroom-after-cropped.jpg" "$MARKETING_DIR/bedroom-after.jpg"
    
    echo "  ✅ Bedroom images cropped and normalized"
else
    echo "  ⚠️ Bedroom images not found, skipping..."
fi

# Function to add subtle overlay guides (optional)
add_comparison_guides() {
    local file="$1"
    local output="$2"
    
    echo "  📐 Adding comparison guides to $file..."
    
    # Add subtle grid lines to help with before/after comparison
    magick "$file" \
        -stroke "rgba(255,255,255,0.1)" \
        -strokewidth 1 \
        -draw "line 0,33% 100%,33%" \
        -draw "line 0,66% 100%,66%" \
        -draw "line 33%,0 33%,100%" \
        -draw "line 66%,0 66%,100%" \
        "$output"
}

# Verify results
echo "🔍 Verifying cropped images..."

for image in "bedroom-before" "bedroom-after"; do
    if [[ -f "$MARKETING_DIR/${image}.jpg" ]]; then
        dims=$(get_dimensions "$MARKETING_DIR/${image}.jpg")
        size=$(du -h "$MARKETING_DIR/${image}.jpg" | cut -f1)
        echo "  ✅ ${image}.jpg: ${dims} (${size})"
    else
        echo "  ❌ ${image}.jpg: Not found"
    fi
done

# Clean up temp files
echo "🧹 Cleaning up temporary files..."
rm -rf "$TEMP_DIR"

echo ""
echo "🎉 Image cropping complete!"
echo ""
echo "📊 Results:"
echo "   - Images normalized to identical dimensions"
echo "   - Consistent JPEG format for optimal compatibility"
echo "   - Centered cropping to maintain important content"
echo "   - Optimized for web (quality 85, metadata stripped)"
echo ""
echo "💡 Tips:"
echo "   - Images now have perfect alignment for before/after slider"
echo "   - If cropping doesn't look right, adjust the -gravity parameter"
echo "   - Consider the focal point when choosing crop area"
echo "   - Test the slider at different screen sizes"
