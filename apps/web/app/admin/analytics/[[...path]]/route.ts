import { NextRequest, NextResponse } from 'next/server';
import { auth0 } from '@/lib/auth0';

/**
 * Proxy route for Metabase Analytics Dashboard
 * 
 * Proxies all requests to the internal Metabase service.
 * Requires authentication and admin role.
 * 
 * Routes:
 * - /admin/analytics -> Metabase root
 * - /admin/analytics/** -> Proxied to Metabase
 */

const METABASE_URL = process.env.METABASE_INTERNAL_URL || 'http://localhost:3001';

export async function GET(
  request: NextRequest,
  { params }: { params: { path?: string[] } }
) {
  return handleProxy(request, params);
}

export async function POST(
  request: NextRequest,
  { params }: { params: { path?: string[] } }
) {
  return handleProxy(request, params);
}

export async function PUT(
  request: NextRequest,
  { params }: { params: { path?: string[] } }
) {
  return handleProxy(request, params);
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: { path?: string[] } }
) {
  return handleProxy(request, params);
}

export async function PATCH(
  request: NextRequest,
  { params }: { params: { path?: string[] } }
) {
  return handleProxy(request, params);
}

async function handleProxy(
  request: NextRequest,
  { path }: { path?: string[] }
) {
  // Check authentication
  const session = await auth0.getSession();
  
  if (!session || !session.user) {
    return NextResponse.json(
      { error: 'Unauthorized' },
      { status: 401 }
    );
  }

  // TODO: Add role-based access control (RBAC)
  // For now, all authenticated users can access
  // Future: Check if user has admin role
  // const userRole = session.user['https://real-staging.ai/roles'];
  // if (!userRole?.includes('admin')) {
  //   return NextResponse.json(
  //     { error: 'Forbidden: Admin access required' },
  //     { status: 403 }
  //   );
  // }

  // Build target URL
  // Strip '/app' prefix if present (used for iframe routing)
  let pathSegments = path || [];
  if (pathSegments[0] === 'app') {
    pathSegments = pathSegments.slice(1);
  }
  
  const pathString = pathSegments.join('/');
  const searchParams = request.nextUrl.searchParams.toString();
  const targetUrl = `${METABASE_URL}/${pathString}${searchParams ? `?${searchParams}` : ''}`;

  try {
    // Prepare headers (exclude host-specific headers)
    const headers = new Headers();
    request.headers.forEach((value, key) => {
      // Skip host-specific headers that shouldn't be forwarded
      if (!['host', 'connection', 'x-forwarded-for', 'x-real-ip'].includes(key.toLowerCase())) {
        headers.set(key, value);
      }
    });

    // Add X-Forwarded headers for Metabase
    headers.set('X-Forwarded-Host', request.headers.get('host') || '');
    headers.set('X-Forwarded-Proto', 'https');
    headers.set('X-Forwarded-For', request.headers.get('x-forwarded-for') || request.headers.get('x-real-ip') || '');

    // Prepare request options
    const fetchOptions: RequestInit = {
      method: request.method,
      headers,
      redirect: 'manual', // Handle redirects manually
    };

    // Add body for non-GET requests
    if (request.method !== 'GET' && request.method !== 'HEAD') {
      fetchOptions.body = await request.arrayBuffer();
    }

    // Proxy request to Metabase
    const response = await fetch(targetUrl, fetchOptions);

    // Handle redirects
    if (response.status >= 300 && response.status < 400) {
      const location = response.headers.get('location');
      if (location) {
        // Rewrite location to point back to our proxy
        const newLocation = location.startsWith('/')
          ? `/admin/analytics${location}`
          : location;
        return NextResponse.redirect(new URL(newLocation, request.url));
      }
    }

    // Prepare response headers
    const responseHeaders = new Headers();
    response.headers.forEach((value, key) => {
      // Skip problematic headers and iframe-blocking headers
      const skipHeaders = [
        'transfer-encoding',
        'connection',
        'keep-alive',
        'x-frame-options',  // Remove to allow iframe embedding
        'content-security-policy'  // Remove CSP that blocks framing
      ];
      if (!skipHeaders.includes(key.toLowerCase())) {
        responseHeaders.set(key, value);
      }
    });

    // Return proxied response
    return new NextResponse(response.body, {
      status: response.status,
      statusText: response.statusText,
      headers: responseHeaders,
    });
  } catch (error) {
    console.error('Metabase proxy error:', error);
    return NextResponse.json(
      { error: 'Failed to connect to analytics service' },
      { status: 502 }
    );
  }
}
