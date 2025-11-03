#!/bin/bash

# Crop and normalize ALL marketing images to match perfectly
# This script processes all before/after image pairs in the marketing directory
# Usage: ./scripts/crop-all-marketing-images.sh

set -e

MARKETING_DIR="apps/web/public/images/marketing"
TEMP_DIR="/tmp/real-staging-crop-all"

echo "🔧 Setting up image cropping workspace for all marketing images..."
mkdir -p "$TEMP_DIR"
mkdir -p "$MARKETING_DIR"

# Check if ImageMagick is available
if ! command -v magick &> /dev/null; then
    echo "❌ ImageMagick is required but not installed."
    echo "   brew install imagemagick"
    exit 1
fi

# Function to normalize both images to a common size
normalize_to_common_size() {
    local before="$1"
    local after="$2"
    local output_before="$3"
    local output_after="$4"
    local room_name="$5"
    
    echo "  📐 Normalizing $room_name images to common dimensions..."
    
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

# Function to get image dimensions
get_dimensions() {
    local file="$1"
    magick identify -format "%wx%h" "$file" 2>/dev/null || echo "unknown"
}

# Function to verify results
verify_images() {
    echo "🔍 Verifying processed images..."
    
    for image in "bedroom-before" "bedroom-after" "living-room-before" "living-room-after"; do
        if [[ -f "$MARKETING_DIR/${image}.jpg" ]]; then
            dims=$(get_dimensions "$MARKETING_DIR/${image}.jpg")
            size=$(du -h "$MARKETING_DIR/${image}.jpg" | cut -f1)
            echo "  ✅ ${image}.jpg: ${dims} (${size})"
        else
            echo "  ⚠️ ${image}.jpg: Not found"
        fi
    done
}

# Find all image pairs in the marketing directory
echo "🔍 Scanning for marketing image pairs..."

# Array to store room names
declare -a room_names=()

# Find all before images and extract room names
for before_file in "$MARKETING_DIR"/*-before.jpg; do
    if [[ -f "$before_file" ]]; then
        basename=$(basename "$before_file" .jpg)
        room_name=${basename%-before}
        after_file="$MARKETING_DIR/${room_name}-after.jpg"
        
        if [[ -f "$after_file" ]]; then
            room_names+=("$room_name")
            echo "  📷 Found pair: $room_name"
        else
            echo "  ⚠️ Missing after image for: $room_name"
        fi
    fi
done

if [[ ${#room_names[@]} -eq 0 ]]; then
    echo "❌ No image pairs found in $MARKETING_DIR"
    exit 1
fi

echo ""
echo "🛠️ Processing ${#room_names[@]} image pair(s)..."

# Process each room
for room_name in "${room_names[@]}"; do
    echo ""
    echo "🏠 Processing $room_name transformation..."
    
    before_file="$MARKETING_DIR/${room_name}-before.jpg"
    after_file="$MARKETING_DIR/${room_name}-after.jpg"
    
    if [[ -f "$before_file" && -f "$after_file" ]]; then
        # Create backup originals
        cp "$before_file" "$TEMP_DIR/${room_name}-before-original.jpg"
        cp "$after_file" "$TEMP_DIR/${room_name}-after-original.jpg"
        
        # Convert after image to JPEG if it's WebP
        if file "$after_file" | grep -q "Web/P"; then
            echo "  🔄 Converting after image from WebP to JPEG..."
            magick "$after_file" "$TEMP_DIR/${room_name}-after-temp.jpg"
            AFTER_TEMP="$TEMP_DIR/${room_name}-after-temp.jpg"
        else
            cp "$after_file" "$TEMP_DIR/${room_name}-after-temp.jpg"
            AFTER_TEMP="$TEMP_DIR/${room_name}-after-temp.jpg"
        fi
        
        # Normalize both images to identical dimensions
        echo "  🎯 Normalizing both images to standard dimensions..."
        normalize_to_common_size "$before_file" "$AFTER_TEMP" "$TEMP_DIR/${room_name}-before-cropped.jpg" "$TEMP_DIR/${room_name}-after-cropped.jpg" "$room_name"
        
        # Move cropped images to marketing directory
        mv "$TEMP_DIR/${room_name}-before-cropped.jpg" "$MARKETING_DIR/${room_name}-before.jpg"
        mv "$TEMP_DIR/${room_name}-after-cropped.jpg" "$MARKETING_DIR/${room_name}-after.jpg"
        
        echo "  ✅ $room_name images cropped and normalized"
    else
        echo "  ❌ Missing files for $room_name"
    fi
done

# Verify results
echo ""
verify_images

# Clean up temp files
echo ""
echo "🧹 Cleaning up temporary files..."
rm -rf "$TEMP_DIR"

echo ""
echo "🎉 All marketing images cropped and normalized!"
echo ""
echo "📊 Results:"
echo "   - ${#room_names[@]} room(s) processed"
echo "   - Images normalized to identical 1200x900 dimensions"
echo "   - Consistent JPEG format for optimal compatibility"
echo "   - Centered cropping to maintain important content"
echo "   - Optimized for web (quality 85, metadata stripped)"
echo ""
echo "🎯 Your landing page will now show perfectly aligned before/after sliders!"
