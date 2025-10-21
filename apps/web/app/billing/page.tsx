'use client';

import { useEffect, useState } from 'react';
import { apiFetch } from '@/lib/api';
import { Clock, CreditCard, Loader2, Package, TrendingUp, AlertCircle } from 'lucide-react';

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
  cancel_at?: string | null;
  canceled_at?: string | null;
  cancel_at_period_end: boolean;
}

interface SubscriptionResponse {
  items: Subscription[];
}

export default function BillingPage() {
  const [usage, setUsage] = useState<UsageStats | null>(null);
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadData() {
      try {
        setLoading(true);
        setError(null);

        const [usageData, subsData] = await Promise.all([
          apiFetch<UsageStats>('/v1/billing/usage'),
          apiFetch<SubscriptionResponse>('/v1/billing/subscriptions'),
        ]);

        setUsage(usageData);
        
        // Find active subscription
        const activeSub = subsData.items?.find(
          (sub: Subscription) => sub.status === 'active' || sub.status === 'trialing'
        );
        setSubscription(activeSub || null);
      } catch (err: unknown) {
        console.error('Failed to load billing data:', err);
        setError(err instanceof Error ? err.message : 'Failed to load billing information');
      } finally {
        setLoading(false);
      }
    }

    loadData();
  }, []);

  const handleManageSubscription = async () => {
    try {
      const response = await apiFetch<{ url: string }>('/v1/billing/portal', {
        method: 'POST',
      });
      window.location.href = response.url;
    } catch (err: unknown) {
      console.error('Failed to create portal session:', err);
      setError('Failed to open billing portal');
    }
  };

  const handleUpgrade = async (planCode: string) => {
    try {
      // Get the price ID based on plan code
      const priceIds: Record<string, string> = {
        free: process.env.NEXT_PUBLIC_STRIPE_PRICE_FREE || 'price_1SK67rLpUWppqPSl2XfvuIlh',
        pro: process.env.NEXT_PUBLIC_STRIPE_PRICE_PRO || 'price_1SJmy5LpUWppqPSlNElnvowM',
        business: process.env.NEXT_PUBLIC_STRIPE_PRICE_BUSINESS || 'price_1SJmyqLpUWppqPSlGhxfz2oQ',
      };

      const response = await apiFetch<{ url: string }>('/v1/billing/create-checkout', {
        method: 'POST',
        body: JSON.stringify({ price_id: priceIds[planCode] }),
      });

      window.location.href = response.url;
    } catch (err: unknown) {
      console.error('Failed to create checkout session:', err);
      setError('Failed to start upgrade process');
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  const getUsagePercentage = () => {
    if (!usage) return 0;
    return (usage.images_used / usage.monthly_limit) * 100;
  };

  const getUsageColor = () => {
    const percentage = getUsagePercentage();
    if (percentage >= 90) return 'text-red-600';
    if (percentage >= 70) return 'text-amber-600';
    return 'text-green-600';
  };

  if (loading) {
    return (
      <div className="container max-w-7xl py-12">
        <div className="flex items-center justify-center min-h-[400px]">
          <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container max-w-7xl py-12">
        <div className="bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-800 rounded-lg p-6">
          <div className="flex items-center gap-3">
            <AlertCircle className="h-5 w-5 text-red-600" />
            <div>
              <h3 className="font-semibold text-red-900 dark:text-red-300">Error</h3>
              <p className="text-sm text-red-700 dark:text-red-400 mt-1">{error}</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="container max-w-7xl py-12 space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">
          Billing & Usage
        </h1>
        <p className="text-gray-600 dark:text-gray-400 mt-2">
          Manage your subscription and track your monthly usage
        </p>
      </div>

      {/* Current Usage Card */}
      {usage && (
        <div className="card">
          <div className="card-body">
            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center gap-3">
                <div className="p-3 bg-blue-100 dark:bg-blue-900/20 rounded-lg">
                  <TrendingUp className="h-6 w-6 text-blue-600 dark:text-blue-400" />
                </div>
                <div>
                  <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
                    Current Usage
                  </h2>
                  <p className="text-sm text-gray-600 dark:text-gray-400">
                    Billing period: {formatDate(usage.period_start)} - {formatDate(usage.period_end)}
                  </p>
                </div>
              </div>
            </div>

            {/* Usage Stats */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="space-y-2">
                <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Images Used</p>
                <p className={`text-3xl font-bold ${getUsageColor()}`}>
                  {usage.images_used.toLocaleString()}
                </p>
              </div>

              <div className="space-y-2">
                <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Monthly Limit</p>
                <p className="text-3xl font-bold text-gray-900 dark:text-gray-100">
                  {usage.monthly_limit.toLocaleString()}
                </p>
              </div>

              <div className="space-y-2">
                <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Remaining</p>
                <p className="text-3xl font-bold text-gray-900 dark:text-gray-100">
                  {usage.remaining_images.toLocaleString()}
                </p>
              </div>
            </div>

            {/* Progress Bar */}
            <div className="mt-6">
              <div className="flex justify-between text-sm mb-2">
                <span className="text-gray-600 dark:text-gray-400">Usage Progress</span>
                <span className={`font-semibold ${getUsageColor()}`}>
                  {getUsagePercentage().toFixed(1)}%
                </span>
              </div>
              <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                <div
                  className={`h-full transition-all duration-500 ${
                    getUsagePercentage() >= 90
                      ? 'bg-red-500'
                      : getUsagePercentage() >= 70
                      ? 'bg-amber-500'
                      : 'bg-green-500'
                  }`}
                  style={{ width: `${Math.min(getUsagePercentage(), 100)}%` }}
                />
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Current Plan Card */}
      <div className="card">
        <div className="card-body">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-3 bg-purple-100 dark:bg-purple-900/20 rounded-lg">
                <Package className="h-6 w-6 text-purple-600 dark:text-purple-400" />
              </div>
              <div>
                <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100">
                  Current Plan
                </h2>
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                  {usage?.plan_code ? `${usage.plan_code.toUpperCase()} Plan` : 'Loading...'}
                  {subscription && ` • ${subscription.status === 'trialing' ? 'Trial' : 'Active'}`}
                </p>
                {subscription && subscription.current_period_start && subscription.current_period_end && (
                  <p className="text-xs text-gray-600 dark:text-gray-400 mt-1 flex items-center gap-1">
                    <svg className="h-3 w-3" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd" />
                    </svg>
                    Billing period: {formatDate(subscription.current_period_start)} - {formatDate(subscription.current_period_end)}
                  </p>
                )}
                {usage?.plan_code === 'free' && !subscription && (
                  <p className="text-xs text-green-600 dark:text-green-400 mt-1 flex items-center gap-1">
                    <svg className="h-3 w-3" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                    </svg>
                    No credit card required
                  </p>
                )}
              </div>
            </div>

            {subscription && (
              <button
                onClick={handleManageSubscription}
                className="flex items-center gap-2 px-4 py-2 bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900 rounded-lg hover:bg-gray-800 dark:hover:bg-gray-200 transition-colors"
              >
                <CreditCard className="h-4 w-4" />
                Manage Subscription
              </button>
            )}
          </div>

          {/* Subscription Details */}
          {subscription && (
            <div className="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {subscription.current_period_start && (
                  <div className="flex items-center gap-3">
                    <Clock className="h-5 w-5 text-gray-400" />
                    <div>
                      <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Period Start</p>
                      <p className="text-sm text-gray-900 dark:text-gray-100 font-medium">
                        {formatDate(subscription.current_period_start)}
                      </p>
                    </div>
                  </div>
                )}

                {subscription.current_period_end && (
                  <div className="flex items-center gap-3">
                    <Clock className="h-5 w-5 text-gray-400" />
                    <div>
                      <p className="text-sm font-medium text-gray-600 dark:text-gray-400">
                        {subscription.cancel_at_period_end ? 'Cancels On' : 'Renews On'}
                      </p>
                      <p className="text-sm text-gray-900 dark:text-gray-100 font-medium">
                        {formatDate(subscription.current_period_end)}
                      </p>
                    </div>
                  </div>
                )}
              </div>

              {subscription.cancel_at_period_end && (
                <div className="mt-4 p-4 bg-amber-50 dark:bg-amber-950/20 border border-amber-200 dark:border-amber-800 rounded-lg">
                  <p className="text-sm text-amber-800 dark:text-amber-300">
                    <strong>Notice:</strong> Your subscription will be canceled at the end of the current billing period.
                  </p>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Upgrade Options (if on free plan or no active subscription) */}
      {usage && (!subscription || usage.plan_code === 'free') && (
        <div className="card">
          <div className="card-body">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-6">
              {subscription ? 'Upgrade Your Plan' : 'Choose a Plan'}
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {/* Free Plan */}
              {!subscription && (
                <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Free</h3>
                  <p className="text-3xl font-bold text-gray-900 dark:text-gray-100 mt-2">$0<span className="text-lg font-normal text-gray-600 dark:text-gray-400">/month</span></p>
                  <ul className="mt-4 space-y-2">
                    <li className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                      <div className="h-1.5 w-1.5 rounded-full bg-gray-600" />
                      10 images per month
                    </li>
                    <li className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                      <div className="h-1.5 w-1.5 rounded-full bg-gray-600" />
                      Standard processing
                    </li>
                    <li className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                      <div className="h-1.5 w-1.5 rounded-full bg-gray-600" />
                      Email support
                    </li>
                  </ul>
                  <button
                    onClick={() => handleUpgrade('free')}
                    className="w-full mt-6 px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition-colors"
                  >
                    Continue with Free
                  </button>
                </div>
              )}
              {/* Pro Plan */}
              <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-6 hover:border-blue-500 dark:hover:border-blue-400 transition-colors">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Pro</h3>
                <p className="text-3xl font-bold text-gray-900 dark:text-gray-100 mt-2">$29<span className="text-lg font-normal text-gray-600 dark:text-gray-400">/month</span></p>
                <ul className="mt-4 space-y-2">
                  <li className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                    <div className="h-1.5 w-1.5 rounded-full bg-blue-600" />
                    100 images per month
                  </li>
                  <li className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                    <div className="h-1.5 w-1.5 rounded-full bg-blue-600" />
                    Priority processing
                  </li>
                  <li className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                    <div className="h-1.5 w-1.5 rounded-full bg-blue-600" />
                    Chat support
                  </li>
                </ul>
                <button
                  onClick={() => handleUpgrade('pro')}
                  className="w-full mt-6 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                >
                  {subscription ? 'Upgrade to Pro' : 'Subscribe to Pro'}
                </button>
              </div>

              {/* Business Plan */}
              <div className="border-2 border-purple-500 dark:border-purple-400 rounded-lg p-6 relative">
                <div className="absolute top-0 right-0 bg-purple-500 text-white text-xs font-semibold px-3 py-1 rounded-bl-lg rounded-tr-lg">
                  Best Value
                </div>
                <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Business</h3>
                <p className="text-3xl font-bold text-gray-900 dark:text-gray-100 mt-2">$99<span className="text-lg font-normal text-gray-600 dark:text-gray-400">/month</span></p>
                <ul className="mt-4 space-y-2">
                  <li className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                    <div className="h-1.5 w-1.5 rounded-full bg-purple-600" />
                    500 images per month
                  </li>
                  <li className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                    <div className="h-1.5 w-1.5 rounded-full bg-purple-600" />
                    Fastest processing
                  </li>
                  <li className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                    <div className="h-1.5 w-1.5 rounded-full bg-purple-600" />
                    Priority support
                  </li>
                </ul>
                <button
                  onClick={() => handleUpgrade('business')}
                  className="w-full mt-6 px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors"
                >
                  {subscription ? 'Upgrade to Business' : 'Subscribe to Business'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
