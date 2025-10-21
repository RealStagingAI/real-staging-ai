/**
 * Cloudflare Worker - Authenticated CDN for Private Images
 * 
 * Provides secure, authenticated access to private B2 images with:
 * - JWT token validation (Auth0)
 * - Image ownership verification
 * - Edge caching with user-specific cache keys
 * - AWS Signature V4 for B2 access
 */

interface Env {
	AUTH0_DOMAIN: string;
	AUTH0_AUDIENCE: string;
	API_BASE_URL: string;
	B2_BUCKET_NAME: string;
	B2_ENDPOINT: string;
	B2_REGION: string;
	B2_ACCESS_KEY_ID: string;
	B2_SECRET_ACCESS_KEY: string;
	WORKER_SECRET: string;
}

interface JWTPayload {
	sub: string;
	exp: number;
	aud: string;
	iss: string;
}

interface ImageOwnershipResponse {
	image_id: string;
	owner_id: string;
	has_access: boolean;
	s3_key?: string;
}

export default {
	async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
		// CORS preflight
		if (request.method === 'OPTIONS') {
			return handleCORS();
		}

		// Only handle GET requests
		if (request.method !== 'GET') {
			return jsonResponse({ error: 'Method not allowed' }, 405);
		}

		try {
			// 1. Extract and validate JWT token
			const authHeader = request.headers.get('Authorization');
			if (!authHeader?.startsWith('Bearer ')) {
				return jsonResponse(
					{ error: 'Unauthorized', message: 'Missing or invalid Authorization header' },
					401,
					{ 'WWW-Authenticate': 'Bearer realm="CDN"' }
				);
			}

			const token = authHeader.substring(7);

			// 2. Verify JWT with Auth0
			const user = await verifyAuth0Token(token, env);
			if (!user) {
				return jsonResponse({ error: 'Unauthorized', message: 'Invalid or expired token' }, 401);
			}

			// 3. Parse request URL
			const url = new URL(request.url);
			const pathParts = url.pathname.split('/').filter(Boolean);

			// Expected: /images/{imageId}/{kind}
			if (pathParts.length < 3 || pathParts[0] !== 'images') {
				return jsonResponse(
					{ error: 'Bad Request', message: 'Invalid path. Expected: /images/{imageId}/{kind}' },
					400
				);
			}

			const imageId = pathParts[1];
			const kind = pathParts[2];

			if (!['original', 'staged'].includes(kind)) {
				return jsonResponse({ error: 'Bad Request', message: 'kind must be "original" or "staged"' }, 400);
			}

			// 4. Check cache first (user-specific)
			const cacheKey = new Request(url.toString(), {
				method: 'GET',
				headers: { 'Authorization': authHeader }
			});
			const cache = caches.default;
			let cachedResponse = await cache.match(cacheKey);

			if (cachedResponse) {
				const headers = new Headers(cachedResponse.headers);
				headers.set('X-Cache-Status', 'HIT');
				headers.set('X-Worker-Version', '1.0');
				return new Response(cachedResponse.body, {
					status: cachedResponse.status,
					headers: addCORSHeaders(headers)
				});
			}

			// 5. Check ownership and get S3 key
			const ownershipData = await checkImageOwnership(user.sub, imageId, kind, env);
			if (!ownershipData.has_access) {
				return jsonResponse({ error: 'Forbidden', message: 'You do not have access to this image' }, 403);
			}

			if (!ownershipData.s3_key) {
				return jsonResponse({ error: 'Not Found', message: 'Image file not found' }, 404);
			}

			// 6. Fetch from private B2
			const b2Response = await fetchFromB2(ownershipData.s3_key, env);
			if (!b2Response.ok) {
				return jsonResponse(
					{ error: 'Not Found', message: 'Image not found in storage' },
					b2Response.status
				);
			}

			// 7. Create cacheable response
			const responseHeaders = new Headers();
			responseHeaders.set('Content-Type', b2Response.headers.get('Content-Type') || 'image/jpeg');
			responseHeaders.set('Cache-Control', 'private, max-age=3600'); // 1 hour
			responseHeaders.set('Vary', 'Authorization');
			responseHeaders.set('X-Cache-Status', 'MISS');
			responseHeaders.set('X-Worker-Version', '1.0');

			const response = new Response(b2Response.body, {
				status: 200,
				headers: addCORSHeaders(responseHeaders)
			});

			// 8. Cache for next time
			ctx.waitUntil(cache.put(cacheKey, response.clone()));

			return response;

		} catch (error) {
			console.error('Worker error:', error);
			return jsonResponse(
				{ error: 'Internal Server Error', message: error instanceof Error ? error.message : 'Unknown error' },
				500
			);
		}
	}
};

/**
 * Verify JWT token with Auth0 using JWKS
 */
async function verifyAuth0Token(token: string, env: Env): Promise<JWTPayload | null> {
	try {
		// Decode JWT parts
		const parts = token.split('.');
		if (parts.length !== 3) {
			return null;
		}

		const [headerB64, payloadB64, signatureB64] = parts;

		// Decode header and payload
		const header = JSON.parse(atob(headerB64.replace(/-/g, '+').replace(/_/g, '/')));
		const payload: JWTPayload = JSON.parse(atob(payloadB64.replace(/-/g, '+').replace(/_/g, '/')));

		// Verify expiry
		const now = Math.floor(Date.now() / 1000);
		if (payload.exp < now) {
			console.log('Token expired');
			return null;
		}

		// Verify audience
		if (payload.aud !== env.AUTH0_AUDIENCE) {
			console.log('Invalid audience');
			return null;
		}

		// Fetch JWKS
		const jwksUrl = `https://${env.AUTH0_DOMAIN}/.well-known/jwks.json`;
		const jwksResponse = await fetch(jwksUrl);
		if (!jwksResponse.ok) {
			console.error('Failed to fetch JWKS');
			return null;
		}

		const jwks = await jwksResponse.json();
		const key = jwks.keys.find((k: any) => k.kid === header.kid);
		if (!key) {
			console.log('Key not found in JWKS');
			return null;
		}

		// Import public key
		const publicKey = await crypto.subtle.importKey(
			'jwk',
			{
				kty: key.kty,
				n: key.n,
				e: key.e,
				alg: key.alg,
				use: key.use
			},
			{ name: 'RSASSA-PKCS1-v1_5', hash: 'SHA-256' },
			false,
			['verify']
		);

		// Verify signature
		const dataToVerify = new TextEncoder().encode(`${headerB64}.${payloadB64}`);
		const signature = base64UrlToArrayBuffer(signatureB64);

		const isValid = await crypto.subtle.verify(
			'RSASSA-PKCS1-v1_5',
			publicKey,
			signature,
			dataToVerify
		);

		if (!isValid) {
			console.log('Invalid signature');
			return null;
		}

		return payload;

	} catch (error) {
		console.error('Token verification error:', error);
		return null;
	}
}

/**
 * Check if user owns the image and get S3 key
 */
async function checkImageOwnership(
	userId: string,
	imageId: string,
	kind: string,
	env: Env
): Promise<ImageOwnershipResponse> {
	try {
		const response = await fetch(`${env.API_BASE_URL}/v1/images/${imageId}/owner`, {
			headers: {
				'X-User-ID': userId,
				'X-Image-Kind': kind,
				'X-Internal-Auth': env.WORKER_SECRET
			}
		});

		if (!response.ok) {
			return { image_id: imageId, owner_id: '', has_access: false };
		}

		return await response.json();
	} catch (error) {
		console.error('Ownership check error:', error);
		return { image_id: imageId, owner_id: '', has_access: false };
	}
}

/**
 * Fetch image from private B2 bucket using AWS Signature V4
 */
async function fetchFromB2(s3Key: string, env: Env): Promise<Response> {
	const url = `${env.B2_ENDPOINT}/${env.B2_BUCKET_NAME}/${s3Key}`;
	const host = new URL(env.B2_ENDPOINT).host;

	// Create signed request
	const signedRequest = await signAWSRequest(
		'GET',
		url,
		host,
		env.B2_REGION,
		's3',
		env.B2_ACCESS_KEY_ID,
		env.B2_SECRET_ACCESS_KEY
	);

	return fetch(signedRequest);
}

/**
 * Sign AWS request using Signature V4
 */
async function signAWSRequest(
	method: string,
	url: string,
	host: string,
	region: string,
	service: string,
	accessKeyId: string,
	secretAccessKey: string
): Promise<Request> {
	const urlObj = new URL(url);
	const now = new Date();
	const dateStamp = now.toISOString().split('T')[0].replace(/-/g, '');
	const amzDate = now.toISOString().replace(/[:-]|\.\d{3}/g, '');

	// Canonical request
	const canonicalUri = urlObj.pathname;
	const canonicalQuerystring = urlObj.search.substring(1);
	const canonicalHeaders = `host:${host}\nx-amz-content-sha256:UNSIGNED-PAYLOAD\nx-amz-date:${amzDate}\n`;
	const signedHeaders = 'host;x-amz-content-sha256;x-amz-date';
	const payloadHash = 'UNSIGNED-PAYLOAD';

	const canonicalRequest = `${method}\n${canonicalUri}\n${canonicalQuerystring}\n${canonicalHeaders}\n${signedHeaders}\n${payloadHash}`;

	// String to sign
	const algorithm = 'AWS4-HMAC-SHA256';
	const credentialScope = `${dateStamp}/${region}/${service}/aws4_request`;
	const canonicalRequestHash = await sha256(canonicalRequest);
	const stringToSign = `${algorithm}\n${amzDate}\n${credentialScope}\n${canonicalRequestHash}`;

	// Signing key
	const signingKey = await getSignatureKey(secretAccessKey, dateStamp, region, service);
	const signature = await hmacSha256(signingKey, stringToSign);

	// Authorization header
	const authorization = `${algorithm} Credential=${accessKeyId}/${credentialScope}, SignedHeaders=${signedHeaders}, Signature=${signature}`;

	// Create request
	const headers = new Headers({
		'Host': host,
		'x-amz-date': amzDate,
		'x-amz-content-sha256': payloadHash,
		'Authorization': authorization
	});

	return new Request(url, { method, headers });
}

/**
 * AWS Signature V4 key derivation
 */
async function getSignatureKey(key: string, dateStamp: string, regionName: string, serviceName: string): Promise<ArrayBuffer> {
	const kDate = await hmacSha256Raw(new TextEncoder().encode(`AWS4${key}`), dateStamp);
	const kRegion = await hmacSha256Raw(kDate, regionName);
	const kService = await hmacSha256Raw(kRegion, serviceName);
	return hmacSha256Raw(kService, 'aws4_request');
}

/**
 * HMAC-SHA256 (returns hex string)
 */
async function hmacSha256(key: ArrayBuffer, data: string): Promise<string> {
	const cryptoKey = await crypto.subtle.importKey('raw', key, { name: 'HMAC', hash: 'SHA-256' }, false, ['sign']);
	const signature = await crypto.subtle.sign('HMAC', cryptoKey, new TextEncoder().encode(data));
	return arrayBufferToHex(signature);
}

/**
 * HMAC-SHA256 (returns ArrayBuffer)
 */
async function hmacSha256Raw(key: ArrayBuffer | Uint8Array, data: string): Promise<ArrayBuffer> {
	const cryptoKey = await crypto.subtle.importKey('raw', key, { name: 'HMAC', hash: 'SHA-256' }, false, ['sign']);
	return crypto.subtle.sign('HMAC', cryptoKey, new TextEncoder().encode(data));
}

/**
 * SHA-256 hash (returns hex string)
 */
async function sha256(data: string): Promise<string> {
	const hash = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(data));
	return arrayBufferToHex(hash);
}

/**
 * Convert ArrayBuffer to hex string
 */
function arrayBufferToHex(buffer: ArrayBuffer): string {
	return Array.from(new Uint8Array(buffer))
		.map(b => b.toString(16).padStart(2, '0'))
		.join('');
}

/**
 * Convert base64url to ArrayBuffer
 */
function base64UrlToArrayBuffer(base64url: string): ArrayBuffer {
	const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/');
	const padding = '='.repeat((4 - (base64.length % 4)) % 4);
	const base64Padded = base64 + padding;
	const binary = atob(base64Padded);
	const bytes = new Uint8Array(binary.length);
	for (let i = 0; i < binary.length; i++) {
		bytes[i] = binary.charCodeAt(i);
	}
	return bytes.buffer;
}

/**
 * JSON response helper
 */
function jsonResponse(data: any, status = 200, extraHeaders: Record<string, string> = {}): Response {
	const headers = new Headers({
		'Content-Type': 'application/json',
		...extraHeaders
	});
	return new Response(JSON.stringify(data), {
		status,
		headers: addCORSHeaders(headers)
	});
}

/**
 * Add CORS headers
 */
function addCORSHeaders(headers: Headers): Headers {
	headers.set('Access-Control-Allow-Origin', '*'); // Update to specific domain in production
	headers.set('Access-Control-Allow-Methods', 'GET, OPTIONS');
	headers.set('Access-Control-Allow-Headers', 'Authorization, Content-Type');
	headers.set('Access-Control-Max-Age', '86400');
	return headers;
}

/**
 * Handle CORS preflight
 */
function handleCORS(): Response {
	return new Response(null, {
		status: 204,
		headers: addCORSHeaders(new Headers())
	});
}
