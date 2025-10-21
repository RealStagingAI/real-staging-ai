import { getAccessToken } from '@auth0/nextjs-auth0';
import { NextRequest, NextResponse } from 'next/server';

/**
 * CDN Proxy - Adds Authorization header to requests to Cloudflare CDN Worker
 * 
 * This endpoint proxies image requests to the CDN worker, adding the required
 * Authorization header that Next.js Image component can't provide directly.
 * 
 * GET /api/cdn/{imageId}/{kind}
 * - imageId: UUID of the image
 * - kind: 'original' or 'staged'
 */
export async function GET(
  request: NextRequest,
  { params }: { params: { imageId: string; kind: string } }
) {
  const CDN_URL = process.env.NEXT_PUBLIC_CDN_URL;
  
  if (!CDN_URL) {
    return NextResponse.json(
      { error: 'CDN not configured' },
      { status: 503 }
    );
  }

  try {
    // Get access token from session
    const accessToken = await getAccessToken();
    
    if (!accessToken) {
      return NextResponse.json(
        { error: 'Unauthorized' },
        { status: 401 }
      );
    }

    const { imageId, kind } = params;
    
    // Validate kind parameter
    if (kind !== 'original' && kind !== 'staged') {
      return NextResponse.json(
        { error: 'Invalid kind parameter. Must be "original" or "staged"' },
        { status: 400 }
      );
    }

    // Fetch from CDN with Authorization header
    const cdnUrl = `${CDN_URL}/images/${imageId}/${kind}`;
    const cdnResponse = await fetch(cdnUrl, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    });

    if (!cdnResponse.ok) {
      return NextResponse.json(
        { error: `CDN request failed: ${cdnResponse.statusText}` },
        { status: cdnResponse.status }
      );
    }

    // Get image data
    const imageData = await cdnResponse.arrayBuffer();
    const contentType = cdnResponse.headers.get('content-type') || 'image/jpeg';
    const cacheStatus = cdnResponse.headers.get('x-cache-status');

    // Return image with appropriate headers
    return new NextResponse(imageData, {
      headers: {
        'Content-Type': contentType,
        'Cache-Control': 'private, max-age=3600', // Cache for 1 hour
        'X-CDN-Cache-Status': cacheStatus || 'UNKNOWN',
      },
    });
  } catch (error) {
    console.error('CDN proxy error:', error);
    return NextResponse.json(
      { error: 'Failed to fetch image from CDN' },
      { status: 500 }
    );
  }
}
