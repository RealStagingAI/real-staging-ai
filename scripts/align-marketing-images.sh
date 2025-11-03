#!/bin/bash

# Advanced image alignment for perfect before/after comparison
# This script provides multiple alignment strategies for pixel-perfect results
# Usage: ./scripts/align-marketing-images.sh [strategy]

set -e

MARKETING_DIR="apps/web/public/images/marketing"
TEMP_DIR="/tmp/real-staging-align"

# Alignment strategies
STRATEGY="${1:-manual}"

echo "🎯 Advanced Image Alignment for Perfect Before/After"
echo "Strategy: $STRATEGY"
echo ""

# Check dependencies
if ! command -v magick &> /dev/null; then
    echo "❌ ImageMagick is required but not installed."
    echo "   brew install imagemagick"
    exit 1
fi

mkdir -p "$TEMP_DIR"
mkdir -p "$MARKETING_DIR"

# Function to show image info
show_image_info() {
    local file="$1"
    local name="$2"
    
    if [[ -f "$file" ]]; then
        local dims=$(magick identify -format "%wx%h" "$file" 2>/dev/null)
        local size=$(du -h "$file" | cut -f1)
        local format=$(file "$file" | cut -d: -f2 | cut -d, -f1)
        echo "  📷 $name: $dims ($size) $format"
    else
        echo "  ❌ $name: Not found"
    fi
}

# Function to create alignment overlay
create_alignment_overlay() {
    local before="$1"
    local after="$2"
    local output="$3"
    
    echo "  🔍 Creating alignment analysis overlay..."
    
    # Create a side-by-side comparison with alignment guides
    magick "$before" "$after" \
        +append \
        -stroke red \
        -strokewidth 2 \
        -fill none \
        -draw "rectangle 0,0 599,899" \
        -draw "rectangle 600,0 1199,899" \
        -draw "line 300,0 300,899" \
        -draw "line 600,0 600,899" \
        -draw "line 900,0 900,899" \
        -draw "line 0,450 599,450" \
        -draw "line 600,450 1199,450" \
        "$output"
}

# Function for manual alignment with visual guides
manual_alignment() {
    local before="$1"
    local after="$2"
    
    echo "🎨 Manual Alignment Mode"
    echo "This will create visual guides to help you align images manually"
    echo ""
    
    # Create alignment overlay
    create_alignment_overlay "$before" "$after" "$TEMP_DIR/alignment-guide.jpg"
    
    echo "📋 Alignment Guide Created: $TEMP_DIR/alignment-guide.jpg"
    echo ""
    echo "🔧 Manual Alignment Steps:"
    echo "1. Open the alignment guide image"
    echo "2. Look for misaligned elements (walls, furniture, windows)"
    echo "3. Use photo editing software to adjust one image to match the other"
    echo "4. Focus on aligning permanent architectural features"
    echo ""
    echo "💡 Tips for manual alignment:"
    echo "   - Align walls, corners, and permanent fixtures"
    echo "   - Don't worry about movable furniture (it should be different!)"
    echo "   - Use the grid lines to check alignment"
    echo "   - Save the aligned images with the same names"
    
    # Open the alignment guide if on macOS
    if command -v open &> /dev/null; then
        echo "🖼️ Opening alignment guide..."
        open "$TEMP_DIR/alignment-guide.jpg"
    fi
}

# Function for automatic alignment using feature detection
auto_alignment() {
    local before="$1"
    local after="$2"
    
    echo "⚡ Automatic Alignment Mode"
    echo "Attempting to automatically align images based on common features..."
    echo ""
    
    # This is a simplified auto-alignment - for true feature detection
    # you'd need more advanced tools like OpenCV
    
    # Try different gravity positions for better alignment
    local gravities=("center" "north" "south" "east" "west" "northeast" "northwest" "southeast" "southwest")
    local target_width="1200"
    local target_height="900"
    
    echo "🔄 Testing different alignment positions..."
    
    for gravity in "${gravities[@]}"; do
        echo "  Testing gravity: $gravity"
        
        magick "$before" \
            -gravity "$gravity" \
            -resize "${target_width}x${target_height}^" \
            -extent "${target_width}x${target_height}" \
            -quality 85 \
            -strip \
            "$TEMP_DIR/before-${gravity}.jpg"
            
        magick "$after" \
            -gravity "$gravity" \
            -resize "${target_width}x${target_height}^" \
            -extent "${target_width}x${target_height}" \
            -quality 85 \
            -strip \
            "$TEMP_DIR/after-${gravity}.jpg"
            
        # Create comparison for this gravity
        create_alignment_overlay "$TEMP_DIR/before-${gravity}.jpg" "$TEMP_DIR/after-${gravity}.jpg" "$TEMP_DIR/comparison-${gravity}.jpg"
    done
    
    echo ""
    echo "📊 Alignment comparisons created in $TEMP_DIR:"
    ls -la "$TEMP_DIR"/comparison-*.jpg
    
    echo ""
    echo "🎯 Review the comparison images and choose the best alignment:"
    echo "   Look for the one where architectural features align best"
    echo ""
    
    # Create a montage of all options for easy comparison
    magick montage "$TEMP_DIR"/comparison-*.jpg \
        -tile 3x3 \
        -geometry +5+5 \
        "$TEMP_DIR/alignment-options.jpg"
        
    echo "🖼️ Created montage: $TEMP_DIR/alignment-options.jpg"
    
    if command -v open &> /dev/null; then
        open "$TEMP_DIR/alignment-options.jpg"
    fi
}

# Function for perspective correction
perspective_correction() {
    local before="$1"
    local after="$2"
    
    echo "📐 Perspective Correction Mode"
    echo "This helps if photos were taken from slightly different angles"
    echo ""
    
    echo "🔧 Perspective Correction Steps:"
    echo "1. Open both images in photo editing software (Photoshop, GIMP, etc.)"
    echo "2. Use perspective transform/perspective crop tool"
    echo "3. Align permanent architectural features:"
    echo "   - Wall corners and edges"
    echo "   - Door and window frames"
    echo "   - Ceiling and floor lines"
    echo "4. Export both images with identical dimensions (1200x900 recommended)"
    echo ""
    
    # Create reference guides
    magick "$before" \
        -stroke red \
        -strokewidth 1 \
        -draw "line 0,0 1200,900" \
        -draw "line 0,900 1200,0" \
        -draw "line 600,0 600,900" \
        -draw "line 0,450 1200,450" \
        "$TEMP_DIR/before-guides.jpg"
        
    magick "$after" \
        -stroke blue \
        -strokewidth 1 \
        -draw "line 0,0 1200,900" \
        -draw "line 0,900 1200,0" \
        -draw "line 600,0 600,900" \
        -draw "line 0,450 1200,450" \
        "$TEMP_DIR/after-guides.jpg"
    
    echo "📐 Created reference guides with alignment lines:"
    echo "   Red lines: Before image"
    echo "   Blue lines: After image"
    echo ""
    echo "Files created in $TEMP_DIR:"
    echo "  - before-guides.jpg (with red alignment lines)"
    echo "  - after-guides.jpg (with blue alignment lines)"
    
    if command -v open &> /dev/null; then
        echo "🖼️ Opening reference guides..."
        open "$TEMP_DIR/before-guides.jpg"
        open "$TEMP_DIR/after-guides.jpg"
    fi
}

# Main execution
echo "📊 Current Image Status:"
show_image_info "$MARKETING_DIR/bedroom-before.jpg" "Before"
show_image_info "$MARKETING_DIR/bedroom-after.jpg" "After"
echo ""

case "$STRATEGY" in
    "manual")
        manual_alignment "$MARKETING_DIR/bedroom-before.jpg" "$MARKETING_DIR/bedroom-after.jpg"
        ;;
    "auto")
        auto_alignment "$MARKETING_DIR/bedroom-before.jpg" "$MARKETING_DIR/bedroom-after.jpg"
        ;;
    "perspective")
        perspective_correction "$MARKETING_DIR/bedroom-before.jpg" "$MARKETING_DIR/bedroom-after.jpg"
        ;;
    *)
        echo "❌ Unknown strategy: $STRATEGY"
        echo ""
        echo "Available strategies:"
        echo "  manual      - Create visual guides for manual alignment"
        echo "  auto        - Test automatic alignment positions"
        echo "  perspective - Create guides for perspective correction"
        echo ""
        echo "Usage: $0 [manual|auto|perspective]"
        exit 1
        ;;
esac

echo ""
echo "🎯 Next Steps:"
echo "1. Review the generated alignment guides"
echo "2. Use photo editing software to perfect the alignment"
echo "3. Replace the original images with your aligned versions"
echo "4. Test the before/after slider"
echo ""
echo "💡 Professional Tips:"
echo "   - Focus on aligning permanent features (walls, windows, doors)"
echo "   - Don't worry about furniture alignment (it should be different!)"
echo "   - Use the same camera position and height for future photos"
echo "   - Consider using a tripod for consistent shots"
