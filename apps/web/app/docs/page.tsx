import { BookOpen, Upload, Wand2, CreditCard, HelpCircle } from 'lucide-react';
import Link from 'next/link';

export const metadata = {
  title: 'How to Use Real Staging AI | Documentation',
  description: 'Learn how to transform your property photos with AI-powered virtual staging. Simple guides for real estate agents and photographers.',
};

export default function DocsPage() {
  return (
    <div className="mx-auto max-w-4xl">
      {/* Header */}
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-5xl mb-4">
          How to Use Real Staging AI
        </h1>
        <p className="text-lg text-gray-600 dark:text-gray-400">
          Transform empty rooms into beautifully staged spaces in minutes
        </p>
      </div>

      {/* Quick Start Guide */}
      <div className="mb-16">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6 flex items-center gap-2">
          <Wand2 className="h-6 w-6 text-indigo-600 dark:text-indigo-400" />
          Quick Start Guide
        </h2>
        
        <div className="space-y-6">
          {/* Step 1 */}
          <div className="border border-gray-200 dark:border-gray-800 rounded-lg p-6 bg-white dark:bg-slate-900">
            <div className="flex items-start gap-4">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-indigo-100 dark:bg-indigo-900 text-indigo-600 dark:text-indigo-400 font-semibold">
                1
              </div>
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
                  Sign Up & Choose Your Plan
                </h3>
                <p className="text-gray-600 dark:text-gray-400 mb-3">
                  Create your account and select a plan that fits your needs. Free tier includes 10 staged images per month.
                </p>
                <Link 
                  href="/billing" 
                  className="inline-flex items-center gap-2 text-indigo-600 dark:text-indigo-400 hover:underline font-medium"
                >
                  View Plans <CreditCard className="h-4 w-4" />
                </Link>
              </div>
            </div>
          </div>

          {/* Step 2 */}
          <div className="border border-gray-200 dark:border-gray-800 rounded-lg p-6 bg-white dark:bg-slate-900">
            <div className="flex items-start gap-4">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-indigo-100 dark:bg-indigo-900 text-indigo-600 dark:text-indigo-400 font-semibold">
                2
              </div>
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
                  Upload Your Photo
                </h3>
                <p className="text-gray-600 dark:text-gray-400 mb-3">
                  Upload a clear photo of an empty room. Works best with good lighting and minimal clutter. Supports JPG, PNG, and HEIC formats.
                </p>
                <Link 
                  href="/upload" 
                  className="inline-flex items-center gap-2 text-indigo-600 dark:text-indigo-400 hover:underline font-medium"
                >
                  Upload Now <Upload className="h-4 w-4" />
                </Link>
              </div>
            </div>
          </div>

          {/* Step 3 */}
          <div className="border border-gray-200 dark:border-gray-800 rounded-lg p-6 bg-white dark:bg-slate-900">
            <div className="flex items-start gap-4">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-indigo-100 dark:bg-indigo-900 text-indigo-600 dark:text-indigo-400 font-semibold">
                3
              </div>
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
                  Choose Your Style
                </h3>
                <p className="text-gray-600 dark:text-gray-400 mb-3">
                  Select from Modern, Contemporary, Traditional, Industrial, or Scandinavian styles. Pick the room type (bedroom, living room, kitchen, etc.) for best results.
                </p>
              </div>
            </div>
          </div>

          {/* Step 4 */}
          <div className="border border-gray-200 dark:border-gray-800 rounded-lg p-6 bg-white dark:bg-slate-900">
            <div className="flex items-start gap-4">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-indigo-100 dark:bg-indigo-900 text-indigo-600 dark:text-indigo-400 font-semibold">
                4
              </div>
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
                  Download Your Staged Images
                </h3>
                <p className="text-gray-600 dark:text-gray-400 mb-3">
                  Processing takes 30-60 seconds. Once complete, download your professionally staged images ready for listings.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Best Practices */}
      <div className="mb-16">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">
          Best Practices for Great Results
        </h2>
        
        <div className="grid gap-6 sm:grid-cols-2">
          <div className="border border-gray-200 dark:border-gray-800 rounded-lg p-6 bg-white dark:bg-slate-900">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              ✓ Take Clear Photos
            </h3>
            <p className="text-gray-600 dark:text-gray-400 text-sm">
              Use good lighting and ensure the room is well-lit. Avoid shadows and dark corners.
            </p>
          </div>

          <div className="border border-gray-200 dark:border-gray-800 rounded-lg p-6 bg-white dark:bg-slate-900">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              ✓ Empty or Minimal Furniture
            </h3>
            <p className="text-gray-600 dark:text-gray-400 text-sm">
              Works best with completely empty rooms or rooms with minimal existing furniture.
            </p>
          </div>

          <div className="border border-gray-200 dark:border-gray-800 rounded-lg p-6 bg-white dark:bg-slate-900">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              ✓ Straight Angles
            </h3>
            <p className="text-gray-600 dark:text-gray-400 text-sm">
              Capture the room from a straight angle, not from corners or awkward positions.
            </p>
          </div>

          <div className="border border-gray-200 dark:border-gray-800 rounded-lg p-6 bg-white dark:bg-slate-900">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              ✓ Match Room Type
            </h3>
            <p className="text-gray-600 dark:text-gray-400 text-sm">
              Select the correct room type (bedroom, living room, etc.) for accurate furniture placement.
            </p>
          </div>
        </div>
      </div>

      {/* FAQ */}
      <div className="mb-16">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6 flex items-center gap-2">
          <HelpCircle className="h-6 w-6 text-indigo-600 dark:text-indigo-400" />
          Frequently Asked Questions
        </h2>
        
        <div className="space-y-4">
          <details className="group border border-gray-200 dark:border-gray-800 rounded-lg p-6 bg-white dark:bg-slate-900">
            <summary className="cursor-pointer text-lg font-semibold text-gray-900 dark:text-white list-none flex items-center justify-between">
              How long does staging take?
              <span className="text-gray-400 group-open:rotate-180 transition-transform">▼</span>
            </summary>
            <p className="mt-4 text-gray-600 dark:text-gray-400">
              Most images are processed in 30-60 seconds. You&apos;ll receive a notification when your staged image is ready.
            </p>
          </details>

          <details className="group border border-gray-200 dark:border-gray-800 rounded-lg p-6 bg-white dark:bg-slate-900">
            <summary className="cursor-pointer text-lg font-semibold text-gray-900 dark:text-white list-none flex items-center justify-between">
              What image formats are supported?
              <span className="text-gray-400 group-open:rotate-180 transition-transform">▼</span>
            </summary>
            <p className="mt-4 text-gray-600 dark:text-gray-400">
              We support JPG, PNG, and HEIC formats. Maximum file size is 10MB per image.
            </p>
          </details>

          <details className="group border border-gray-200 dark:border-gray-800 rounded-lg p-6 bg-white dark:bg-slate-900">
            <summary className="cursor-pointer text-lg font-semibold text-gray-900 dark:text-white list-none flex items-center justify-between">
              Can I stage multiple rooms at once?
              <span className="text-gray-400 group-open:rotate-180 transition-transform">▼</span>
            </summary>
            <p className="mt-4 text-gray-600 dark:text-gray-400">
              Yes! You can upload and process multiple images simultaneously. Each image counts toward your monthly limit.
            </p>
          </details>

          <details className="group border border-gray-200 dark:border-gray-800 rounded-lg p-6 bg-white dark:bg-slate-900">
            <summary className="cursor-pointer text-lg font-semibold text-gray-900 dark:text-white list-none flex items-center justify-between">
              What if I&apos;m not happy with the result?
              <span className="text-gray-400 group-open:rotate-180 transition-transform">▼</span>
            </summary>
            <p className="mt-4 text-gray-600 dark:text-gray-400">
              You can try different styles for the same room. Each style variation counts as one image toward your monthly limit.
            </p>
          </details>

          <details className="group border border-gray-200 dark:border-gray-800 rounded-lg p-6 bg-white dark:bg-slate-900">
            <summary className="cursor-pointer text-lg font-semibold text-gray-900 dark:text-white list-none flex items-center justify-between">
              Do my images remain private?
              <span className="text-gray-400 group-open:rotate-180 transition-transform">▼</span>
            </summary>
            <p className="mt-4 text-gray-600 dark:text-gray-400">
              Yes, your images are private and secure. Only you can access your uploaded and staged images.
            </p>
          </details>
        </div>
      </div>

      {/* Developer Resources */}
      <div className="border-t border-gray-200 dark:border-gray-800 pt-12">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6 flex items-center gap-2">
          <BookOpen className="h-6 w-6 text-indigo-600 dark:text-indigo-400" />
          Developer Resources
        </h2>
        
        <div className="grid gap-6 sm:grid-cols-2">
          <a
            href="https://api.real-staging.ai/api/v1/docs/"
            target="_blank"
            rel="noopener noreferrer"
            className="border border-gray-200 dark:border-gray-800 rounded-lg p-6 bg-white dark:bg-slate-900 hover:border-indigo-500 dark:hover:border-indigo-400 transition-colors"
          >
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              API Documentation →
            </h3>
            <p className="text-gray-600 dark:text-gray-400 text-sm">
              Complete REST API reference with authentication, endpoints, and examples.
            </p>
          </a>

          <a
            href="https://docs.real-staging.ai"
            target="_blank"
            rel="noopener noreferrer"
            className="border border-gray-200 dark:border-gray-800 rounded-lg p-6 bg-white dark:bg-slate-900 hover:border-indigo-500 dark:hover:border-indigo-400 transition-colors"
          >
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              Technical Documentation →
            </h3>
            <p className="text-gray-600 dark:text-gray-400 text-sm">
              Architecture guides, deployment instructions, and integration tutorials.
            </p>
          </a>
        </div>
      </div>
    </div>
  );
}
