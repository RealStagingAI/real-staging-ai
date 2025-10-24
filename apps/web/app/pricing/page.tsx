'use client';

import { Check } from 'lucide-react';
import Link from 'next/link';
import { useUser } from '@auth0/nextjs-auth0';

export default function PricingPage() {
  const { user } = useUser();

  return (
    <div className="mx-auto max-w-7xl">
      {/* Header */}
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-5xl mb-4">
          Simple, Transparent Pricing
        </h1>
        <p className="text-lg text-gray-600 dark:text-gray-400">
          Choose the plan that fits your needs. Upgrade or downgrade anytime.
        </p>
      </div>

      {/* Pricing Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-12">
        {/* Free Plan */}
        <div className="border border-gray-200 dark:border-gray-800 rounded-2xl p-8 bg-white dark:bg-slate-900">
          <div className="mb-8">
            <h3 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">Free</h3>
            <div className="flex items-baseline mb-4">
              <span className="text-5xl font-bold text-gray-900 dark:text-white">$0</span>
              <span className="text-gray-600 dark:text-gray-400 ml-2">/month</span>
            </div>
            <p className="text-gray-600 dark:text-gray-400">
              Perfect for trying out the platform
            </p>
          </div>

          <ul className="space-y-4 mb-8">
            <li className="flex items-start gap-3">
              <Check className="h-5 w-5 text-green-600 dark:text-green-400 flex-shrink-0 mt-0.5" />
              <span className="text-gray-700 dark:text-gray-300">10 images per month</span>
            </li>
            <li className="flex items-start gap-3">
              <Check className="h-5 w-5 text-green-600 dark:text-green-400 flex-shrink-0 mt-0.5" />
              <span className="text-gray-700 dark:text-gray-300">Standard processing speed</span>
            </li>
            <li className="flex items-start gap-3">
              <Check className="h-5 w-5 text-green-600 dark:text-green-400 flex-shrink-0 mt-0.5" />
              <span className="text-gray-700 dark:text-gray-300">All room types & styles</span>
            </li>
            <li className="flex items-start gap-3">
              <Check className="h-5 w-5 text-green-600 dark:text-green-400 flex-shrink-0 mt-0.5" />
              <span className="text-gray-700 dark:text-gray-300">Email support</span>
            </li>
          </ul>

          {user ? (
            <Link
              href="/billing"
              className="block w-full text-center px-6 py-3 border-2 border-gray-300 dark:border-gray-700 text-gray-900 dark:text-white rounded-lg hover:bg-gray-50 dark:hover:bg-slate-800 transition-colors font-medium"
            >
              Manage Plan
            </Link>
          ) : (
            <Link
              href="/auth/login"
              className="block w-full text-center px-6 py-3 border-2 border-gray-300 dark:border-gray-700 text-gray-900 dark:text-white rounded-lg hover:bg-gray-50 dark:hover:bg-slate-800 transition-colors font-medium"
            >
              Get Started Free
            </Link>
          )}
        </div>

        {/* Pro Plan */}
        <div className="border-2 border-indigo-500 dark:border-indigo-400 rounded-2xl p-8 bg-white dark:bg-slate-900 relative shadow-xl">
          <div className="absolute -top-4 left-1/2 -translate-x-1/2 px-4 py-1 bg-indigo-500 text-white text-sm font-semibold rounded-full">
            Most Popular
          </div>

          <div className="mb-8">
            <h3 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">Pro</h3>
            <div className="flex items-baseline mb-4">
              <span className="text-5xl font-bold text-gray-900 dark:text-white">$29</span>
              <span className="text-gray-600 dark:text-gray-400 ml-2">/month</span>
            </div>
            <p className="text-gray-600 dark:text-gray-400">
              For active real estate agents
            </p>
          </div>

          <ul className="space-y-4 mb-8">
            <li className="flex items-start gap-3">
              <Check className="h-5 w-5 text-indigo-600 dark:text-indigo-400 flex-shrink-0 mt-0.5" />
              <span className="text-gray-700 dark:text-gray-300"><strong>100 images</strong> per month</span>
            </li>
            <li className="flex items-start gap-3">
              <Check className="h-5 w-5 text-indigo-600 dark:text-indigo-400 flex-shrink-0 mt-0.5" />
              <span className="text-gray-700 dark:text-gray-300">Priority processing</span>
            </li>
            <li className="flex items-start gap-3">
              <Check className="h-5 w-5 text-indigo-600 dark:text-indigo-400 flex-shrink-0 mt-0.5" />
              <span className="text-gray-700 dark:text-gray-300">All room types & styles</span>
            </li>
            <li className="flex items-start gap-3">
              <Check className="h-5 w-5 text-indigo-600 dark:text-indigo-400 flex-shrink-0 mt-0.5" />
              <span className="text-gray-700 dark:text-gray-300">Batch upload support</span>
            </li>
            <li className="flex items-start gap-3">
              <Check className="h-5 w-5 text-indigo-600 dark:text-indigo-400 flex-shrink-0 mt-0.5" />
              <span className="text-gray-700 dark:text-gray-300">Chat support</span>
            </li>
          </ul>

          {user ? (
            <Link
              href="/billing"
              className="block w-full text-center px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors font-medium"
            >
              Upgrade to Pro
            </Link>
          ) : (
            <Link
              href="/auth/login?returnTo=/billing"
              className="block w-full text-center px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors font-medium"
            >
              Subscribe to Pro
            </Link>
          )}
        </div>

        {/* Business Plan */}
        <div className="border border-gray-200 dark:border-gray-800 rounded-2xl p-8 bg-white dark:bg-slate-900">
          <div className="mb-8">
            <h3 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">Business</h3>
            <div className="flex items-baseline mb-4">
              <span className="text-5xl font-bold text-gray-900 dark:text-white">$99</span>
              <span className="text-gray-600 dark:text-gray-400 ml-2">/month</span>
            </div>
            <p className="text-gray-600 dark:text-gray-400">
              For teams and high-volume users
            </p>
          </div>

          <ul className="space-y-4 mb-8">
            <li className="flex items-start gap-3">
              <Check className="h-5 w-5 text-purple-600 dark:text-purple-400 flex-shrink-0 mt-0.5" />
              <span className="text-gray-700 dark:text-gray-300"><strong>500 images</strong> per month</span>
            </li>
            <li className="flex items-start gap-3">
              <Check className="h-5 w-5 text-purple-600 dark:text-purple-400 flex-shrink-0 mt-0.5" />
              <span className="text-gray-700 dark:text-gray-300">Fastest processing</span>
            </li>
            <li className="flex items-start gap-3">
              <Check className="h-5 w-5 text-purple-600 dark:text-purple-400 flex-shrink-0 mt-0.5" />
              <span className="text-gray-700 dark:text-gray-300">All room types & styles</span>
            </li>
            <li className="flex items-start gap-3">
              <Check className="h-5 w-5 text-purple-600 dark:text-purple-400 flex-shrink-0 mt-0.5" />
              <span className="text-gray-700 dark:text-gray-300">Unlimited batch uploads</span>
            </li>
            <li className="flex items-start gap-3">
              <Check className="h-5 w-5 text-purple-600 dark:text-purple-400 flex-shrink-0 mt-0.5" />
              <span className="text-gray-700 dark:text-gray-300">Priority support</span>
            </li>
            <li className="flex items-start gap-3">
              <Check className="h-5 w-5 text-purple-600 dark:text-purple-400 flex-shrink-0 mt-0.5" />
              <span className="text-gray-700 dark:text-gray-300">API access</span>
            </li>
          </ul>

          {user ? (
            <Link
              href="/billing"
              className="block w-full text-center px-6 py-3 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors font-medium"
            >
              Upgrade to Business
            </Link>
          ) : (
            <Link
              href="/auth/login?returnTo=/billing"
              className="block w-full text-center px-6 py-3 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors font-medium"
            >
              Subscribe to Business
            </Link>
          )}
        </div>
      </div>

      {/* FAQ Section */}
      <div className="border-t border-gray-200 dark:border-gray-800 pt-12">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white text-center mb-8">
          Frequently Asked Questions
        </h2>
        
        <div className="grid md:grid-cols-2 gap-8 max-w-4xl mx-auto">
          <div>
            <h3 className="font-semibold text-gray-900 dark:text-white mb-2">
              Can I change plans anytime?
            </h3>
            <p className="text-gray-600 dark:text-gray-400 text-sm">
              Yes! You can upgrade or downgrade your plan at any time. Changes take effect immediately with prorated billing.
            </p>
          </div>

          <div>
            <h3 className="font-semibold text-gray-900 dark:text-white mb-2">
              What happens if I exceed my limit?
            </h3>
            <p className="text-gray-600 dark:text-gray-400 text-sm">
              Your uploads will be paused until the next billing cycle. You can upgrade to a higher plan anytime to continue processing.
            </p>
          </div>

          <div>
            <h3 className="font-semibold text-gray-900 dark:text-white mb-2">
              Do unused images roll over?
            </h3>
            <p className="text-gray-600 dark:text-gray-400 text-sm">
              No, image credits reset at the beginning of each monthly billing cycle.
            </p>
          </div>

          <div>
            <h3 className="font-semibold text-gray-900 dark:text-white mb-2">
              Can I cancel anytime?
            </h3>
            <p className="text-gray-600 dark:text-gray-400 text-sm">
              Yes, you can cancel your subscription at any time. You&apos;ll continue to have access until the end of your current billing period.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
