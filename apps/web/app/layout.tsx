import './globals.css'
import Link from 'next/link'
import Image from 'next/image'
import type { Metadata, Viewport } from 'next'
import AuthButton from '@/components/AuthButton'
import UserProvider from '@/components/UserProvider'
import ProtectedNav from '@/components/ProtectedNav'
import MobileNav from '@/components/MobileNav'
import { ThemeProvider } from '@/components/ThemeProvider'
import { ThemeToggle } from '@/components/ThemeToggle'

export const metadata: Metadata = {
  title: 'Real Staging AI | Transform Properties with AI',
  description: 'Professional AI-powered virtual staging for real estate. Transform empty spaces into beautifully furnished rooms in seconds.',
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  maximumScale: 5,
  userScalable: true,
  themeColor: [
    { media: '(prefers-color-scheme: light)', color: '#ffffff' },
    { media: '(prefers-color-scheme: dark)', color: '#020617' },
  ],
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className="h-full" data-scroll-behavior="smooth" suppressHydrationWarning>
      <body className="h-full">
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          <UserProvider>
          <div className="flex min-h-screen flex-col">
            {/* Gradient Header */}
            <header className="sticky top-0 z-50 w-full border-b border-gray-200/60 dark:border-gray-800/60 bg-white/80 dark:bg-slate-950/80 backdrop-blur-xl supports-[backdrop-filter]:bg-white/60 dark:supports-[backdrop-filter]:bg-slate-950/60">
              <nav className="container flex h-14 sm:h-16 items-center justify-between">
                <div className="flex items-center gap-4 sm:gap-8 flex-1 min-w-0">
                  <Link href="/" className="flex items-center gap-2 sm:gap-3 group flex-shrink-0">
                    <Image
                      src="/logo.png"
                      alt="Real Staging AI Logo"
                      width={60}
                      height={60}
                      style={{ height: 'auto' }}
                      className="h-8 w-auto sm:h-10 transition-opacity group-hover:opacity-80"
                      priority
                    />
                  </Link>
                  <div className="hidden items-center gap-1 md:flex">
                    <ProtectedNav />
                  </div>
                </div>
                <div className="flex items-center gap-1 sm:gap-2 flex-shrink-0">
                  <ThemeToggle />
                  <div className="hidden md:block">
                    <AuthButton />
                  </div>
                  <MobileNav />
                </div>
              </nav>
            </header>

            {/* Main Content */}
            <main className="flex-1">
              <div className="container py-4 sm:py-6 lg:py-12 animate-in">
                {children}
              </div>
            </main>

            {/* Footer */}
            <footer className="border-t border-gray-200/60 dark:border-gray-800/60 bg-white/80 dark:bg-slate-950/80 backdrop-blur-sm pb-safe">
              <div className="container py-4 sm:py-6">
                <div className="flex flex-col items-center justify-between gap-2 sm:gap-4 sm:flex-row">
                  <p className="text-xs sm:text-sm text-gray-600 dark:text-gray-400 text-center sm:text-left">
                    © {new Date().getFullYear()} Real Staging AI. Built with Next.js & Replicate.
                  </p>
                  <div className="flex gap-4">
                    {/* Protected links removed from footer - only available when authenticated */}
                  </div>
                </div>
              </div>
            </footer>
          </div>
          </UserProvider>
        </ThemeProvider>
      </body>
    </html>
  )
}
