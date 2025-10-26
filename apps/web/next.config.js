/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  images: {
    remotePatterns: [
      // Local development - MinIO
      {
        protocol: 'http',
        hostname: 'localhost',
        port: '9000',
        pathname: '/**',
      },
      // Local development - LocalStack
      {
        protocol: 'http',
        hostname: 'localhost',
        port: '4566',
        pathname: '/**',
      },
      // AWS S3
      {
        protocol: 'https',
        hostname: '**.s3.amazonaws.com',
        pathname: '/**',
      },
      {
        protocol: 'https',
        hostname: 's3.amazonaws.com',
        pathname: '/**',
      },
      // Backblaze B2
      {
        protocol: 'https',
        hostname: '**.backblazeb2.com',
        pathname: '/**',
      },
      {
        protocol: 'https',
        hostname: 's3.us-west-004.backblazeb2.com',
        pathname: '/**',
      },
    ],
  },
  async rewrites() {
    // Use environment variable for API URL, fallback to localhost for dev
    // API_URL can be either a full URL or just a hostname (from Render's fromService.property: hostport)
    const apiHost = process.env.API_URL || 'http://localhost:8080';
    const apiUrl = apiHost.startsWith('http')
      ? apiHost
      : `http://${apiHost}`;  // Internal Render services use HTTP, not HTTPS
    
    return [
      {
        source: '/api/:path*',
        destination: `${apiUrl}/api/:path*`,
      },
    ];
  },
};

module.exports = nextConfig;
