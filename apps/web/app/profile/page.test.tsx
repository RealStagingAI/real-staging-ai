import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import React from 'react'
import { createRoot } from 'react-dom/client'
import { act } from 'react-dom/test-utils'

// Mock Auth0 useUser to simulate logged-in user
vi.mock('@auth0/nextjs-auth0', () => ({
  useUser: vi.fn(() => ({
    user: { email: 'user@example.com', name: 'Auth User' },
    error: null,
    isLoading: false,
  })),
}))

// Mock next/navigation router
vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
  useSearchParams: () => ({ get: vi.fn(() => null) }),
}))

// Mock api client
const apiFetchMock = vi.fn()
vi.mock('@/lib/api', () => ({
  apiFetch: (...args: unknown[]) => apiFetchMock(...args),
}))

import ProfilePage from './page'

function render(ui: React.ReactElement) {
  const container = document.createElement('div')
  document.body.appendChild(container)
  const root = createRoot(container)
  act(() => {
    root.render(ui)
  })
  return { container, root }
}

describe('ProfilePage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })
  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('success: fetches profile and subscriptions via apiFetch on mount', async () => {
    // Arrange responses:
    // 1) user profile
    apiFetchMock
      .mockResolvedValueOnce({
        id: 'id',
        role: 'user',
        created_at: '',
        updated_at: '',
        full_name: 'Test User',
        billing_address: { city: 'Metropolis' },
        preferences: { email_notifications: true },
      })
      // 2) billing usage
      .mockResolvedValueOnce({
        images_used: 5,
        monthly_limit: 100,
        plan_code: 'pro',
        period_start: '2025-10-20T00:00:00Z',
        period_end: '2025-11-20T00:00:00Z',
        has_subscription: true,
        remaining_images: 95
      })
      // 3) billing subscriptions
      .mockResolvedValueOnce({ 
        items: [{ 
          id: 'sub_1', 
          status: 'active',
          price_id: 'price_123',
          current_period_start: '2025-10-20T00:00:00Z',
          current_period_end: '2025-11-20T00:00:00Z',
          cancel_at_period_end: false
        }] 
      })

    const { container } = render(<ProfilePage />)

    // Wait for effects to run
    await act(async () => {
      await Promise.resolve()
    })

    // Assert calls
    expect(apiFetchMock.mock.calls.some((c) => c[0] === '/v1/user/profile')).toBe(true)
    expect(apiFetchMock.mock.calls.some((c) => c[0] === '/v1/billing/usage')).toBe(true)
    expect(apiFetchMock.mock.calls.some((c) => c[0] === '/v1/billing/subscriptions')).toBe(true)

    // Renders subscription UI (shows plan info)
    expect(container.textContent || '').toContain('Pro Plan')

    // Full name input should reflect mapped value
    const inputs = Array.from(container.querySelectorAll('input')) as HTMLInputElement[]
    const fullNameInput = inputs.find((el) => el.placeholder === 'John Doe')
    expect(fullNameInput?.value).toBe('Test User')
  })

  it('success: clicking Save sends PATCH with snake_case payload', async () => {
    // Arrange: profile and subscriptions initial
    apiFetchMock
      .mockResolvedValueOnce({
        id: 'id',
        role: 'user',
        created_at: '',
        updated_at: '',
        full_name: 'Snake Case',
      })
      .mockResolvedValueOnce({
        images_used: 0,
        monthly_limit: 10,
        plan_code: 'free',
        period_start: '2025-10-20T00:00:00Z',
        period_end: '2025-11-20T00:00:00Z',
        has_subscription: false,
        remaining_images: 10
      })
      .mockResolvedValueOnce({ items: [] })
      // PATCH update response (return profile)
      .mockResolvedValueOnce({
        id: 'id',
        role: 'user',
        created_at: '',
        updated_at: '',
        full_name: 'Snake Case',
      })

    const { container } = render(<ProfilePage />)

    await act(async () => {
      await Promise.resolve()
    })

    // Click Save button
    const saveBtn = Array.from(container.querySelectorAll('button')).find((b) => b.textContent?.includes('Save Changes')) as HTMLButtonElement
    expect(saveBtn).toBeTruthy()

    await act(async () => {
      saveBtn.click()
      await Promise.resolve()
    })

    // Expect third call to be PATCH '/v1/user/profile' with snake_case body
    const patchCall = apiFetchMock.mock.calls.find((c) => c[0] === '/v1/user/profile' && (c[1]?.method === 'PATCH'))
    expect(patchCall).toBeTruthy()

    const body = patchCall?.[1]?.body as string
    expect(typeof body).toBe('string')
    expect(body).toContain('"full_name"')
    expect(body).toContain('"billing_address"')

    // Success message shown
    expect(container.textContent || '').toContain('Profile updated successfully!')
  })

  it('success: clicking Subscribe triggers create-checkout call and redirects', async () => {
    // Mock environment variables for price IDs
    const originalEnv = process.env
    process.env = {
      ...originalEnv,
      NEXT_PUBLIC_STRIPE_PRICE_FREE: 'price_free_test',
      NEXT_PUBLIC_STRIPE_PRICE_PRO: 'price_pro_test', 
      NEXT_PUBLIC_STRIPE_PRICE_BUSINESS: 'price_business_test',
    }

    // Arrange: profile then no subscriptions -> shows Subscribe button
    apiFetchMock
      .mockResolvedValueOnce({ id: 'id', role: 'user', created_at: '', updated_at: '' })
      .mockResolvedValueOnce({
        images_used: 0,
        monthly_limit: 10,
        plan_code: 'free',
        period_start: '2025-10-20T00:00:00Z',
        period_end: '2025-11-20T00:00:00Z',
        has_subscription: false,
        remaining_images: 10
      })
      .mockResolvedValueOnce({ items: [] })
      .mockResolvedValueOnce({ 
        subscriptionId: 'sub_test_123',
        clientSecret: 'pi_test_123_secret'
      })

    // Mock window.location to allow setting href
    const originalLocation = window.location
    Object.defineProperty(window, 'location', { value: { href: '' } as Location, configurable: true })

    const { container } = render(<ProfilePage />)

    await act(async () => {
      await Promise.resolve()
    })

    // Click Subscribe/Upgrade button
    const subscribeBtn = Array.from(container.querySelectorAll('button')).find((b) => 
      b.textContent?.includes('Upgrade to') || b.textContent?.includes('Subscribe')
    ) as HTMLButtonElement
    expect(subscribeBtn).toBeTruthy()

    await act(async () => {
      subscribeBtn.click()
      await Promise.resolve()
    })

    // Expect POST to create-subscription-elements
    const call = apiFetchMock.mock.calls.find((c) => c[0] === '/v1/billing/create-subscription-elements')
    expect(call).toBeTruthy()
    expect(call?.[1]?.method).toBe('POST')

    // Restore location and env
    Object.defineProperty(window, 'location', { value: originalLocation, configurable: true })
    process.env = originalEnv
  })

  it('success: renders Pro plan with active subscription', async () => {
    // Arrange: profile with active Pro subscription
    apiFetchMock
      .mockResolvedValueOnce({ id: 'id', role: 'user', created_at: '', updated_at: '' })
      .mockResolvedValueOnce({
        images_used: 5,
        monthly_limit: 100,
        plan_code: 'pro',
        period_start: '2025-10-20T00:00:00Z',
        period_end: '2025-11-20T00:00:00Z',
        has_subscription: true,
        remaining_images: 95
      })
      .mockResolvedValueOnce({ 
        items: [{ 
          id: 'sub_1', 
          status: 'active',
          price_id: 'price_123',
          current_period_start: '2025-10-20T00:00:00Z',
          current_period_end: '2025-11-20T00:00:00Z',
          cancel_at_period_end: false
        }] 
      })

    const { container } = render(<ProfilePage />)

    // Wait for all async effects
    await act(async () => {
      await Promise.resolve()
      await Promise.resolve()
      await Promise.resolve()
    })

    // Verify Pro plan is displayed
    expect(container.textContent || '').toContain('Pro Plan')
    expect(container.textContent || '').toContain('100 images per month')
    expect(container.textContent || '').toContain('5 used')
    
    // Verify all API calls were made
    expect(apiFetchMock.mock.calls.some((c) => c[0] === '/v1/user/profile')).toBe(true)
    expect(apiFetchMock.mock.calls.some((c) => c[0] === '/v1/billing/usage')).toBe(true)
    expect(apiFetchMock.mock.calls.some((c) => c[0] === '/v1/billing/subscriptions')).toBe(true)
  })
})
