import { redirect } from 'next/navigation';
import { auth0 } from '@/lib/auth0';

/**
 * Analytics Dashboard Page
 * 
 * This is a client-side iframe that loads the Metabase dashboard
 * through the proxy route.
 */

export default async function AnalyticsPage() {
  const session = await auth0.getSession();
  
  if (!session || !session.user) {
    redirect('/api/auth/login');
  }

  // TODO: Add RBAC check
  // const userRole = session.user['https://real-staging.ai/roles'];
  // if (!userRole?.includes('admin')) {
  //   return (
  //     <div className="flex items-center justify-center h-screen">
  //       <div className="text-center">
  //         <h1 className="text-2xl font-bold mb-4">Access Denied</h1>
  //         <p className="text-gray-600">You need admin privileges to access analytics.</p>
  //       </div>
  //     </div>
  //   );
  // }

  return (
    <div className="flex flex-col h-screen">
      <div className="bg-white border-b border-gray-200 px-6 py-4">
        <h1 className="text-2xl font-bold text-gray-900">Analytics Dashboard</h1>
        <p className="text-sm text-gray-600 mt-1">
          Business intelligence and metrics powered by Metabase
        </p>
      </div>
      <div className="flex-1 relative">
        <iframe
          src="/admin/analytics/"
          className="absolute inset-0 w-full h-full border-0"
          title="Analytics Dashboard"
          sandbox="allow-same-origin allow-scripts allow-forms allow-popups allow-downloads"
        />
      </div>
    </div>
  );
}
