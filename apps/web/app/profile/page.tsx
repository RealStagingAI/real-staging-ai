'use client';

import { useUser } from '@auth0/nextjs-auth0';
import { useState, useEffect, useCallback, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { 
  User, 
  Mail, 
  Phone, 
  Building2, 
  CreditCard, 
  Settings,
  Save,
  Loader2,
  AlertCircle,
  CheckCircle,
  Bell,
  Palette,
  Home,
} from 'lucide-react';
import { apiFetch } from '@/lib/api';
import { toFormData, buildUpdatePayload } from '@/lib/profile';
import type { BackendProfile } from '@/lib/profile';
import { PaymentElementForm } from '@/components/stripe/PaymentElementForm';

interface SubscriptionAPI {
  id: string;
  status: string;
  price_id?: string;
  current_period_start?: string;
  current_period_end?: string;
  cancel_at?: string;
  canceled_at?: string;
  cancel_at_period_end: boolean;
}

interface Subscription {
  id: string;
  status: string;
  priceId?: string;
  currentPeriodStart?: string;
  currentPeriodEnd?: string;
  cancelAt?: string;
  canceledAt?: string;
  cancelAtPeriodEnd: boolean;
}

interface UsageStats {
  images_used: number;
  monthly_limit: number;
  plan_code: string;
  period_start: string;
  period_end: string;
  has_subscription: boolean;
  remaining_images: number;
}

function ProfilePageContent() {
  const { user, isLoading: authLoading } = useUser();
  const router = useRouter();
  const searchParams = useSearchParams();
  
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [usage, setUsage] = useState<UsageStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [pollingSubscription, setPollingSubscription] = useState(false);
  
  // Stripe Elements state
  const [showPaymentForm, setShowPaymentForm] = useState(false);
  const [clientSecret, setClientSecret] = useState('');
  const [selectedPlan, setSelectedPlan] = useState<string | null>(null);
  const [upgradeLoading, setUpgradeLoading] = useState<string | null>(null);
  const [manageLoading, setManageLoading] = useState(false);
  
  const [formData, setFormData] = useState({
    fullName: '',
    companyName: '',
    phone: '',
    addressLine1: '',
    addressLine2: '',
    city: '',
    state: '',
    postalCode: '',
    country: 'US',
    emailNotifications: true,
    marketingEmails: false,
    defaultRoomType: 'living_room',
    defaultStyle: 'modern',
  });

  // Define functions before they're used in useEffect
  const fetchProfileAndSubscription = useCallback(async () => {
    try {
      // Fetch profile, usage, and subscription in parallel
      const [profileData, usageData, subscriptionData] = await Promise.all([
        apiFetch<BackendProfile>('/v1/user/profile'),
        apiFetch<UsageStats>('/v1/billing/usage'),
        apiFetch<{ items: SubscriptionAPI[] }>('/v1/billing/subscriptions')
      ]);

      // Populate form from backend (snake_case) using mapper
      setFormData(toFormData(profileData));
      setUsage(usageData);
      
      // Set subscription if available and active, mapping snake_case to camelCase
      if (subscriptionData.items && subscriptionData.items.length > 0) {
        const activeSub = subscriptionData.items.find(
          (sub: SubscriptionAPI) => sub.status === 'active' || sub.status === 'trialing'
        );
        if (activeSub) {
          setSubscription({
            id: activeSub.id,
            status: activeSub.status,
            priceId: activeSub.price_id,
            currentPeriodStart: activeSub.current_period_start,
            currentPeriodEnd: activeSub.current_period_end,
            cancelAt: activeSub.cancel_at,
            canceledAt: activeSub.canceled_at,
            cancelAtPeriodEnd: activeSub.cancel_at_period_end
          });
        } else {
          setSubscription(null);
        }
      }
        
      // If we were polling and now have an active subscription, stop and show success
      if (pollingSubscription && subscriptionData.items && subscriptionData.items.length > 0) {
        const hasActiveSub = subscriptionData.items.some(
          (sub: SubscriptionAPI) => sub.status === 'active' || sub.status === 'trialing'
        );
        if (hasActiveSub) {
          setPollingSubscription(false);
          setMessage({ 
            type: 'success', 
            text: 'Subscription activated successfully! Welcome to your new plan.' 
          });
          setTimeout(() => setMessage(null), 5000);
          // Clean up URL
          router.replace('/profile');
        }
      }
    } catch (error) {
      console.error('Failed to fetch profile or subscription:', error);
    } finally {
      setLoading(false);
    }
  }, [pollingSubscription, router]);

  const pollSubscriptionWithRetry = useCallback(async () => {
    // Poll up to 10 times with 2 second intervals (20 seconds total)
    for (let i = 0; i < 10; i++) {
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      try {
        const data = await apiFetch<{ items: SubscriptionAPI[] }>('/v1/billing/subscriptions');
        if (data.items && data.items.length > 0) {
          const sub = data.items[0];
          setSubscription({
            id: sub.id,
            status: sub.status,
            priceId: sub.price_id,
            currentPeriodStart: sub.current_period_start,
            currentPeriodEnd: sub.current_period_end,
            cancelAt: sub.cancel_at,
            canceledAt: sub.canceled_at,
            cancelAtPeriodEnd: sub.cancel_at_period_end
          });
          setPollingSubscription(false);
          setMessage({ 
            type: 'success', 
            text: 'Subscription activated successfully! Welcome to your new plan.' 
          });
          setTimeout(() => setMessage(null), 5000);
          // Clean up URL
          router.replace('/profile');
          return; // Success, stop polling
        }
      } catch (error) {
        console.error('Poll attempt failed:', error);
      }
    }
    
    // If we get here, polling timed out
    setPollingSubscription(false);
    setMessage({ 
      type: 'error', 
      text: 'Subscription is taking longer than expected to activate. Please refresh the page in a moment.' 
    });
  }, [router]);

  // Redirect if not authenticated
  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/api/auth/login?returnTo=/profile');
    }
  }, [user, authLoading, router]);

  // Fetch user profile and handle checkout success
  useEffect(() => {
    if (user) {
      const checkoutStatus = searchParams.get('checkout');
      
      if (checkoutStatus === 'success') {
        // Show success message and poll for subscription
        setMessage({ 
          type: 'success', 
          text: 'Payment successful! Your subscription is being activated...' 
        });
        setPollingSubscription(true);
        
        // Poll for subscription (webhook may take a few seconds)
        pollSubscriptionWithRetry();
      } else if (checkoutStatus === 'canceled') {
        setMessage({ 
          type: 'error', 
          text: 'Checkout was canceled. You can try again anytime.' 
        });
        setTimeout(() => setMessage(null), 5000);
      }
      
      fetchProfileAndSubscription();
    }
  }, [user, searchParams, fetchProfileAndSubscription, pollSubscriptionWithRetry]);

  const handleSave = useCallback(async () => {
    setSaving(true);
    setMessage(null);

    try {
      const payload = buildUpdatePayload(formData);
      await apiFetch<BackendProfile>('/v1/user/profile', {
        method: 'PATCH',
        body: JSON.stringify(payload),
      });
      setMessage({ type: 'success', text: 'Profile updated successfully!' });
      fetchProfileAndSubscription(); // Refresh
    } catch (error) {
      setMessage({ type: 'error', text: 'Failed to update profile. Please try again.' });
    } finally {
      setSaving(false);
      setTimeout(() => setMessage(null), 5000);
    }
  }, [formData, fetchProfileAndSubscription]);

  const handleSubscribe = async (planCode: 'free' | 'pro' | 'business') => {
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
    
    try {
      setUpgradeLoading(planCode);
      
      // Create subscription with Elements (returns client secret)
      const data = await apiFetch<{ 
        subscriptionId: string;
        clientSecret: string;
      }>('/v1/billing/create-subscription-elements', {
        method: 'POST',
        body: JSON.stringify({ price_id: priceId }),
      });
      
      if (data?.clientSecret) {
        // Show payment form
        setClientSecret(data.clientSecret);
        setSelectedPlan(planCode);
        setShowPaymentForm(true);
      } else {
        console.error('No clientSecret in response:', data);
        setMessage({ type: 'error', text: 'Invalid response from server: missing client secret' });
      }
    } catch (error) {
      setMessage({ type: 'error', text: 'Failed to create subscription. Please try again.' });
    } finally {
      setUpgradeLoading(null);
    }
  };

  const handlePaymentSuccess = async () => {
    setMessage({ type: 'success', text: 'Payment successful! Your subscription is now active.' });
    setShowPaymentForm(false);
    setClientSecret('');
    setSelectedPlan(null);
    
    // Start polling for subscription activation
    setPollingSubscription(true);
    pollSubscriptionWithRetry();
  };

  const handlePaymentError = (error: Error) => {
    console.error('Payment error:', error);
    
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

  const handleManageSubscription = async () => {
    try {
      setManageLoading(true);
      const data = await apiFetch<{ url: string }>('/v1/billing/portal', {
        method: 'POST',
      });
      if (data?.url) {
        window.location.href = data.url; // Redirect to Stripe Customer Portal
      }
    } catch (error) {
      setMessage({ type: 'error', text: 'Failed to open billing portal. Please try again.' });
    } finally {
      setManageLoading(false);
    }
  };

  // Keyboard shortcut: Cmd+Enter (Mac) or Ctrl+Enter (Windows) to save
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
        e.preventDefault();
        if (!saving) {
          handleSave();
        }
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [saving, handleSave]);

  if (authLoading || loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] space-y-4">
        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
        <p className="text-gray-600 dark:text-gray-400">Loading your profile...</p>
      </div>
    );
  }

  if (!user) return null;

  return (
    <div className="max-w-5xl mx-auto space-y-4 sm:space-y-6">
      {/* Header */}
      <div className="space-y-2">
        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight">Profile Settings</h1>
        <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400">
          Manage your account information, billing, and preferences
        </p>
      </div>

      {/* Message Banner */}
      {message && (
        <div className={`rounded-lg border p-4 flex items-start gap-3 ${
          message.type === 'success' 
            ? 'bg-green-50 dark:bg-green-950/20 border-green-200 dark:border-green-800' 
            : 'bg-red-50 dark:bg-red-950/20 border-red-200 dark:border-red-800'
        }`}>
          {message.type === 'success' ? (
            <CheckCircle className="h-5 w-5 text-green-600 dark:text-green-400 mt-0.5" />
          ) : (
            <AlertCircle className="h-5 w-5 text-red-600 dark:text-red-400 mt-0.5" />
          )}
          <p className={message.type === 'success' ? 'text-green-800 dark:text-green-300' : 'text-red-800 dark:text-red-300'}>
            {message.text}
          </p>
        </div>
      )}

      {/* Personal Information */}
      <div className="card">
        <div className="card-header">
          <div className="flex items-center gap-2">
            <User className="h-5 w-5 text-blue-600" />
            <h2 className="text-xl font-semibold">Personal Information</h2>
          </div>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
            Your basic account details
          </p>
        </div>
        <div className="card-body space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-2">Full Name</label>
              <div className="relative">
                <User className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <input
                  type="text"
                  value={formData.fullName}
                  onChange={(e) => setFormData({ ...formData, fullName: e.target.value })}
                  className="w-full pl-10 pr-4 py-3 sm:py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900 text-base"
                  placeholder="John Doe"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">Email</label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <input
                  type="email"
                  value={user.email || ''}
                  disabled
                  className="w-full pl-10 pr-4 py-3 sm:py-2 border border-gray-300 dark:border-gray-700 rounded-lg bg-gray-50 dark:bg-slate-800 text-gray-500 cursor-not-allowed text-base"
                />
              </div>
              <p className="text-xs text-gray-500 mt-1">Email is managed by your Auth0 account</p>
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">Phone Number</label>
              <div className="relative">
                <Phone className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
                <input
                  type="tel"
                  value={formData.phone}
                  onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                  className="w-full pl-10 pr-4 py-3 sm:py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900 text-base"
                  placeholder="+1 (555) 123-4567"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Business Information */}
      <div className="card">
        <div className="card-header">
          <div className="flex items-center gap-2">
            <Building2 className="h-5 w-5 text-blue-600" />
            <h2 className="text-xl font-semibold">Business Information</h2>
          </div>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
            Company details and billing address
          </p>
        </div>
        <div className="card-body space-y-4">
          <div>
            <label className="block text-sm font-medium mb-2">Company Name</label>
            <div className="relative">
              <Building2 className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
              <input
                type="text"
                value={formData.companyName}
                onChange={(e) => setFormData({ ...formData, companyName: e.target.value })}
                className="w-full pl-10 pr-4 py-3 sm:py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900 text-base"
                placeholder="Acme Real Estate"
              />
            </div>
          </div>

          <div className="space-y-4">
            <label className="block text-sm font-medium">Billing Address</label>
            
            <div>
              <input
                type="text"
                value={formData.addressLine1}
                onChange={(e) => setFormData({ ...formData, addressLine1: e.target.value })}
                className="w-full px-4 py-3 sm:py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900 text-base"
                placeholder="Street address"
              />
            </div>

            <div>
              <input
                type="text"
                value={formData.addressLine2}
                onChange={(e) => setFormData({ ...formData, addressLine2: e.target.value })}
                className="w-full px-4 py-3 sm:py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900 text-base"
                placeholder="Apartment, suite, etc. (optional)"
              />
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <div className="col-span-2">
                <input
                  type="text"
                  value={formData.city}
                  onChange={(e) => setFormData({ ...formData, city: e.target.value })}
                  className="w-full px-4 py-3 sm:py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900 text-base"
                  placeholder="City"
                />
              </div>

              <div>
                <input
                  type="text"
                  value={formData.state}
                  onChange={(e) => setFormData({ ...formData, state: e.target.value })}
                  className="w-full px-4 py-3 sm:py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900 text-base"
                  placeholder="State"
                />
              </div>

              <div>
                <input
                  type="text"
                  value={formData.postalCode}
                  onChange={(e) => setFormData({ ...formData, postalCode: e.target.value })}
                  className="w-full px-4 py-3 sm:py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900 text-base"
                  placeholder="ZIP"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Payment & Billing */}
      <div className="card">
        <div className="card-header">
          <div className="flex items-center gap-2">
            <CreditCard className="h-5 w-5 text-blue-600" />
            <h2 className="text-xl font-semibold">Payment & Billing</h2>
          </div>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
          <button
            onClick={() => handleManageSubscription()}
            disabled={manageLoading}
            className="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-blue-400 disabled:cursor-not-allowed transition-colors font-medium flex items-center justify-center gap-2"
          >
            {manageLoading ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Opening...
              </>
            ) : (
              'Manage your subscription and payment methods'
            )}
          </button>
          </p>
        </div>
        <div className="card-body space-y-6">
          {usage && usage.plan_code !== '' && (
            <div className="space-y-4">
              {/* Current Plan Display */}
              <div className="flex items-center justify-between p-4 bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-950/20 dark:to-indigo-950/20 border border-blue-200 dark:border-blue-800 rounded-lg">
                <div>
                  <p className="text-sm font-medium text-blue-700 dark:text-blue-400">Current Plan</p>
                  <p className="text-2xl font-bold text-blue-900 dark:text-blue-300 mt-1">
                    {usage.plan_code === 'free' ? 'Free' : usage.plan_code === 'pro' ? 'Pro' : 'Business'} Plan
                  </p>
                  <p className="text-sm text-blue-600 dark:text-blue-500 mt-1">
                    {usage.monthly_limit} images per month • {usage.images_used} used
                  </p>
                  {subscription && subscription.currentPeriodStart && subscription.currentPeriodEnd && (
                    <p className="text-xs text-blue-600 dark:text-blue-500 mt-1">
                      Billing period: {new Date(subscription.currentPeriodStart).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })} - {new Date(subscription.currentPeriodEnd).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                    </p>
                  )}
                </div>

                {/* Manage Subscription Button (for paid plans) */}
                <div className="text-right">
                  {usage.plan_code === 'free' && (
                    <div className="text-3xl font-bold text-blue-900 dark:text-blue-300">$0<span className="text-lg text-blue-600 dark:text-blue-500">/mo</span></div>
                  )}
                  {usage.plan_code === 'pro' && (
                    <div className="text-3xl font-bold text-blue-900 dark:text-blue-300">$29<span className="text-lg text-blue-600 dark:text-blue-500">/mo</span></div>
                  )}
                  {usage.plan_code === 'business' && (
                    <div className="text-3xl font-bold text-blue-900 dark:text-blue-300">$99<span className="text-lg text-blue-600 dark:text-blue-500">/mo</span></div>
                  )}
                </div>
              </div>

              {/* Upgrade Options */}
              {usage.plan_code !== 'business' && (
                <div className="border-t border-gray-200 dark:border-gray-700 pt-6">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
                    {usage.plan_code === 'free' ? 'Upgrade Your Plan' : 'Available Upgrades'}
                  </h3>
                  
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    {/* Show Pro if on Free */}
                    {(usage.plan_code === 'free' || usage.plan_code === '') && (
                      <div className="border border-blue-200 dark:border-blue-700 rounded-lg p-5 hover:border-blue-400 dark:hover:border-blue-500 transition-colors">
                        <div className="flex items-start justify-between mb-3">
                          <div>
                            <h4 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Pro</h4>
                            <p className="text-2xl font-bold text-blue-600 mt-1">$29<span className="text-sm font-normal text-gray-600 dark:text-gray-400">/month</span></p>
                          </div>
                        </div>
                        <ul className="space-y-2 mb-4">
                          <li className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                            <CheckCircle className="h-4 w-4 text-blue-600" />
                            100 images per month
                          </li>
                          <li className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                            <CheckCircle className="h-4 w-4 text-blue-600" />
                            Priority processing
                          </li>
                          <li className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                            <CheckCircle className="h-4 w-4 text-blue-600" />
                            Chat support
                          </li>
                        </ul>
                        <button
                          onClick={() => handleSubscribe('pro')}
                          disabled={upgradeLoading === 'pro'}
                          className="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-blue-400 disabled:cursor-not-allowed transition-colors font-medium flex items-center justify-center gap-2"
                        >
                          {upgradeLoading === 'pro' ? (
                            <>
                              <Loader2 className="h-4 w-4 animate-spin" />
                              Processing...
                            </>
                          ) : (
                            'Upgrade to Pro'
                          )}
                        </button>
                      </div>
                    )}

                    {/* Show Business (available for both Free and Pro) */}
                    <div className="border-2 border-purple-300 dark:border-purple-600 rounded-lg p-5 relative hover:border-purple-400 dark:hover:border-purple-500 transition-colors">
                      <div className="absolute top-0 right-0 bg-purple-500 text-white text-xs font-semibold px-3 py-1 rounded-bl-lg rounded-tr-lg">
                        Best Value
                      </div>
                      <div className="flex items-start justify-between mb-3">
                        <div>
                          <h4 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Business</h4>
                          <p className="text-2xl font-bold text-purple-600 mt-1">$99<span className="text-sm font-normal text-gray-600 dark:text-gray-400">/month</span></p>
                        </div>
                      </div>
                      <ul className="space-y-2 mb-4">
                        <li className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                          <CheckCircle className="h-4 w-4 text-purple-600" />
                          500 images per month
                        </li>
                        <li className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                          <CheckCircle className="h-4 w-4 text-purple-600" />
                          Fastest processing
                        </li>
                        <li className="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                          <CheckCircle className="h-4 w-4 text-purple-600" />
                          Priority support
                        </li>
                      </ul>
                      <button
                        onClick={() => handleSubscribe('business')}
                        disabled={upgradeLoading === 'business'}
                        className="w-full px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 disabled:bg-purple-400 disabled:cursor-not-allowed transition-colors font-medium flex items-center justify-center gap-2"
                      >
                        {upgradeLoading === 'business' ? (
                          <>
                            <Loader2 className="h-4 w-4 animate-spin" />
                            Processing...
                          </>
                        ) : (
                          'Upgrade to Business'
                        )}
                      </button>
                    </div>
                  </div>
                </div>
              )}

              {/* Top Tier Message */}
              {usage.plan_code === 'business' && (
                <div className="p-4 bg-gradient-to-r from-purple-50 to-indigo-50 dark:from-purple-950/20 dark:to-indigo-950/20 border border-purple-200 dark:border-purple-800 rounded-lg text-center">
                  <p className="text-purple-900 dark:text-purple-300 font-medium">
                    🎉 You&apos;re on our top tier plan!
                  </p>
                  <p className="text-sm text-purple-700 dark:text-purple-400 mt-1">
                    Thank you for your business. You have access to all premium features.
                  </p>
                </div>
              )}
            </div>
          )}

          {/* No usage data available - show all plans */}
          {(!usage || usage.plan_code === '') && !subscription && (
            <div className="space-y-4">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
                Choose a Plan
              </h3>
              
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {/* Free Plan */}
                <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
                  <h4 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Free</h4>
                  <p className="text-xs text-green-600 dark:text-green-400 mt-1 flex items-center gap-1">
                    <svg className="h-3 w-3" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                    </svg>
                    No credit card required
                  </p>
                  <p className="text-3xl font-bold text-gray-900 dark:text-gray-100 mt-2">$0<span className="text-lg font-normal text-gray-600 dark:text-gray-400">/month</span></p>
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
                    onClick={() => handleSubscribe('free')}
                    disabled={upgradeLoading === 'free'}
                    className="w-full mt-6 px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
                  >
                    {upgradeLoading === 'free' ? (
                      <>
                        <Loader2 className="h-4 w-4 animate-spin" />
                        Processing...
                      </>
                    ) : (
                      usage?.plan_code === 'free' ? 'Continue with Free' : 'Subscribe to Free'
                    )}
                  </button>
                </div>

                {/* Pro Plan */}
                <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-6 hover:border-blue-500 dark:hover:border-blue-400 transition-colors">
                  <h4 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Pro</h4>
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
                    onClick={() => handleSubscribe('pro')}
                    disabled={upgradeLoading === 'pro'}
                    className="w-full mt-6 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-blue-400 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
                  >
                    {upgradeLoading === 'pro' ? (
                      <>
                        <Loader2 className="h-4 w-4 animate-spin" />
                        Processing...
                      </>
                    ) : (
                      'Subscribe to Pro'
                    )}
                  </button>
                </div>

                {/* Business Plan */}
                <div className="border-2 border-purple-500 dark:border-purple-400 rounded-lg p-6 relative">
                  <div className="absolute top-0 right-0 bg-purple-500 text-white text-xs font-semibold px-3 py-1 rounded-bl-lg rounded-tr-lg">
                    Best Value
                  </div>
                  <h4 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Business</h4>
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
                    onClick={() => handleSubscribe('business')}
                    disabled={upgradeLoading === 'business'}
                    className="w-full mt-6 px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 disabled:bg-purple-400 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
                  >
                    {upgradeLoading === 'business' ? (
                      <>
                        <Loader2 className="h-4 w-4 animate-spin" />
                        Processing...
                      </>
                    ) : (
                      'Subscribe to Business'
                    )}
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Preferences */}
      <div className="card">
        <div className="card-header">
          <div className="flex items-center gap-2">
            <Settings className="h-5 w-5 text-blue-600" />
            <h2 className="text-xl font-semibold">Preferences</h2>
          </div>
          <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
            Customize your experience
          </p>
        </div>
        <div className="card-body space-y-6">
          {/* Notifications */}
          <div className="space-y-4">
            <h3 className="font-medium flex items-center gap-2">
              <Bell className="h-4 w-4" />
              Notifications
            </h3>
            
            <label className="flex items-center justify-between p-3 border border-gray-200 dark:border-gray-800 rounded-lg cursor-pointer hover:bg-gray-50 dark:hover:bg-slate-800/50">
              <div>
                <p className="font-medium">Email Notifications</p>
                <p className="text-sm text-gray-500">Receive updates about your image processing</p>
              </div>
              <input
                type="checkbox"
                checked={formData.emailNotifications}
                onChange={(e) => setFormData({ ...formData, emailNotifications: e.target.checked })}
                className="h-5 w-5 text-blue-600 rounded focus:ring-2 focus:ring-blue-500"
              />
            </label>

            <label className="flex items-center justify-between p-3 border border-gray-200 dark:border-gray-800 rounded-lg cursor-pointer hover:bg-gray-50 dark:hover:bg-slate-800/50">
              <div>
                <p className="font-medium">Marketing Emails</p>
                <p className="text-sm text-gray-500">Receive news, tips, and special offers</p>
              </div>
              <input
                type="checkbox"
                checked={formData.marketingEmails}
                onChange={(e) => setFormData({ ...formData, marketingEmails: e.target.checked })}
                className="h-5 w-5 text-blue-600 rounded focus:ring-2 focus:ring-blue-500"
              />
            </label>
          </div>

          {/* Default Settings */}
          <div className="space-y-4">
            <h3 className="font-medium flex items-center gap-2">
              <Palette className="h-4 w-4" />
              Default Staging Settings
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-2 flex items-center gap-2">
                  <Home className="h-4 w-4" />
                  Default Room Type
                </label>
                <select
                  value={formData.defaultRoomType}
                  onChange={(e) => setFormData({ ...formData, defaultRoomType: e.target.value })}
                  className="w-full px-4 py-3 sm:py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900 text-base"
                >
                  <option value="living_room">Living Room</option>
                  <option value="bedroom">Bedroom</option>
                  <option value="kitchen">Kitchen</option>
                  <option value="bathroom">Bathroom</option>
                  <option value="dining_room">Dining Room</option>
                  <option value="office">Office</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium mb-2 flex items-center gap-2">
                  <Palette className="h-4 w-4" />
                  Default Style
                </label>
                <select
                  value={formData.defaultStyle}
                  onChange={(e) => setFormData({ ...formData, defaultStyle: e.target.value })}
                  className="w-full px-4 py-3 sm:py-2 border border-gray-300 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent bg-white dark:bg-slate-900 text-base"
                >
                  <option value="modern">Modern</option>
                  <option value="contemporary">Contemporary</option>
                  <option value="traditional">Traditional</option>
                  <option value="scandinavian">Scandinavian</option>
                  <option value="industrial">Industrial</option>
                  <option value="bohemian">Bohemian</option>
                </select>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Spacer for sticky button bar */}
      <div className="h-20 sm:h-24" />

      {/* Sticky Save Button Bar */}
      <div className="fixed bottom-0 left-0 right-0 z-40 border-t border-gray-200 dark:border-gray-800 bg-white/80 dark:bg-slate-950/80 backdrop-blur-xl supports-[backdrop-filter]:bg-white/60 dark:supports-[backdrop-filter]:bg-slate-950/60 pb-safe">
        <div className="container max-w-5xl mx-auto py-3 sm:py-4 flex justify-between items-center">
          {/* Keyboard shortcut hint */}
          <div className="hidden md:flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
            <kbd className="px-2 py-1 bg-gray-100 dark:bg-slate-800 border border-gray-300 dark:border-gray-700 rounded text-xs font-mono">
              {typeof navigator !== 'undefined' && navigator.platform.includes('Mac') ? '⌘' : 'Ctrl'}
            </kbd>
            <span>+</span>
            <kbd className="px-2 py-1 bg-gray-100 dark:bg-slate-800 border border-gray-300 dark:border-gray-700 rounded text-xs font-mono">
              Enter
            </kbd>
            <span>to save</span>
          </div>
          
          <div className="flex gap-2 sm:gap-3 ml-auto w-full sm:w-auto">
            <button
              onClick={() => router.push('/')}
              className="btn btn-secondary flex-1 sm:flex-none"
            >
              Cancel
            </button>
            <button
              onClick={handleSave}
              disabled={saving}
              className="btn btn-primary flex items-center justify-center gap-2 flex-1 sm:flex-none"
            >
              {saving ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Saving...
                </>
              ) : (
                <>
                  <Save className="h-4 w-4" />
                  Save Changes
                </>
              )}
            </button>
          </div>
        </div>
      </div>

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

// Wrap with Suspense boundary to handle useSearchParams
export default function ProfilePage() {
  return (
    <Suspense fallback={
      <div className="flex flex-col items-center justify-center min-h-[60vh] space-y-4">
        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
        <p className="text-gray-600 dark:text-gray-400">Loading your profile...</p>
      </div>
    }>
      <ProfilePageContent />
    </Suspense>
  );
}
