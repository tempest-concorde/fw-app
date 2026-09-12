import { useEffect } from 'react';
import {
  Button,
  EmptyState,
  EmptyStateBody,
  LoginMainBody,
  LoginMainHeader,
  LoginPage,
  Spinner,
} from '@patternfly/react-core';
import { useAuth } from '../auth';

export function LandingPage() {
  const auth = useAuth();

  useEffect(() => {
    // Already-authenticated users are sent straight to the success state
    // rather than being re-prompted to sign in.
    if (auth.status === 'authenticated') {
      window.location.assign('/app');
    }
  }, [auth.status]);

  // While checking an existing session (or redirecting an authenticated user),
  // avoid flashing the sign-in form.
  if (auth.status !== 'unauthenticated') {
    return (
      <EmptyState>
        <EmptyStateBody>
          <Spinner />
        </EmptyStateBody>
      </EmptyState>
    );
  }

  const signIn = () => {
    // Full-page navigation is required: the OAuth flow must be a top-level
    // browser redirect, not an XHR/fetch request.
    window.location.assign('/auth/login');
  };

  return (
    <LoginPage
      loginTitle="Flight Wall"
      loginSubtitle="Sign in to manage your Flight Wall display"
    >
      <LoginMainBody>
        <LoginMainHeader
          title="Sign in"
          subtitle="Use your GitHub account to continue."
        />
        <Button variant="primary" onClick={signIn}>
          Sign in with GitHub
        </Button>
      </LoginMainBody>
    </LoginPage>
  );
}
