/**
 * Marketing image configuration for before/after sliders
 * Uses local images for consistent branding and performance
 */

export interface MarketingImage {
  beforeSrc: string;
  afterSrc: string;
  beforeAlt: string;
  afterAlt: string;
  title: string;
}

// Marketing images using local assets
export const marketingImages: MarketingImage[] = [
  {
    title: "Bedroom Transformation",
    beforeSrc: "/images/marketing/bedroom-before.jpg",
    afterSrc: "/images/marketing/bedroom-after.jpg",
    beforeAlt: "Empty bedroom with neutral walls",
    afterAlt: "Professionally staged bedroom with cozy furnishings"
  },
  {
    title: "Living Room Transformation",
    beforeSrc: "/images/marketing/living-room-before.jpg",
    afterSrc: "/images/marketing/living-room-after.jpg",
    beforeAlt: "Empty living room with hardwood floors",
    afterAlt: "Professionally staged living room with modern furniture"
  },
  {
    title: "Kitchen Renovation",
    beforeSrc: "/images/marketing/kitchen-before.jpg",
    afterSrc: "/images/marketing/kitchen-after.jpg",
    beforeAlt: "Basic kitchen with outdated appliances",
    afterAlt: "Modern staged kitchen with premium appliances"
  },
  {
    title: "Home Office Setup",
    beforeSrc: "/images/marketing/office-before.jpg",
    afterSrc: "/images/marketing/office-after.jpg",
    beforeAlt: "Empty room perfect for home office",
    afterAlt: "Productively staged home office space"
  }
];

// Helper function to get marketing images
export function getMarketingImages(): MarketingImage[] {
  return marketingImages;
}
