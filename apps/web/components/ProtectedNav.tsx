'use client';

import { useUser } from '@auth0/nextjs-auth0';
import Link from 'next/link';
import { Upload, ImageIcon, CreditCard, BookOpen, DollarSign } from 'lucide-react';

/**
 * ProtectedNav renders navigation links visible to all users (public + authenticated)
 */
export default function ProtectedNav() {
  const { user, isLoading } = useUser();

  return (
    <>
      <Link
        href="/docs"
        className="flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors hover:bg-gray-100 dark:hover:bg-slate-800"
      >
        <BookOpen className="h-4 w-4" />
        Docs
      </Link>
      <Link
        href="/pricing"
        className="flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors hover:bg-gray-100 dark:hover:bg-slate-800"
      >
        <DollarSign className="h-4 w-4" />
        Pricing
      </Link>
      {!isLoading && user && (
        <>
          <Link
            href="/upload"
            className="flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors hover:bg-gray-100 dark:hover:bg-slate-800"
          >
            <Upload className="h-4 w-4" />
            Upload
          </Link>
          <Link
            href="/images"
            className="flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors hover:bg-gray-100 dark:hover:bg-slate-800"
          >
            <ImageIcon className="h-4 w-4" />
            Images
          </Link>
          <Link
            href="/billing"
            className="flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors hover:bg-gray-100 dark:hover:bg-slate-800"
          >
            <CreditCard className="h-4 w-4" />
            Billing
          </Link>
        </>
      )}
    </>
  );
}
