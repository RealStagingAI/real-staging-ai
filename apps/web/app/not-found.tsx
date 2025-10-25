"use client";

import Link from "next/link";
import { Home, ArrowLeft, Search } from "lucide-react";
import { Button } from "@/components/ui/button";

export default function NotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center px-4 bg-gradient-to-b from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-950">
      <div className="max-w-2xl w-full text-center">
        {/* 404 Illustration */}
        <div className="mb-8">
          <div className="relative inline-block">
            <h1 className="text-9xl font-bold text-gray-200 dark:text-gray-800">404</h1>
            <div className="absolute inset-0 flex items-center justify-center">
              <Search className="h-24 w-24 text-gray-400 dark:text-gray-600" />
            </div>
          </div>
        </div>

        {/* Message */}
        <h2 className="text-3xl font-bold text-gray-900 dark:text-gray-100 mb-4">
          Page Not Found
        </h2>
        <p className="text-lg text-gray-600 dark:text-gray-400 mb-8">
          Oops! The page you&apos;re looking for doesn&apos;t exist or has been moved.
        </p>

        {/* Actions */}
        <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
          <Link href="/" className="min-w-[200px]">
            <Button size="lg" className="w-full">
              <Home className="mr-2 h-5 w-5" />
              Go Home
            </Button>
          </Link>
          <Button
            variant="outline"
            size="lg"
            className="min-w-[200px]"
            onClick={() => window.history.back()}
          >
            <ArrowLeft className="mr-2 h-5 w-5" />
            Go Back
          </Button>
        </div>

        {/* Helpful Links */}
        <div className="mt-12 pt-8 border-t border-gray-200 dark:border-gray-800">
          <p className="text-sm text-gray-500 dark:text-gray-500 mb-4">
            Looking for something specific?
          </p>
          <div className="flex flex-wrap gap-4 justify-center text-sm">
            <Link
              href="/pricing"
              className="text-blue-600 dark:text-blue-400 hover:underline"
            >
              Pricing
            </Link>
            <Link
              href="/docs"
              className="text-blue-600 dark:text-blue-400 hover:underline"
            >
              Documentation
            </Link>
            <Link
              href="/auth/login"
              className="text-blue-600 dark:text-blue-400 hover:underline"
            >
              Sign In
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
