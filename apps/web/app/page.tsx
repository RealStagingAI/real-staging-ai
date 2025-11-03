'use client';

import { useUser } from '@auth0/nextjs-auth0';
import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { Sparkles, Upload, ImageIcon, Zap, Shield, ArrowRight, AlertCircle, Home, TrendingUp, Users, CheckCircle, Move } from 'lucide-react';
import { apiFetch } from '@/lib/api';
import type { BackendProfile } from '@/lib/profile';
import BeforeAfterSlider from '@/components/ui/BeforeAfterSlider';
import { getMarketingImages } from '@/lib/marketingImages';

export default function Page() {
  const { user, isLoading, error } = useUser();
  const [profileName, setProfileName] = useState<string | null>(null);
  const [profileLoading, setProfileLoading] = useState(false);
  
  // Set profileLoading to true immediately when we have a user but no profile yet
  const shouldLoadProfile = user && !profileName && !profileLoading;
  if (shouldLoadProfile) {
    setProfileLoading(true);
  }

  const fallbackDisplayName = useMemo(() => {
    if (user?.name && !user.name.includes('@')) return user.name;
    if (user?.given_name) return user.family_name ? `${user.given_name} ${user.family_name}` : user.given_name;
    if (user?.nickname) return user.nickname;
    const local = user?.email?.split('@')[0];
    if (local) {
      const pretty = local
        .replace(/[._-]+/g, ' ')
        .split(' ')
        .filter(Boolean)
        .map(w => w.charAt(0).toUpperCase() + w.slice(1))
        .join(' ');
      return pretty || local;
    }
    return 'there';
  }, [user]);

  const displayName = profileName && profileName.trim().length > 0 ? profileName : fallbackDisplayName;

  useEffect(() => {
    if (!user) {
      setProfileLoading(false);
      return;
    }

    if (profileName) {
      return;
    }

    let cancelled = false;

    (async () => {
      try {
        const p = await apiFetch<BackendProfile>('/v1/user/profile');

        if (cancelled) return;

        if (p?.full_name && p.full_name.trim().length > 0) {
          setProfileName(p.full_name.trim());
          return;
        }

        if (p?.id) {
          setProfileName(fallbackDisplayName);
        }
      } catch {
        // ignore, fall back to Auth0 name heuristics
      } finally {
        if (!cancelled) {
          setProfileLoading(false);
        }
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [user, profileName, fallbackDisplayName]);

  if (isLoading || profileLoading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] space-y-4">
        <div className="relative">
          <div className="h-16 w-16 rounded-full border-4 border-gray-200 border-t-blue-600 animate-spin"></div>
          <Sparkles className="absolute inset-0 m-auto h-6 w-6 text-blue-600" />
        </div>
        <p className="text-gray-600 animate-pulse">Loading your workspace...</p>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="space-y-16 sm:space-y-24">
        {/* Hero Section with Before/After */}
        <section className="relative overflow-hidden">
          <div className="absolute inset-0 bg-gradient-to-br from-blue-50 via-white to-indigo-50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950 -z-10"></div>
          
          {/* Background Pattern */}
          <div className="absolute inset-0 opacity-5 dark:opacity-10">
            <div className="absolute inset-0" style={{
              backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23000000' fill-opacity='0.1'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`,
            }}></div>
          </div>
          
          <div className="relative">
            {/* Top Hero Content */}
            <div className="text-center space-y-6 sm:space-y-8 py-12 sm:py-16 lg:py-20">
              <div className="inline-flex items-center gap-2 rounded-full bg-gradient-to-r from-blue-100 to-indigo-100 dark:from-blue-900/30 dark:to-indigo-900/30 px-4 sm:px-6 py-2 sm:py-3 text-sm sm:text-base font-semibold text-blue-700 dark:text-blue-300 border border-blue-200/50 dark:border-blue-800/50 shadow-lg shadow-blue-500/10">
                <Sparkles className="h-4 w-4 sm:h-5 sm:w-5" />
                AI-Powered Virtual Staging for Real Estate
              </div>
              
              <div className="space-y-4 sm:space-y-6 px-4 sm:px-0">
                <h1 className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-bold tracking-tight leading-tight">
                  Transform Empty Properties into
                  <span className="gradient-text block mt-2 sm:mt-3">Showcase Homes That Sell</span>
                </h1>
                <p className="mx-auto max-w-3xl text-lg sm:text-xl text-gray-600 dark:text-gray-400 leading-relaxed px-4 sm:px-0">
                  Professional virtual staging powered by cutting-edge AI. Help buyers envision their future home 
                  and sell properties 73% faster with stunning, photorealistic staging in seconds.
                </p>
              </div>

              {error && error.message?.includes('failed to fetch') ? (
                <div className="card mx-auto max-w-2xl border-red-200 bg-red-50/50">
                  <div className="card-body space-y-4">
                    <div className="flex items-start gap-3">
                      <AlertCircle className="h-5 w-5 text-red-600 mt-0.5" />
                      <div className="flex-1 text-left">
                        <h2 className="text-lg font-semibold text-red-800">Auth0 Configuration Required</h2>
                        <p className="text-sm text-gray-700 mt-2">
                          Please configure Auth0 environment variables in <code className="bg-red-100 px-2 py-1 rounded text-xs">.env.local</code>
                        </p>
                        <div className="mt-4 space-y-2 text-sm text-gray-600">
                          <p className="font-medium">Required variables:</p>
                          <ul className="list-disc list-inside pl-4 space-y-1 text-xs">
                            <li><code className="bg-red-50 px-1 py-0.5 rounded">AUTH0_DOMAIN</code></li>
                            <li><code className="bg-red-50 px-1 py-0.5 rounded">AUTH0_CLIENT_ID</code></li>
                            <li><code className="bg-red-50 px-1 py-0.5 rounded">AUTH0_CLIENT_SECRET</code></li>
                            <li><code className="bg-red-50 px-1 py-0.5 rounded">AUTH0_SECRET</code> (generate with: <code className="text-xs">openssl rand -hex 32</code>)</li>
                            <li><code className="bg-red-50 px-1 py-0.5 rounded">APP_BASE_URL=http://localhost:3000</code></li>
                          </ul>
                          <p className="mt-3">
                            See <code className="bg-red-100 px-1 py-0.5 rounded text-xs">apps/web/env.example</code> for details.
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-center gap-4 sm:gap-6 pt-4 px-4 sm:px-0">
                  <a href="/auth/login" className="btn btn-primary group text-base sm:text-lg px-8 py-4 shadow-xl hover:shadow-2xl transform hover:scale-105 transition-all duration-200">
                    Start Staging Properties
                    <ArrowRight className="h-5 w-5 transition-transform group-hover:translate-x-1" />
                  </a>
                  <a href="#showcase" className="btn btn-secondary group text-base sm:text-lg px-8 py-4 border-2 hover:border-blue-300 transition-all duration-200">
                    See Results
                    <Move className="h-5 w-5 transition-transform group-hover:translate-y-1" />
                  </a>
                </div>
              )}
            </div>
            
            {/* Before/After Showcase */}
            <div id="showcase" className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-12 sm:py-16">
              <div className="text-center space-y-4 mb-12">
                <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 dark:text-white">
                  See the Difference in Seconds
                </h2>
                <p className="text-lg text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
                  Slide to compare empty rooms with beautifully staged spaces that captivate buyers
                </p>
              </div>
              
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 sm:gap-12">
                {getMarketingImages().filter(image => 
                  image.title === "Bedroom Transformation" || image.title === "Living Room Transformation"
                ).map((image, index) => (
                  <div key={index} className="space-y-4">
                    <h3 className="text-xl font-semibold text-center text-gray-800 dark:text-gray-200">
                      {image.title}
                    </h3>
                    <BeforeAfterSlider
                      beforeSrc={image.beforeSrc}
                      afterSrc={image.afterSrc}
                      beforeAlt={image.beforeAlt}
                      afterAlt={image.afterAlt}
                    />
                  </div>
                ))}
              </div>
            </div>
          </div>
        </section>

        {/* Stats Section */}
        {/* <section className="relative py-16 sm:py-20">
          <div className="absolute inset-0 bg-gradient-to-r from-blue-600 to-indigo-600 -z-10"></div>
          <div className="relative max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-8 text-center">
              <div className="space-y-2">
                <div className="text-3xl sm:text-4xl font-bold text-white">73%</div>
                <div className="text-blue-100 text-sm sm:text-base">Faster Sales</div>
              </div>
              <div className="space-y-2">
                <div className="text-3xl sm:text-4xl font-bold text-white">25%</div>
                <div className="text-blue-100 text-sm sm:text-base">Higher Offers</div>
              </div>
              <div className="space-y-2">
                <div className="text-3xl sm:text-4xl font-bold text-white">10sec</div>
                <div className="text-blue-100 text-sm sm:text-base">Processing Time</div>
              </div>
              <div className="space-y-2">
                <div className="text-3xl sm:text-4xl font-bold text-white">500+</div>
                <div className="text-blue-100 text-sm sm:text-base">Properties Staged</div>
              </div>
            </div>
          </div>
        </section> */}

        {/* Features Section */}
        <section id="features" className="space-y-12 sm:space-y-16">
          <div className="text-center space-y-4 sm:space-y-6 px-4 sm:px-0">
            <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 dark:text-white">
              Why Real Estate Professionals Choose Real Staging AI
            </h2>
            <p className="text-lg text-gray-600 dark:text-gray-400 max-w-3xl mx-auto leading-relaxed">
              Industry-leading technology designed specifically for real estate agents, property managers, 
              and home sellers who need professional results without the hassle or expense.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 sm:gap-8">
            <div className="card group cursor-default border-0 shadow-xl hover:shadow-2xl transform hover:-translate-y-1 transition-all duration-300">
              <div className="card-body space-y-6">
                <div className="rounded-2xl bg-gradient-to-br from-blue-500 to-indigo-500 p-4 w-fit shadow-lg shadow-blue-500/30 transition-all group-hover:shadow-xl group-hover:shadow-blue-500/40">
                  <Zap className="h-7 w-7 text-white" />
                </div>
                <div>
                  <h3 className="font-semibold text-xl mb-3 text-gray-900 dark:text-white">Lightning Fast Results</h3>
                  <p className="text-gray-600 dark:text-gray-400 leading-relaxed">
                    Transform empty spaces into beautifully furnished rooms in under 1 minute. 
                    No more waiting days for traditional staging services.
                  </p>
                </div>
              </div>
            </div>

            <div className="card group cursor-default border-0 shadow-xl hover:shadow-2xl transform hover:-translate-y-1 transition-all duration-300">
              <div className="card-body space-y-6">
                <div className="rounded-2xl bg-gradient-to-br from-emerald-500 to-green-500 p-4 w-fit shadow-lg shadow-emerald-500/30 transition-all group-hover:shadow-xl group-hover:shadow-emerald-500/40">
                  <TrendingUp className="h-7 w-7 text-white" />
                </div>
                <div>
                  <h3 className="font-semibold text-xl mb-3 text-gray-900 dark:text-white">Sell Properties Faster</h3>
                  <p className="text-gray-600 dark:text-gray-400 leading-relaxed">
                    Professionally staged homes sell 73% faster and for 25% more. 
                    Help buyers visualize the potential and close deals quicker.
                  </p>
                </div>
              </div>
            </div>

            <div className="card group cursor-default border-0 shadow-xl hover:shadow-2xl transform hover:-translate-y-1 transition-all duration-300">
              <div className="card-body space-y-6">
                <div className="rounded-2xl bg-gradient-to-br from-purple-500 to-pink-500 p-4 w-fit shadow-lg shadow-purple-500/30 transition-all group-hover:shadow-xl group-hover:shadow-purple-500/40">
                  <ImageIcon className="h-7 w-7 text-white" />
                </div>
                <div>
                  <h3 className="font-semibold text-xl mb-3 text-gray-900 dark:text-white">Multiple Design Styles</h3>
                  <p className="text-gray-600 dark:text-gray-400 leading-relaxed">
                    Modern, traditional, minimalist, contemporary — choose from 
                    professional interior design styles that match your target demographic.
                  </p>
                </div>
              </div>
            </div>

            <div className="card group cursor-default border-0 shadow-xl hover:shadow-2xl transform hover:-translate-y-1 transition-all duration-300">
              <div className="card-body space-y-6">
                <div className="rounded-2xl bg-gradient-to-br from-orange-500 to-red-500 p-4 w-fit shadow-lg shadow-orange-500/30 transition-all group-hover:shadow-xl group-hover:shadow-orange-500/40">
                  <Home className="h-7 w-7 text-white" />
                </div>
                <div>
                  <h3 className="font-semibold text-xl mb-3 text-gray-900 dark:text-white">Every Room Type</h3>
                  <p className="text-gray-600 dark:text-gray-400 leading-relaxed">
                    Living rooms, bedrooms, kitchens, dining rooms, bathrooms, 
                    and home offices — stage any space to maximize its appeal.
                  </p>
                </div>
              </div>
            </div>

            <div className="card group cursor-default border-0 shadow-xl hover:shadow-2xl transform hover:-translate-y-1 transition-all duration-300">
              <div className="card-body space-y-6">
                <div className="rounded-2xl bg-gradient-to-br from-cyan-500 to-blue-500 p-4 w-fit shadow-lg shadow-cyan-500/30 transition-all group-hover:shadow-xl group-hover:shadow-cyan-500/40">
                  <Shield className="h-7 w-7 text-white" />
                </div>
                <div>
                  <h3 className="font-semibold text-xl mb-3 text-gray-900 dark:text-white">Enterprise Security</h3>
                  <p className="text-gray-600 dark:text-gray-400 leading-relaxed">
                    Bank-level encryption and secure cloud storage ensure your 
                    property photos and client data remain completely confidential.
                  </p>
                </div>
              </div>
            </div>

            <div className="card group cursor-default border-0 shadow-xl hover:shadow-2xl transform hover:-translate-y-1 transition-all duration-300">
              <div className="card-body space-y-6">
                <div className="rounded-2xl bg-gradient-to-br from-violet-500 to-purple-500 p-4 w-fit shadow-lg shadow-violet-500/30 transition-all group-hover:shadow-xl group-hover:shadow-violet-500/40">
                  <Users className="h-7 w-7 text-white" />
                </div>
                <div>
                  <h3 className="font-semibold text-xl mb-3 text-gray-900 dark:text-white">Built for Professionals</h3>
                  <p className="text-gray-600 dark:text-gray-400 leading-relaxed">
                    Designed specifically for real estate agents with batch processing, 
                    project management, and tools to grow your business.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Testimonials Section */}
        {/* <section className="space-y-12 sm:space-y-16 py-16 sm:py-20 bg-gradient-to-br from-gray-50 to-blue-50 dark:from-slate-900 dark:to-blue-900/20 rounded-3xl">
          <div className="text-center space-y-4 sm:space-y-6 px-4 sm:px-0">
            <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 dark:text-white">
              Trusted by Real Estate Professionals
            </h2>
            <p className="text-lg text-gray-600 dark:text-gray-400 max-w-3xl mx-auto">
              Join thousands of agents who are selling properties faster with AI-powered staging
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 sm:gap-8 max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="card border-0 shadow-lg hover:shadow-xl transition-all duration-300">
              <div className="card-body space-y-4">
                <div className="flex items-center gap-1">
                  {[...Array(5)].map((_, i) => (
                    <Star key={i} className="h-4 w-4 fill-yellow-400 text-yellow-400" />
                  ))}
                </div>
                <p className="text-gray-700 dark:text-gray-300 italic leading-relaxed">
                  &quot;This tool has revolutionized my listing presentations. Properties that sat on the market for months are now selling in weeks.&quot;
                </p>
                <div className="flex items-center gap-3 pt-2">
                  <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-indigo-500 flex items-center justify-center text-white font-semibold">
                    SM
                  </div>
                  <div>
                    <div className="font-semibold text-gray-900 dark:text-white">Sarah Mitchell</div>
                    <div className="text-sm text-gray-600 dark:text-gray-400">Senior Real Estate Agent</div>
                  </div>
                </div>
              </div>
            </div>

            <div className="card border-0 shadow-lg hover:shadow-xl transition-all duration-300">
              <div className="card-body space-y-4">
                <div className="flex items-center gap-1">
                  {[...Array(5)].map((_, i) => (
                    <Star key={i} className="h-4 w-4 fill-yellow-400 text-yellow-400" />
                  ))}
                </div>
                <p className="text-gray-700 dark:text-gray-300 italic leading-relaxed">
                  &quot;The quality is incredible and the speed is unmatched. I can stage an entire property in the time it takes to grab coffee.&quot;
                </p>
                <div className="flex items-center gap-3 pt-2">
                  <div className="w-10 h-10 rounded-full bg-gradient-to-br from-emerald-500 to-green-500 flex items-center justify-center text-white font-semibold">
                    JC
                  </div>
                  <div>
                    <div className="font-semibold text-gray-900 dark:text-white">James Chen</div>
                    <div className="text-sm text-gray-600 dark:text-gray-400">Property Developer</div>
                  </div>
                </div>
              </div>
            </div>

            <div className="card border-0 shadow-lg hover:shadow-xl transition-all duration-300">
              <div className="card-body space-y-4">
                <div className="flex items-center gap-1">
                  {[...Array(5)].map((_, i) => (
                    <Star key={i} className="h-4 w-4 fill-yellow-400 text-yellow-400" />
                  ))}
                </div>
                <p className="text-gray-700 dark:text-gray-300 italic leading-relaxed">
                  &quot;My clients are amazed by the transformations. It helps them see the potential and makes decision-making so much easier.&quot;
                </p>
                <div className="flex items-center gap-3 pt-2">
                  <div className="w-10 h-10 rounded-full bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center text-white font-semibold">
                    ER
                  </div>
                  <div>
                    <div className="font-semibold text-gray-900 dark:text-white">Emily Rodriguez</div>
                    <div className="text-sm text-gray-600 dark:text-gray-400">Luxury Real Estate Broker</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section> */}

        {/* CTA Section */}
        <section className="relative py-16 sm:py-20">
          <div className="absolute inset-0 bg-gradient-to-r from-blue-600 to-indigo-600 rounded-3xl -z-10"></div>
          <div className="relative max-w-4xl mx-auto text-center px-4 sm:px-6 lg:px-8">
            <div className="space-y-6 sm:space-y-8">
              <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-white">
                Ready to Sell Properties Faster?
              </h2>
              <p className="text-lg sm:text-xl text-blue-100 leading-relaxed">
                Join thousands of real estate professionals using AI to stage properties 
                and close deals faster. Start your free trial today.
              </p>
              <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-center gap-4 sm:gap-6">
                <a href="/auth/login" className="bg-white text-blue-600 hover:bg-gray-50 px-8 py-4 rounded-xl font-semibold text-lg shadow-xl hover:shadow-2xl transform hover:scale-105 transition-all duration-200">
                  Start Free Trial
                  <ArrowRight className="inline-block ml-2 h-5 w-5" />
                </a>
                <div className="flex items-center justify-center gap-2 text-blue-100">
                  <CheckCircle className="h-5 w-5" />
                  <span className="text-sm sm:text-base">No credit card required</span>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>
    );
  }

  return (
    <div className="space-y-8 sm:space-y-12">
      {/* Welcome Hero */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-950/20 dark:to-indigo-950/20 opacity-50 rounded-3xl -z-10"></div>
        <div className="px-4 sm:px-8 py-8 sm:py-12 text-center space-y-3 sm:space-y-4">
          <div className="inline-flex items-center gap-2 rounded-full bg-gradient-to-r from-blue-600 to-indigo-600 px-3 sm:px-4 py-1.5 text-xs sm:text-sm font-medium text-white shadow-lg shadow-blue-500/30">
            <Sparkles className="h-3.5 w-3.5 sm:h-4 sm:w-4" />
            Dashboard
          </div>
          <h1 className="text-2xl sm:text-3xl md:text-4xl font-bold">
            Welcome back, <span className="gradient-text">{displayName}</span>!
          </h1>
          <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400 max-w-2xl mx-auto">
            Start staging properties or manage your existing projects. Everything you need is right at your fingertips.
          </p>
        </div>
      </section>

      {/* Quick Actions */}
      <section className="grid grid-cols-1 lg:grid-cols-2 gap-4 sm:gap-6">
        <Link href="/upload" className="card group">
          <div className="card-body space-y-4">
            <div className="flex items-start justify-between">
              <div className="rounded-xl bg-gradient-to-br from-blue-500 to-indigo-500 p-3 shadow-lg shadow-blue-500/30 transition-all group-hover:shadow-xl group-hover:shadow-blue-500/40">
                <Upload className="h-6 w-6 text-white" />
              </div>
              <ArrowRight className="h-5 w-5 text-gray-400 transition-all group-hover:translate-x-1 group-hover:text-blue-600" />
            </div>
            <div>
              <h3 className="font-semibold text-xl mb-2 group-hover:text-blue-600 transition-colors">Upload & Stage</h3>
              <p className="text-gray-600 text-sm">
                Upload new property photos and start the AI staging process. Choose room types and styles to match your vision.
              </p>
            </div>
          </div>
        </Link>

        <Link href="/images" className="card group">
          <div className="card-body space-y-4">
            <div className="flex items-start justify-between">
              <div className="rounded-xl bg-gradient-to-br from-purple-500 to-pink-500 p-3 shadow-lg shadow-purple-500/30 transition-all group-hover:shadow-xl group-hover:shadow-purple-500/40">
                <ImageIcon className="h-6 w-6 text-white" />
              </div>
              <ArrowRight className="h-5 w-5 text-gray-400 transition-all group-hover:translate-x-1 group-hover:text-purple-600" />
            </div>
            <div>
              <h3 className="font-semibold text-xl mb-2 group-hover:text-purple-600 transition-colors">View Images</h3>
              <p className="text-gray-600 text-sm">
                Browse your staged images by project. Monitor processing status and download results with live updates.
              </p>
            </div>
          </div>
        </Link>
      </section>
    </div>
  );
}
