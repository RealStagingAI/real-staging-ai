'use client';

import { useEffect, useState } from 'react';
import { apiFetch } from '@/lib/api';
import { Clock, CreditCard, Loader2, Package, TrendingUp, AlertCircle, CheckCircle } from 'lucide-react';
import { PaymentElementForm } from '@/components/stripe/PaymentElementForm';

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
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  
  // Stripe Elements state
  const [showPaymentForm, setShowPaymentForm] = useState(false);
  const [clientSecret, setClientSecret] = useState('');
  const [selectedPlan, setSelectedPlan] = useState<string | null>(null);
  const [upgradeLoading, setUpgradeLoading] = useState<string | null>(null);
  const [manageLoading, setManageLoading] = useState(false);

  useEffect(() => {
    async function loadData() {
      try {
        setLoading(true);
        setMessage(null);

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
        setMessage({ type: 'error', text: err instanceof Error ? err.message : 'Failed to load billing information' });
      } finally {
        setLoading(false);
      }
    }

    loadData();
  }, []);

  const handleManageSubscription = async () => {
    try {
      setManageLoading(true);
      const response = await apiFetch<{ url: string }>('/v1/billing/portal', {
        method: 'POST',
      });
      window.location.href = response.url;
    } catch (err: unknown) {
      console.error('Failed to create portal session:', err);
      setMessage({ type: 'error', text: 'Failed to open billing portal' });
    } finally {
      setManageLoading(false);
    }
  };

  const handleUpgrade = async (planCode: string) => {
    try {
      setUpgradeLoading(planCode);
      
      // Get the price ID based on plan code
      const priceIds: Record<string, string | undefined> = {
        free: process.env.NEXT_PUBLIC_STRIPE_PRICE_FREE,
        pro: process.env.NEXT_PUBLIC_STRIPE_PRICE_PRO,
        business: process.env.NEXT_PUBLIC_STRIPE_PRICE_BUSINESS,
      };

      const priceId = priceIds[planCode];
      if (!priceId) {
        setMessage({ type: 'error', text: `Price ID not configured for ${planCode} plan` });
        return;
      }

      // Check if user has existing subscription
      if (subscription) {
        // Show confirmation prompt before upgrading
        const planNames: Record<string, string> = {
          free: 'Free',
          pro: 'Pro',
          business: 'Business'
        };
        
        const currentPlanName = usage ? planNames[usage.plan_code] || 'Unknown' : 'Current';
        const newPlanName = planNames[planCode] || 'Unknown';
        
        const confirmed = window.confirm(
          `Are you sure you want to upgrade from ${currentPlanName} to ${newPlanName}?\n\n` +
          `This will modify your existing subscription and may result in additional charges.\n\n` +
          `Click OK to continue or Cancel to keep your current plan.`
        );
        
        if (!confirmed) {
          return;
        }
        
        // Use upgrade endpoint for existing subscriptions
        const response = await apiFetch<{ 
          clientSecret?: string;
          success?: boolean;
          message?: string;
          subscriptionId?: string;
        }>('/v1/billing/upgrade-subscription', {
          method: 'POST',
          body: JSON.stringify({ price_id: priceId }),
        });

        // Handle immediate successful upgrade (no payment required)
        if (response.success) {
          setMessage({ type: 'success', text: response.message || 'Subscription updated successfully!' });
          
          // Reload billing data
          try {
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
            
            // Clear message after 3 seconds and return to normal view
            setTimeout(() => {
              setMessage(null);
            }, 3000);
          } catch (err: unknown) {
            console.error('Failed to reload billing data:', err);
          }
          return;
        }

        if (!response.clientSecret) {
          console.error('No clientSecret in upgrade response:', response);
          setMessage({ type: 'error', text: 'Invalid response from server: missing client secret' });
          return;
        }

        // Show payment form for upgrade
        setClientSecret(response.clientSecret);
        setSelectedPlan(planCode);
        setShowPaymentForm(true);
        return;
      }

      // Special handling for free plans - no payment required (new subscription)
      if (planCode === 'free') {
        // Create subscription directly without payment form
        const response = await apiFetch<{ 
          subscriptionId: string;
          clientSecret: string;
        }>('/v1/billing/create-subscription-elements', {
          method: 'POST',
          body: JSON.stringify({ price_id: priceId }),
        });

        if (response?.subscriptionId) {
          setMessage({ type: 'success', text: 'Free plan activated successfully!' });
          
          // Reload billing data and refresh page
          try {
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
            
            // Refresh the entire page to ensure UI consistency
            window.location.reload();
          } catch (err: unknown) {
            console.error('Failed to reload billing data:', err);
            // Still refresh page even if data reload fails
            window.location.reload();
          }
        } else {
          console.error('No subscriptionId in response:', response);
          setMessage({ type: 'error', text: 'Invalid response from server: missing subscription ID' });
        }
        return;
      }

      // For paid plans, show payment form (new subscription)
      const response = await apiFetch<{ 
        subscriptionId: string;
        clientSecret: string;
      }>('/v1/billing/create-subscription-elements', {
        method: 'POST',
        body: JSON.stringify({ price_id: priceId }),
      });

      if (!response.clientSecret) {
        console.error('No clientSecret in response:', response);
        setMessage({ type: 'error', text: 'Invalid response from server: missing client secret' });
        return;
      }

      // Show payment form
      setClientSecret(response.clientSecret);
      setSelectedPlan(planCode);
      setShowPaymentForm(true);
    } catch (err: unknown) {
      console.error('Failed to create subscription:', err);
      setMessage({ type: 'error', text: err instanceof Error ? err.message : 'Failed to create subscription' });
    } finally {
      setUpgradeLoading(null);
    }
  };

  const handlePaymentSuccess = async () => {
    setMessage({ type: 'success', text: 'Payment successful! Your subscription is now active.' });
    setShowPaymentForm(false);
    setClientSecret('');
    setSelectedPlan(null);
    
    // Reload billing data
    try {
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
      
      // Refresh the entire page to ensure UI consistency
      window.location.reload();
    } catch (err: unknown) {
      console.error('Failed to reload billing data:', err);
      // Still refresh page even if data reload fails
      window.location.reload();
    }

    // Clear message after 3 seconds and return to normal view
    setTimeout(() => {
      setMessage(null);
    }, 3000);
  };

  const handlePaymentError = (error: Error) => {
    // Check if it's an ad blocker issue
    if (error.message.includes('Failed to load') || error.message.includes('ERR_BLOCKED_BY_CLIENT')) {
      setMessage({ 
        type: 'error', 
        text: 'Payment services are blocked. Please disable ad blockers for this site and try again.' 
      });
    } else {
      setMessage({ type: 'error', text: error.message || 'Payment failed' });
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
    const used = Number(usage?.images_used ?? 0);
    const limit = Number(usage?.monthly_limit ?? 0);
    if (!isFinite(used) || !isFinite(limit) || limit <= 0) return 0;
    const pct = (used / limit) * 100;
    if (!isFinite(pct) || isNaN(pct)) return 0;
    return Math.max(0, Math.min(100, pct));
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

  if (message) {
    const isError = message.type === 'error';
    return (
      <div className="container max-w-7xl py-12">
        <div className={`${isError ? 'bg-red-50 dark:bg-red-950/20 border-red-200 dark:border-red-800' : 'bg-green-50 dark:bg-green-950/20 border-green-200 dark:border-green-800'} border rounded-lg p-6`}>
          <div className="flex items-center gap-3">
            {isError ? (
              <AlertCircle className="h-5 w-5 text-red-600" />
            ) : (
              <CheckCircle className="h-5 w-5 text-green-600" />
            )}
            <div>
              <h3 className={`font-semibold ${isError ? 'text-red-900 dark:text-red-300' : 'text-green-900 dark:text-green-300'}`}>
                {isError ? 'Error' : 'Success'}
              </h3>
              <p className={`text-sm mt-1 ${isError ? 'text-red-700 dark:text-red-400' : 'text-green-700 dark:text-green-400'}`}>
                {message.text}
              </p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="container max-w-7xl py-6 sm:py-8 lg:py-12 space-y-6 sm:space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-gray-100">
          Billing & Usage
        </h1>
        <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 mt-2">
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
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 sm:gap-6">
              <div className="space-y-2">
                <p className="text-xs sm:text-sm font-medium text-gray-600 dark:text-gray-400">Images Used</p>
                <p className={`text-2xl sm:text-3xl font-bold ${getUsageColor()}`}>
                  {usage.images_used.toLocaleString()}
                </p>
              </div>

              <div className="space-y-2">
                <p className="text-xs sm:text-sm font-medium text-gray-600 dark:text-gray-400">Monthly Limit</p>
                <p className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-gray-100">
                  {usage.monthly_limit.toLocaleString()}
                </p>
              </div>

              <div className="space-y-2">
                <p className="text-xs sm:text-sm font-medium text-gray-600 dark:text-gray-400">Remaining</p>
                <p className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-gray-100">
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
                  {usage?.plan_code ? `${usage.plan_code.toUpperCase()} Plan` : 'No Plan'}
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

            {subscription && usage?.plan_code == 'business' && (
              <button
                onClick={handleManageSubscription}
                disabled={manageLoading}
                className="flex items-center justify-center gap-2 px-4 py-2 bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900 rounded-lg hover:bg-gray-800 dark:hover:bg-gray-200 disabled:bg-gray-600 dark:disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors touch-manipulation w-full sm:w-auto"
              >
                {manageLoading ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" />
                    <span className="text-sm sm:text-base">Opening...</span>
                  </>
                ) : (
                  <>
                    <CreditCard className="h-4 w-4" />
                    <span className="text-sm sm:text-base">Manage Subscription</span>
                  </>
                )}
              </button>
            )}
          </div>

          {/* Subscription Details */}
          {subscription && (
            <div className="mt-4 sm:mt-6 pt-4 sm:pt-6 border-t border-gray-200 dark:border-gray-700">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
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
                <div className="mt-4 p-3 sm:p-4 bg-amber-50 dark:bg-amber-950/20 border border-amber-200 dark:border-amber-800 rounded-lg">
                  <p className="text-xs sm:text-sm text-amber-800 dark:text-amber-300">
                    <strong>Notice:</strong> Your subscription will be canceled at the end of the current billing period.
                  </p>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Upgrade Options (show for free and pro users, not business) */}
      {usage && usage.plan_code !== 'business' && (
        <div className="card">
          <div className="card-body">
            <h2 className="text-lg sm:text-xl font-semibold text-gray-900 dark:text-gray-100 mb-4 sm:mb-6">
              {(usage.plan_code === 'free' || usage.plan_code === '') ? 'Choose a Plan' : 'Upgrade Your Plan'}
            </h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
              {/* Free Plan */}
              {(usage.plan_code === 'free' || usage.plan_code === '') && !subscription && (
                <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 sm:p-6">
                  <h3 className="text-base sm:text-lg font-semibold text-gray-900 dark:text-gray-100">Free</h3>
                  <p className="text-xs text-green-600 dark:text-green-400 mt-1 flex items-center gap-1">
                    <svg className="h-3 w-3" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                    </svg>
                    No credit card required
                  </p>
                  <p className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-gray-100 mt-2">$0<span className="text-base sm:text-lg font-normal text-gray-600 dark:text-gray-400">/month</span></p>
                  <ul className="mt-4 space-y-2">
                    <li className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                      <div className="h-1.5 w-1.5 rounded-full bg-gray-600" />
                      100 images per month (Limited Time)
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
                    className="w-full mt-4 sm:mt-6 px-4 py-2.5 sm:py-2 bg-gray-600 text-white text-sm sm:text-base rounded-lg hover:bg-gray-700 transition-colors touch-manipulation"
                  >
                  {usage.plan_code === 'free' ? 'Continue with Free' : 'Subscribe to Free'}
                  </button>
                </div>
              )}
              {/* Pro Plan - only show to free users */}
              {(usage.plan_code === 'free' || usage.plan_code === '') && (
              <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 sm:p-6 hover:border-blue-500 dark:hover:border-blue-400 transition-colors">
                <h3 className="text-base sm:text-lg font-semibold text-gray-900 dark:text-gray-100">Pro</h3>
                <p className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-gray-100 mt-2">$29<span className="text-base sm:text-lg font-normal text-gray-600 dark:text-gray-400">/month</span></p>
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
                  disabled={upgradeLoading === 'pro'}
                  className="w-full mt-4 sm:mt-6 px-4 py-2.5 sm:py-2 bg-blue-600 text-white text-sm sm:text-base rounded-lg hover:bg-blue-700 disabled:bg-blue-400 disabled:cursor-not-allowed transition-colors touch-manipulation flex items-center justify-center gap-2"
                >
                  {upgradeLoading === 'pro' ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin" />
                      Processing...
                    </>
                  ) : (
                    subscription ? 'Upgrade to Pro' : 'Subscribe to Pro'
                  )}
                </button>
              </div>
              )}

              {/* Business Plan - show to free and pro users */}
              <div className="border-2 border-purple-500 dark:border-purple-400 rounded-lg p-4 sm:p-6 relative">
                <div className="absolute top-0 right-0 bg-purple-500 text-white text-xs font-semibold px-3 py-1 rounded-bl-lg rounded-tr-lg">
                  Best Value
                </div>
                <h3 className="text-base sm:text-lg font-semibold text-gray-900 dark:text-gray-100">Business</h3>
                <p className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-gray-100 mt-2">$99<span className="text-base sm:text-lg font-normal text-gray-600 dark:text-gray-400">/month</span></p>
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
                  disabled={upgradeLoading === 'business'}
                  className="w-full mt-4 sm:mt-6 px-4 py-2.5 sm:py-2 bg-purple-600 text-white text-sm sm:text-base rounded-lg hover:bg-purple-700 disabled:bg-purple-400 disabled:cursor-not-allowed transition-colors touch-manipulation flex items-center justify-center gap-2"
                >
                  {upgradeLoading === 'business' ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin" />
                      Processing...
                    </>
                  ) : (
                    subscription ? 'Upgrade to Business' : 'Subscribe to Business'
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Stripe Elements Payment Form Modal */}
      {showPaymentForm && clientSecret && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-md w-full">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
              Complete Your {selectedPlan ? selectedPlan.charAt(0).toUpperCase() + selectedPlan.slice(1) : ''} Subscription
            </h3>
            
            <PaymentElementForm
              clientSecret={clientSecret}
              onSuccess={handlePaymentSuccess}
              onError={handlePaymentError}
              buttonText={`Complete ${selectedPlan ? selectedPlan.charAt(0).toUpperCase() + selectedPlan.slice(1) : ''} Subscription`}
            />
            
            <div className="mt-4 flex justify-end">
              <button
                onClick={() => {
                  setShowPaymentForm(false);
                  setClientSecret('');
                  setSelectedPlan(null);
                }}
                className="px-4 py-2 text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
