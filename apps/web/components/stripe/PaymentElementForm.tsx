'use client';

import { useState } from 'react';
import { PaymentElement, useStripe, useElements } from '@stripe/react-stripe-js';
import { Loader2, CreditCard } from 'lucide-react';
import { StripeElementsProvider } from './StripeElementsProvider';

interface PaymentElementFormProps {
  clientSecret: string;
  returnUrl?: string;
  onSuccess?: () => void;
  onError?: (error: Error) => void;
  buttonText?: string;
}

export function PaymentElementForm({
  clientSecret,
  returnUrl,
  onSuccess,
  onError,
  buttonText = 'Complete Payment'
}: PaymentElementFormProps) {
  // Build absolute return URL if relative path provided
  const absoluteReturnUrl = returnUrl 
    ? (returnUrl.startsWith('http') ? returnUrl : `${window.location.origin}${returnUrl}`)
    : `${window.location.origin}/profile?payment=success`;

  return (
    <StripeElementsProvider clientSecret={clientSecret}>
      <PaymentElementInner
        returnUrl={absoluteReturnUrl}
        onSuccess={onSuccess}
        onError={onError}
        buttonText={buttonText}
      />
    </StripeElementsProvider>
  );
}

interface PaymentElementInnerProps {
  returnUrl?: string;
  onSuccess?: () => void;
  onError?: (error: Error) => void;
  buttonText?: string;
}

function PaymentElementInner({
  returnUrl = '/profile?payment=success',
  onSuccess,
  onError,
  buttonText = 'Complete Payment'
}: PaymentElementInnerProps) {
  const stripe = useStripe();
  const elements = useElements();
  const [isLoading, setIsLoading] = useState(false);
  const [message, setMessage] = useState<string>('');

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    
    if (!stripe || !elements) {
      return;
    }

    setIsLoading(true);
    setMessage('');

    const { error } = await stripe.confirmPayment({
      elements,
      confirmParams: {
        return_url: returnUrl,
      },
      redirect: 'if_required', // Only redirect if necessary (e.g., 3D Secure)
    });

    if (error) {
      setMessage(error.message || 'An unexpected error occurred.');
      onError?.(new Error(error.message || 'An unexpected error occurred'));
    } else {
      onSuccess?.();
    }

    setIsLoading(false);
  };

  return (
    <form onSubmit={handleSubmit} className="w-full">
      <div className="mb-6">
        <PaymentElement 
          options={{
            layout: 'tabs'
          }}
        />
      </div>
      
      {message && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
          {message}
        </div>
      )}

      <button
        type="submit"
        disabled={isLoading || !stripe || !elements}
        className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        {isLoading ? (
          <>
            <Loader2 className="h-4 w-4 animate-spin" />
            Processing...
          </>
        ) : (
          <>
            <CreditCard className="h-4 w-4" />
            {buttonText}
          </>
        )}
      </button>
    </form>
  );
}
