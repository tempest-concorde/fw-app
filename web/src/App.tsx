import { LandingPage } from './pages/LandingPage';
import { LoginSuccess } from './pages/LoginSuccess';
import { LoginFailure } from './pages/LoginFailure';

/**
 * Minimal path-based router. All transitions are full-page navigations driven
 * by the server (OAuth redirects), so a client-side history API is unnecessary.
 */
export function App() {
  const path = window.location.pathname;

  if (path === '/app') {
    return <LoginSuccess />;
  }

  if (path.startsWith('/login')) {
    return <LoginFailure />;
  }

  return <LandingPage />;
}
