import { useCallback, useEffect, useState } from 'react';
import { me as fetchMe, logout as doLogout, MeResponse } from './api';

export type AuthState =
  | { status: 'loading' }
  | { status: 'unauthenticated' }
  | { status: 'authenticated'; user: MeResponse };

/**
 * Determines the current auth state from the backend session.
 * Authorization itself remains server-side; this hook only reflects it.
 */
export function useAuth(): AuthState {
  const [state, setState] = useState<AuthState>({ status: 'loading' });

  useEffect(() => {
    let active = true;
    fetchMe().then((user) => {
      if (!active) {
        return;
      }
      setState(user ? { status: 'authenticated', user } : { status: 'unauthenticated' });
    });
    return () => {
      active = false;
    };
  }, []);

  return state;
}

/**
 * Signs out and navigates back to the landing page.
 */
export function useSignOut(): () => Promise<void> {
  return useCallback(async () => {
    await doLogout();
    window.location.assign('/');
  }, []);
}
