'use client';

import { Check, Loader2, CreditCard } from 'lucide-react';
import Link from 'next/link';
import { useUser } from '@auth0/nextjs-auth0';
import { useState, useEffect } from 'react';
import { apiFetch } from '@/lib/api';

interface UsageStats {
  images_used: number;
  monthly_limit: number;
  plan_code: string;
  period_start: string;
  period_end: string;
  has_subscription: boolean;
  remaining_images: number;
}

interface Subscription {
  id: string;
  status: string;
  price_id?: string;
  current_period_start?: string;
  current_period_end?: string;
  cancel_at?: string;
  canceled_at?: string;
  cancel_at_period_end: boolean;
}

export default function PricingPage() {
  const { user, isLoading: authLoading } = useUser();
  const [usage, setUsage] = useState<UsageStats | null>(null);
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [loading, setLoading] = useState(false);

  // Load user's subscription data if logged in
  useEffect(() => {
    if (user && !authLoading) {
      loadBillingData();
    }
  }, [user, authLoading]);

  const loadBillingData = async () => {
    try {
      setLoading(true);
      const [usageData, subsData] = await Promise.all([
        apiFetch<UsageStats>('/v1/billing/usage'),
        apiFetch<{ items: Subscription[] }>('/v1/billing/subscriptions'),
      ]);
      setUsage(usageData);
      
      const activeSub = subsData.items?.find(
        (sub: Subscription) => sub.status === 'active' || sub.status === 'trialing'
      );
      setSubscription(activeSub || null);
    } catch (err) {
      console.error('Failed to load billing data:', err);
    } finally {
      setLoading(false);
    }
  };

  // If user is not logged in, show standard pricing page
  if (!user && !authLoading) {
    return <PublicPricingPage />;
  }

  // If loading auth or billing data
  if (authLoading || loading) {
    return (
      <div className="mx-auto max-w-7xl">
        <div className="flex items-center justify-center py-20">
          <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
          <span className="ml-2 text-gray-600 dark:text-gray-400">Loading your subscription...</span>
        </div>
      </div>
    );
  }

  // Show personalized pricing for logged-in users
  return (
    <div className="mx-auto max-w-7xl">
      {/* Header */}
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-5xl mb-4">
          Your Subscription
        </h1>
        <p className="text-lg text-gray-600 dark:text-gray-400">
          Manage your plan or upgrade to get more features
        </p>
      </div>

      {/* Current Plan Status */}
      {usage && subscription && (
        <div className="mb-8 p-6 bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-950/20 dark:to-indigo-950/20 border border-blue-200 dark:border-blue-800 rounded-lg">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-lg font-semibold text-blue-900 dark:text-blue-300">
                Current Plan: {usage.plan_code.charAt(0).toUpperCase() + usage.plan_code.slice(1)}
              </h3>
              <p className="text-blue-700 dark:text-blue-400 mt-1">
                {usage.images_used} / {usage.monthly_limit} images used this month
              </p>
              {subscription.current_period_start && subscription.current_period_end && (
                <p className="text-blue-600 dark:text-blue-500 text-sm mt-1">
                  Billing period: {new Date(subscription.current_period_start).toLocaleDateString('en-US', { 
                    month: 'short', 
                    day: 'numeric', 
                    year: 'numeric' 
                  })} - {new Date(subscription.current_period_end).toLocaleDateString('en-US', { 
                    month: 'short', 
                    day: 'numeric', 
                    year: 'numeric' 
                  })}
                </p>
              )}
            </div>
            <div className="text-right">
              <div className="text-2xl font-bold text-blue-900 dark:text-blue-300">
                ${usage.plan_code === 'free' ? '0' : usage.plan_code === 'pro' ? '29' : '99'}
                <span className="text-lg text-blue-600 dark:text-blue-500">/mo</span>
              </div>
              <Link
                href="/billing"
                className="inline-flex items-center gap-2 mt-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium"
              >
                <CreditCard className="h-4 w-4" />
                Manage Subscription
              </Link>
            </div>
          </div>
        </div>
      )}

      {/* Pricing Cards - Show available upgrades */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-12">
        {/* Free Plan */}
        <div className={`border ${
          usage?.plan_code === 'free' 
            ? 'border-blue-500 dark:border-blue-400 bg-blue-50 dark:bg-blue-950/20' 
            : 'border-gray-200 dark:border-gray-800 bg-white dark:bg-slate-900'
        } rounded-2xl p-8 relative`}>
          {usage?.plan_code === 'free' && (
            <div className="absolute -top-4 left-1/2 -translate-x-1/2 px-4 py-1 bg-blue-500 text-white text-sm font-semibold rounded-full">
              Current Plan
            </div>
          )}

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

          {usage?.plan_code === 'free' ? (
            <div className="block w-full text-center px-6 py-3 border-2 border-blue-300 dark:border-blue-700 text-blue-900 dark:text-blue-300 rounded-lg bg-blue-100 dark:bg-blue-900/30 font-medium">
              Your Current Plan
            </div>
          ) : (
            <Link
              href="/billing"
              className="block w-full text-center px-6 py-3 border-2 border-gray-300 dark:border-gray-700 text-gray-900 dark:text-white rounded-lg hover:bg-gray-50 dark:hover:bg-slate-800 transition-colors font-medium"
            >
              Downgrade to Free
            </Link>
          )}
        </div>

        {/* Pro Plan */}
        <div className={`border ${
          usage?.plan_code === 'pro' 
            ? 'border-indigo-500 dark:border-indigo-400 bg-indigo-50 dark:bg-indigo-950/20' 
            : 'border-gray-200 dark:border-gray-800 bg-white dark:bg-slate-900'
        } rounded-2xl p-8 relative`}>
          {usage?.plan_code === 'pro' && (
            <div className="absolute -top-4 left-1/2 -translate-x-1/2 px-4 py-1 bg-indigo-500 text-white text-sm font-semibold rounded-full">
              Current Plan
            </div>
          )}
          {usage?.plan_code !== 'pro' && (
            <div className="absolute -top-4 left-1/2 -translate-x-1/2 px-4 py-1 bg-indigo-500 text-white text-sm font-semibold rounded-full">
              Most Popular
            </div>
          )}

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

          {usage?.plan_code === 'pro' ? (
            <div className="block w-full text-center px-6 py-3 border-2 border-indigo-300 dark:border-indigo-700 text-indigo-900 dark:text-indigo-300 rounded-lg bg-indigo-100 dark:bg-indigo-900/30 font-medium">
              Your Current Plan
            </div>
          ) : usage?.plan_code === 'business' ? (
            <Link
              href="/billing"
              className="block w-full text-center px-6 py-3 border-2 border-gray-300 dark:border-gray-700 text-gray-900 dark:text-white rounded-lg hover:bg-gray-50 dark:hover:bg-slate-800 transition-colors font-medium"
            >
              Downgrade to Pro
            </Link>
          ) : (
            <Link
              href="/billing"
              className="block w-full text-center px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors font-medium"
            >
              Upgrade to Pro
            </Link>
          )}
        </div>

        {/* Business Plan */}
        <div className={`border ${
          usage?.plan_code === 'business' 
            ? 'border-purple-500 dark:border-purple-400 bg-purple-50 dark:bg-purple-950/20' 
            : 'border-gray-200 dark:border-gray-800 bg-white dark:bg-slate-900'
        } rounded-2xl p-8 relative`}>
          {usage?.plan_code === 'business' && (
            <div className="absolute -top-4 left-1/2 -translate-x-1/2 px-4 py-1 bg-purple-500 text-white text-sm font-semibold rounded-full">
              Current Plan
            </div>
          )}

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

          {usage?.plan_code === 'business' ? (
            <div className="block w-full text-center px-6 py-3 border-2 border-purple-300 dark:border-purple-700 text-purple-900 dark:text-purple-300 rounded-lg bg-purple-100 dark:bg-purple-900/30 font-medium">
              Your Current Plan
            </div>
          ) : (
            <Link
              href="/billing"
              className="block w-full text-center px-6 py-3 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors font-medium"
            >
              Upgrade to Business
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

// Separate component for public (logged-out) pricing page
function PublicPricingPage() {
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

          <Link
            href="/auth/login"
            className="block w-full text-center px-6 py-3 border-2 border-gray-300 dark:border-gray-700 text-gray-900 dark:text-white rounded-lg hover:bg-gray-50 dark:hover:bg-slate-800 transition-colors font-medium"
          >
            Get Started Free
          </Link>
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

          <Link
            href="/auth/login?returnTo=/billing"
            className="block w-full text-center px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors font-medium"
          >
            Subscribe to Pro
          </Link>
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

          <Link
            href="/auth/login?returnTo=/billing"
            className="block w-full text-center px-6 py-3 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors font-medium"
          >
            Subscribe to Business
          </Link>
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
