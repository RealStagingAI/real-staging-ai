'use client';

import { useState, useEffect } from 'react';
import { useUser } from '@auth0/nextjs-auth0';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Upload, ImageIcon, CreditCard, User, Home, LogOut, LogIn, BookOpen, DollarSign } from 'lucide-react';
import { cn } from '@/lib/utils';

/**
 * MobileNav provides a hamburger menu navigation for mobile devices
 * Includes smooth animations, touch-optimized targets, and safe area support
 */
export default function MobileNav() {
  const { user, isLoading } = useUser();
  const pathname = usePathname();
  const [isOpen, setIsOpen] = useState(false);

  // Close menu when route changes
  useEffect(() => {
    setIsOpen(false);
  }, [pathname]);

  // Prevent body scroll when menu is open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
      // Prevent touch move on body when menu is open to avoid swipe gestures
      document.body.style.touchAction = 'none';
    } else {
      document.body.style.overflow = '';
      document.body.style.touchAction = '';
    }
    return () => {
      document.body.style.overflow = '';
      document.body.style.touchAction = '';
    };
  }, [isOpen]);

  // Close menu on escape key
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setIsOpen(false);
    };
    if (isOpen) {
      window.addEventListener('keydown', handleEscape);
      return () => window.removeEventListener('keydown', handleEscape);
    }
  }, [isOpen]);

  // Prevent swipe gestures from revealing closed menu
  useEffect(() => {
    const handleTouchMove = (e: TouchEvent) => {
      if (!isOpen) {
        // Check if touch is starting from the right edge and moving left
        const touch = e.touches[0];
        if (touch && touch.clientX > window.innerWidth - 50) {
          e.preventDefault();
        }
      }
    };

    const handleTouchStart = (e: TouchEvent) => {
      if (!isOpen) {
        const touch = e.touches[0];
        if (touch && touch.clientX > window.innerWidth - 50) {
          e.preventDefault();
        }
      }
    };

    if (!isOpen) {
      document.addEventListener('touchmove', handleTouchMove, { passive: false });
      document.addEventListener('touchstart', handleTouchStart, { passive: false });
      return () => {
        document.removeEventListener('touchmove', handleTouchMove);
        document.removeEventListener('touchstart', handleTouchStart);
      };
    }
  }, [isOpen]);

  if (isLoading) return null;

  const navLinks = user
    ? [
        { href: '/', label: 'Home', icon: Home },
        { href: '/upload', label: 'Upload', icon: Upload },
        { href: '/images', label: 'Images', icon: ImageIcon },
        { href: '/docs', label: 'Docs', icon: BookOpen },
        { href: '/pricing', label: 'Pricing', icon: DollarSign },
        { href: '/billing', label: 'Billing', icon: CreditCard },
        { href: '/profile', label: 'Profile', icon: User },
      ]
    : [
        { href: '/', label: 'Home', icon: Home },
        { href: '/docs', label: 'Docs', icon: BookOpen },
        { href: '/pricing', label: 'Pricing', icon: DollarSign },
      ];

  return (
    <>
      {/* Hamburger Button - Touch Optimized (48x48 minimum) */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="md:hidden relative z-50 p-2 -mr-2 rounded-lg hover:bg-gray-100 dark:hover:bg-slate-800 active:scale-95 transition-all touch-manipulation"
        aria-label={isOpen ? 'Close menu' : 'Open menu'}
        aria-expanded={isOpen}
      >
        <div className="relative w-6 h-6">
          <span
            className={cn(
              'absolute left-0 top-2 w-6 h-0.5 bg-gray-900 dark:bg-gray-100 transition-all duration-300',
              isOpen && 'top-3 rotate-45'
            )}
          />
          <span
            className={cn(
              'absolute left-0 top-3 w-6 h-0.5 bg-gray-900 dark:bg-gray-100 transition-all duration-300',
              isOpen && 'opacity-0'
            )}
          />
          <span
            className={cn(
              'absolute left-0 top-4 w-6 h-0.5 bg-gray-900 dark:bg-gray-100 transition-all duration-300',
              isOpen && 'top-3 -rotate-45'
            )}
          />
        </div>
      </button>

      {/* Mobile Menu Overlay */}
      <div
        className={cn(
          'fixed inset-0 bg-black/50 backdrop-blur-sm z-40 md:hidden transition-opacity duration-300',
          isOpen ? 'opacity-100' : 'opacity-0 pointer-events-none'
        )}
        onClick={() => setIsOpen(false)}
        aria-hidden="true"
      />

      {/* Mobile Menu Panel - Slides in from right */}
      <nav
        className={cn(
          'fixed top-0 right-0 h-[100dvh] w-80 max-w-[85vw] bg-white dark:bg-slate-950 z-40 md:hidden',
          'shadow-2xl transform transition-transform duration-300 ease-out',
          'flex flex-col pt-safe', // Use flexbox for proper layout with top safe area
          isOpen ? 'translate-x-0' : 'translate-x-full pointer-events-none'
        )}
        aria-label="Mobile navigation"
        aria-hidden={!isOpen}
        {...({ inert: !isOpen } as React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement>)}
        style={{ 
          touchAction: isOpen ? 'pan-y' : 'none',
          willChange: 'transform'
        }}
      >
        {/* Menu Header */}
        <div className="flex-shrink-0 px-6 pt-6 pb-4 border-b border-gray-200 dark:border-gray-800">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-100">Menu</h2>
        </div>

        {/* Navigation Links - Scrollable area that takes remaining space */}
        <div className="flex-1 overflow-y-auto min-h-0">
          <div className="px-3 py-4 space-y-1">
            {navLinks.map(({ href, label, icon: Icon }) => {
              const isActive = pathname === href;
              return (
                <Link
                  key={href}
                  href={href}
                  className={cn(
                    'flex items-center gap-3 px-4 py-3 rounded-xl font-medium transition-all touch-manipulation',
                    'active:scale-95',
                    isActive
                      ? 'bg-blue-50 dark:bg-blue-950/30 text-blue-600 dark:text-blue-400'
                      : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-slate-800'
                  )}
                >
                  <Icon className="h-5 w-5 flex-shrink-0" />
                  <span className="text-base">{label}</span>
                  {isActive && (
                    <div className="ml-auto w-1.5 h-1.5 rounded-full bg-blue-600 dark:bg-blue-400" />
                  )}
                </Link>
              );
            })}
          </div>
        </div>

        {/* Auth Button - Always visible at bottom */}
        <div className="flex-shrink-0 px-4 pt-4 pb-safe pb-6 border-t border-gray-200 dark:border-gray-800 space-y-2 bg-white dark:bg-slate-950">
          {user ? (
            <a
              href="/auth/logout"
              className="flex items-center justify-center gap-2 w-full px-4 py-3 rounded-xl font-medium text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-950/30 transition-all active:scale-95 touch-manipulation"
            >
              <LogOut className="h-5 w-5" />
              <span>Sign Out</span>
            </a>
          ) : (
            <a
              href="/auth/login"
              className="flex items-center justify-center gap-2 w-full px-4 py-3 rounded-xl font-medium bg-gradient-to-r from-blue-600 to-indigo-600 text-white shadow-lg shadow-blue-500/30 hover:shadow-xl hover:shadow-blue-500/40 transition-all active:scale-95 touch-manipulation"
            >
              <LogIn className="h-5 w-5" />
              <span>Sign In</span>
            </a>
          )}
        </div>
      </nav>
    </>
  );
}
