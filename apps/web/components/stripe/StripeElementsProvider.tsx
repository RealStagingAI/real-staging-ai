'use client';

import { loadStripe } from '@stripe/stripe-js';
import { Elements } from '@stripe/react-stripe-js';
import { ReactNode } from 'react';

const stripePromise = loadStripe(process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY!);

interface StripeElementsProviderProps {
  children: ReactNode;
  clientSecret?: string;
}

export function StripeElementsProvider({ children, clientSecret }: StripeElementsProviderProps) {
  const options = clientSecret ? { clientSecret } : {};
  
  return (
    <Elements stripe={stripePromise} options={options}>
      {children}
    </Elements>
  );
}
