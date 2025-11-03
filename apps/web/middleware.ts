import { NextRequest, NextResponse } from 'next/server';
import { auth0 } from './lib/auth0';

/**
 * Middleware to handle authentication.
 * 
 * Auth0 SDK v4 middleware automatically handles:
 * - /auth/login - Redirects to Auth0 Universal Login
 * - /auth/logout - Logs out and clears session
 * - /auth/callback - Handles Auth0 callback
 * - /auth/profile - Returns user profile
 * - /auth/access-token - Returns access token (if enabled)
 * - Session management and rolling sessions
 * - Protecting routes based on authentication
 */
export async function middleware(request: NextRequest) {
  const response = await auth0.middleware(request);
  
  // Protected routes that require authentication
  const userProtectedPaths = ['/upload', '/images', '/profile', '/billing'];
  const adminProtectedPaths = ['/admin'];
  const pathname = request.nextUrl.pathname;
  
  // Exclude marketing images from protection
  const isMarketingImage = pathname.startsWith('/images/marketing/');
  
  // Check if the current path is protected
  const isUserProtectedPath = !isMarketingImage && userProtectedPaths.some(path => pathname.startsWith(path));
  const isAdminProtectedPath = adminProtectedPaths.some(path => pathname.startsWith(path));
  
  if (isUserProtectedPath || isAdminProtectedPath) {
    // Get user session
    const session = await auth0.getSession(request);
    
    // If no session, handle based on route type
    if (!session) {
      if (isAdminProtectedPath) {
        // For admin routes, rewrite to 404 page to hide existence from unauthorized users
        const notFoundUrl = new URL('/not-found', request.url);
        return NextResponse.rewrite(notFoundUrl, { status: 404 });
      } else {
        // For user routes, redirect to login with returnTo
        const loginUrl = new URL('/auth/login', request.url);
        loginUrl.searchParams.set('returnTo', pathname);
        return NextResponse.redirect(loginUrl);
      }
    }
  }
  
  return response;
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico, sitemap.xml, robots.txt (metadata files)
     */
    '/((?!_next/static|_next/image|favicon.ico|sitemap.xml|robots.txt).*)',
  ],
};
