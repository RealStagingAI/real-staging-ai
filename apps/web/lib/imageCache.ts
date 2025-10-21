/**
 * Image URL caching utilities
 * Caches presigned URLs in localStorage to avoid redundant API calls
 */

const CACHE_KEY = 'realstaging_image_urls';
const CACHE_TTL = 3600000; // 1 hour in milliseconds

export interface CachedImageUrl {
  url: string;
  expiry: number;
}

export interface ImageUrlCache {
  [key: string]: CachedImageUrl;
}

/**
 * Get cached presigned URL if available and not expired
 */
export function getCachedUrl(imageId: string, kind: 'original' | 'staged'): string | null {
  if (typeof window === 'undefined') return null;
  
  try {
    const cacheStr = localStorage.getItem(CACHE_KEY);
    if (!cacheStr) return null;
    
    const cache: ImageUrlCache = JSON.parse(cacheStr);
    const key = `${imageId}_${kind}`;
    const cached = cache[key];
    
    if (cached && Date.now() < cached.expiry) {
      return cached.url;
    }
    
    // Clean up expired entry
    if (cached) {
      delete cache[key];
      localStorage.setItem(CACHE_KEY, JSON.stringify(cache));
    }
    
    return null;
  } catch (error) {
    console.error('Error reading image cache:', error);
    return null;
  }
}

/**
 * Cache a presigned URL with expiry
 */
export function setCachedUrl(imageId: string, kind: 'original' | 'staged', url: string): void {
  if (typeof window === 'undefined') return;
  
  try {
    const cacheStr = localStorage.getItem(CACHE_KEY);
    const cache: ImageUrlCache = cacheStr ? JSON.parse(cacheStr) : {};
    
    const key = `${imageId}_${kind}`;
    cache[key] = {
      url,
      expiry: Date.now() + CACHE_TTL
    };
    
    localStorage.setItem(CACHE_KEY, JSON.stringify(cache));
  } catch (error) {
    console.error('Error writing image cache:', error);
    // If quota exceeded, clear old cache and try again
    if (error instanceof DOMException && error.name === 'QuotaExceededError') {
      clearExpiredCache();
      try {
        const cache: ImageUrlCache = {};
        cache[`${imageId}_${kind}`] = {
          url,
          expiry: Date.now() + CACHE_TTL
        };
        localStorage.setItem(CACHE_KEY, JSON.stringify(cache));
      } catch {
        // Silently fail if still can't cache
      }
    }
  }
}

/**
 * Clear expired entries from cache
 */
export function clearExpiredCache(): void {
  if (typeof window === 'undefined') return;
  
  try {
    const cacheStr = localStorage.getItem(CACHE_KEY);
    if (!cacheStr) return;
    
    const cache: ImageUrlCache = JSON.parse(cacheStr);
    const now = Date.now();
    let hasChanges = false;
    
    Object.keys(cache).forEach(key => {
      if (cache[key].expiry < now) {
        delete cache[key];
        hasChanges = true;
      }
    });
    
    if (hasChanges) {
      localStorage.setItem(CACHE_KEY, JSON.stringify(cache));
    }
  } catch (error) {
    console.error('Error clearing expired cache:', error);
  }
}

/**
 * Clear all cached URLs
 */
export function clearCache(): void {
  if (typeof window === 'undefined') return;
  
  try {
    localStorage.removeItem(CACHE_KEY);
  } catch (error) {
    console.error('Error clearing cache:', error);
  }
}

/**
 * Get cache statistics
 */
export function getCacheStats(): { count: number; size: number } {
  if (typeof window === 'undefined') return { count: 0, size: 0 };
  
  try {
    const cacheStr = localStorage.getItem(CACHE_KEY);
    if (!cacheStr) return { count: 0, size: 0 };
    
    const cache: ImageUrlCache = JSON.parse(cacheStr);
    const count = Object.keys(cache).length;
    const size = new Blob([cacheStr]).size;
    
    return { count, size };
  } catch (error) {
    return { count: 0, size: 0 };
  }
}
