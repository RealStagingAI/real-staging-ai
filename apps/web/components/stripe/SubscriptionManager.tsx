'use client';

import { useState, useEffect } from 'react';
import { CreditCard, ArrowUpDown, Trash2, Plus } from 'lucide-react';
import { apiFetch } from '@/lib/api';

interface Subscription {
  id: string;
  status: string;
  priceId?: string;
  currentPeriodStart?: string;
  currentPeriodEnd?: string;
  cancelAtPeriodEnd: boolean;
}

interface PaymentMethod {
  id: string;
  type: string;
  card?: {
    brand: string;
    last4: string;
    expMonth: number;
    expYear: number;
  };
  isDefault: boolean;
}

interface SubscriptionManagerProps {
  subscription: Subscription | null;
  onSubscriptionChange: () => void;
}

export function SubscriptionManager({ 
  subscription, 
  onSubscriptionChange 
}: SubscriptionManagerProps) {
  const [paymentMethods, setPaymentMethods] = useState<PaymentMethod[]>([]);
  const [showAddPaymentMethod, setShowAddPaymentMethod] = useState(false);

  useEffect(() => {
    loadPaymentMethods();
  }, []);

  const loadPaymentMethods = async () => {
    try {
      const response = await apiFetch<{ paymentMethods: PaymentMethod[] }>('/v1/billing/payment-methods');
      setPaymentMethods(response.paymentMethods || []);
    } catch (error) {
      console.error('Failed to load payment methods:', error);
    }
  };

  const handleUpgradePlan = async (priceId: string) => {
    try {
      await apiFetch<{ clientSecret: string }>('/v1/billing/upgrade-subscription', {
        method: 'POST',
        body: JSON.stringify({ priceId }),
      });
      
      // Show payment form for upgrade
      setShowAddPaymentMethod(true);
      // TODO: Pass clientSecret to PaymentElementForm
    } catch (error) {
      console.error('Failed to upgrade subscription:', error);
    }
  };

  const handleCancelSubscription = async () => {
    if (!confirm('Are you sure you want to cancel your subscription?')) return;
    
    try {
      await apiFetch('/v1/billing/cancel-subscription', {
        method: 'POST',
      });
      onSubscriptionChange();
    } catch (error) {
      console.error('Failed to cancel subscription:', error);
    }
  };

  const handleSetDefaultPaymentMethod = async (paymentMethodId: string) => {
    try {
      await apiFetch('/v1/billing/set-default-payment-method', {
        method: 'POST',
        body: JSON.stringify({ paymentMethodId }),
      });
      loadPaymentMethods();
    } catch (error) {
      console.error('Failed to set default payment method:', error);
    }
  };

  const handleRemovePaymentMethod = async (paymentMethodId: string) => {
    try {
      await apiFetch('/v1/billing/remove-payment-method', {
        method: 'POST',
        body: JSON.stringify({ paymentMethodId }),
      });
      loadPaymentMethods();
    } catch (error) {
      console.error('Failed to remove payment method:', error);
    }
  };

  return (
    <div className="space-y-6">
      {/* Current Subscription */}
      {subscription && (
        <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
            Current Subscription
          </h3>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Status: <span className="font-medium capitalize">{subscription.status}</span>
              </p>
              {subscription.currentPeriodEnd && (
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  Renews: {new Date(subscription.currentPeriodEnd).toLocaleDateString()}
                </p>
              )}
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => handleUpgradePlan(process.env.NEXT_PUBLIC_STRIPE_PRICE_PRO!)} // Pro plan
                className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              >
                <ArrowUpDown className="h-4 w-4" />
                Upgrade
              </button>
              <button
                onClick={handleCancelSubscription}
                className="flex items-center gap-2 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
              >
                <Trash2 className="h-4 w-4" />
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Payment Methods */}
      <div className="border border-gray-200 dark:border-gray-700 rounded-lg p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
            Payment Methods
          </h3>
          <button
            onClick={() => setShowAddPaymentMethod(true)}
            className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
          >
            <Plus className="h-4 w-4" />
            Add Payment Method
          </button>
        </div>

        {paymentMethods.length === 0 ? (
          <p className="text-gray-600 dark:text-gray-400">No payment methods on file</p>
        ) : (
          <div className="space-y-3">
            {paymentMethods.map((method) => (
              <div
                key={method.id}
                className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg"
              >
                <div className="flex items-center gap-3">
                  <CreditCard className="h-5 w-5 text-gray-400" />
                  <div>
                    <p className="font-medium text-gray-900 dark:text-gray-100">
                      {method.card?.brand?.toUpperCase()} •••• {method.card?.last4}
                    </p>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      Expires {method.card?.expMonth}/{method.card?.expYear}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {method.isDefault && (
                    <span className="px-2 py-1 bg-green-100 text-green-800 text-xs font-medium rounded">
                      Default
                    </span>
                  )}
                  {!method.isDefault && (
                    <button
                      onClick={() => handleSetDefaultPaymentMethod(method.id)}
                      className="text-sm text-blue-600 hover:text-blue-700"
                    >
                      Set as default
                    </button>
                  )}
                  <button
                    onClick={() => handleRemovePaymentMethod(method.id)}
                    className="text-sm text-red-600 hover:text-red-700"
                  >
                    Remove
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Add Payment Method Modal */}
      {showAddPaymentMethod && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white dark:bg-gray-900 rounded-lg p-6 max-w-md w-full">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
              Add Payment Method
            </h3>
            {/* TODO: Integrate PaymentElementForm here */}
            <div className="flex gap-2">
              <button
                onClick={() => setShowAddPaymentMethod(false)}
                className="flex-1 px-4 py-2 bg-gray-200 text-gray-800 rounded-lg hover:bg-gray-300 transition-colors"
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
