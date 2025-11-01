'use client';

import { useState, useEffect } from 'react';
import { useUser } from '@auth0/nextjs-auth0';
import { useRouter } from 'next/navigation';
import { apiFetch } from '@/lib/api';
import { StripeElementsProvider } from '@/components/stripe/StripeElementsProvider';
import { PaymentElementForm } from '@/components/stripe/PaymentElementForm';
import { SubscriptionManager } from '@/components/stripe/SubscriptionManager';
import { CheckCircle, Loader2, CreditCard, ArrowRight } from 'lucide-react';

interface Plan {
  id: string;
  name: string;
  price: number;
  currency: string;
  interval: string;
  features: string[];
  popular?: boolean;
  priceId: string;
}

const PLANS: Plan[] = [
  {
    id: 'free',
    name: 'Free',
    price: 0,
    currency: 'USD',
    interval: 'month',
    features: [
      '100 images per month',
      'Standard processing',
      'Email support',
    ],
    priceId: process.env.NEXT_PUBLIC_STRIPE_PRICE_FREE!,
  },
  {
    id: 'pro',
    name: 'Pro',
    price: 29,
    currency: 'USD',
    interval: 'month',
    features: [
      '100 images per month',
      'Priority processing',
      'Chat support',
    ],
    popular: true,
    priceId: process.env.NEXT_PUBLIC_STRIPE_PRICE_PRO!,
  },
  {
    id: 'business',
    name: 'Business',
    price: 99,
    currency: 'USD',
    interval: 'month',
    features: [
      '500 images per month',
      'Fastest processing',
      'Priority support',
    ],
    priceId: process.env.NEXT_PUBLIC_STRIPE_PRICE_BUSINESS!,
  },
];

interface Subscription {
  id: string;
  status: string;
  priceId?: string;
  currentPeriodStart?: string;
  currentPeriodEnd?: string;
  cancelAtPeriodEnd: boolean;
}

export default function SubscribePage() {
  const { user, isLoading: authLoading } = useUser();
  const router = useRouter();
  const [selectedPlan, setSelectedPlan] = useState<Plan | null>(null);
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [clientSecret, setClientSecret] = useState<string>('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string>('');
  const [success, setSuccess] = useState<string>('');

  useEffect(() => {
    if (user) {
      loadCurrentSubscription();
    }
  }, [user]);

  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/api/auth/login?returnTo=/subscribe');
    }
  }, [user, authLoading, router]);

  const loadCurrentSubscription = async () => {
    try {
      const response = await apiFetch<{ items: Subscription[] }>('/v1/billing/subscriptions');
      const activeSub = response.items?.find(
        (sub: Subscription) => sub.status === 'active' || sub.status === 'trialing'
      );
      setSubscription(activeSub || null);
    } catch (err) {
      console.error('Failed to load subscription:', err);
    }
  };

  const handleSelectPlan = async (plan: Plan) => {
    setSelectedPlan(plan);
    setError('');
    setSuccess('');
    
    if (plan.price === 0) {
      // Handle free plan directly
      await handleFreePlanSubscription(plan);
      return;
    }

    // Create subscription with Elements for paid plans
    setIsLoading(true);
    try {
      const response = await apiFetch<{
        subscriptionId: string;
        clientSecret: string;
      }>('/v1/billing/create-subscription-elements', {
        method: 'POST',
        body: JSON.stringify({ priceId: plan.priceId }),
      });
      
      setClientSecret(response.clientSecret);
    } catch (err: unknown) {
      const error = err as Error;
      setError(error.message || 'Failed to create subscription');
    } finally {
      setIsLoading(false);
    }
  };

  const handleFreePlanSubscription = async (plan: Plan) => {
    setIsLoading(true);
    try {
      await apiFetch<{
        subscriptionId: string;
        clientSecret: string;
      }>('/v1/billing/create-subscription-elements', {
        method: 'POST',
        body: JSON.stringify({ priceId: plan.priceId }),
      });
      
      // For free plans, we can confirm immediately
      setSuccess('Successfully subscribed to Free plan!');
      await loadCurrentSubscription();
      setSelectedPlan(null);
    } catch (err: unknown) {
      const error = err as Error;
      setError(error.message || 'Failed to subscribe to Free plan');
    } finally {
      setIsLoading(false);
    }
  };

  const handlePaymentSuccess = async () => {
    setSuccess('Payment successful! Your subscription is now active.');
    await loadCurrentSubscription();
    setSelectedPlan(null);
    setClientSecret('');
  };

  const handlePaymentError = (error: Error) => {
    setError(error.message || 'Payment failed');
  };

  const getCurrentPlan = () => {
    if (!subscription) return null;
    return PLANS.find(plan => plan.priceId === subscription.priceId);
  };

  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
      </div>
    );
  }

  if (!user) return null;

  const currentPlan = getCurrentPlan();

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <div className="container max-w-7xl mx-auto px-4 py-12">
        {/* Header */}
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold text-gray-900 dark:text-gray-100 mb-4">
            Choose Your Plan
          </h1>
          <p className="text-xl text-gray-600 dark:text-gray-400">
            Start with our free plan, upgrade when you&apos;re ready
          </p>
        </div>

        {/* Current Plan Alert */}
        {currentPlan && (
          <div className="mb-8 p-4 bg-green-50 dark:bg-green-950/20 border border-green-200 dark:border-green-800 rounded-lg">
            <div className="flex items-center gap-3">
              <CheckCircle className="h-5 w-5 text-green-600" />
              <div>
                <p className="text-gray-600 dark:text-gray-400">
                  You&apos;re currently on the {currentPlan.name} plan
                </p>
                <p className="text-sm text-green-700 dark:text-green-400">
                  {subscription?.currentPeriodEnd && (
                    <>Renews on {new Date(subscription.currentPeriodEnd).toLocaleDateString()}</>
                  )}
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Messages */}
        {error && (
          <div className="mb-8 p-4 bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-800 rounded-lg">
            <p className="text-red-900 dark:text-red-300">{error}</p>
          </div>
        )}
        
        {success && (
          <div className="mb-8 p-4 bg-green-50 dark:bg-green-950/20 border border-green-200 dark:border-green-800 rounded-lg">
            <p className="text-green-900 dark:text-green-300">{success}</p>
          </div>
        )}

        {/* Payment Form */}
        {selectedPlan && clientSecret && (
          <div className="mb-12">
            <div className="max-w-md mx-auto">
              <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
                <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">
                  Complete Your {selectedPlan.name} Subscription
                </h2>
                <p className="text-gray-600 dark:text-gray-400 mb-6">
                  ${selectedPlan.price}/{selectedPlan.interval}
                </p>
                
                <StripeElementsProvider>
                  <PaymentElementForm
                    clientSecret={clientSecret}
                    onSuccess={handlePaymentSuccess}
                    onError={handlePaymentError}
                    buttonText={`Subscribe to ${selectedPlan.name}`}
                  />
                </StripeElementsProvider>
                
                <button
                  onClick={() => {
                    setSelectedPlan(null);
                    setClientSecret('');
                    setError('');
                  }}
                  className="w-full mt-4 px-4 py-2 bg-gray-200 text-gray-800 rounded-lg hover:bg-gray-300 transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Plans Grid */}
        {!selectedPlan && (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-12">
            {PLANS.map((plan) => {
              const isCurrentPlan = currentPlan?.id === plan.id;
              const isUpgrade = currentPlan && 
                PLANS.indexOf(plan) > PLANS.indexOf(currentPlan);
              
              return (
                <div
                  key={plan.id}
                  className={`relative bg-white dark:bg-gray-800 rounded-lg shadow-lg p-8 ${
                    plan.popular ? 'ring-2 ring-blue-500' : ''
                  } ${isCurrentPlan ? 'ring-2 ring-green-500' : ''}`}
                >
                  {plan.popular && (
                    <div className="absolute top-0 right-0 bg-blue-500 text-white text-xs font-semibold px-3 py-1 rounded-bl-lg rounded-tr-lg">
                      Popular
                    </div>
                  )}
                  
                  {isCurrentPlan && (
                    <div className="absolute top-0 right-0 bg-green-500 text-white text-xs font-semibold px-3 py-1 rounded-bl-lg rounded-tr-lg">
                      Current Plan
                    </div>
                  )}

                  <h3 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">
                    {plan.name}
                  </h3>
                  <p className="text-4xl font-bold text-gray-900 dark:text-gray-100 mb-6">
                    ${plan.price}
                    <span className="text-lg font-normal text-gray-600 dark:text-gray-400">
                      /{plan.interval}
                    </span>
                  </p>

                  <ul className="space-y-3 mb-8">
                    {plan.features.map((feature, index) => (
                      <li key={index} className="flex items-center gap-3">
                        <CheckCircle className="h-5 w-5 text-green-600 flex-shrink-0" />
                        <span className="text-gray-600 dark:text-gray-400">{feature}</span>
                      </li>
                    ))}
                  </ul>

                  <button
                    onClick={() => handleSelectPlan(plan)}
                    disabled={isLoading || isCurrentPlan}
                    className={`w-full flex items-center justify-center gap-2 px-6 py-3 rounded-lg font-medium transition-colors ${
                      isCurrentPlan
                        ? 'bg-gray-200 text-gray-800 cursor-not-allowed'
                        : plan.popular
                        ? 'bg-blue-600 text-white hover:bg-blue-700'
                        : 'bg-gray-900 text-white hover:bg-gray-800'
                    } disabled:opacity-50`}
                  >
                    {isCurrentPlan ? (
                      'Current Plan'
                    ) : isUpgrade ? (
                      <>
                        Upgrade
                        <ArrowRight className="h-4 w-4" />
                      </>
                    ) : plan.price === 0 ? (
                      'Get Started'
                    ) : (
                      <>
                        <CreditCard className="h-4 w-4" />
                        Subscribe
                      </>
                    )}
                  </button>
                </div>
              );
            })}
          </div>
        )}

        {/* Subscription Management */}
        {subscription && (
          <div className="max-w-4xl mx-auto">
            <SubscriptionManager
              subscription={subscription}
              onSubscriptionChange={loadCurrentSubscription}
            />
          </div>
        )}
      </div>
    </div>
  );
}
