import {
  Button,
  EmptyState,
  EmptyStateBody,
  EmptyStateFooter,
  Spinner,
} from '@patternfly/react-core';
import { useAuth, useSignOut } from '../auth';
import { LandingPage } from './LandingPage';

export function LoginSuccess() {
  const auth = useAuth();
  const signOut = useSignOut();

  if (auth.status === 'loading') {
    return (
      <EmptyState>
        <EmptyStateBody>
          <Spinner />
        </EmptyStateBody>
      </EmptyState>
    );
  }

  if (auth.status === 'unauthenticated') {
    // Session expired or invalid — show the landing page instead.
    return <LandingPage />;
  }

  return (
    <EmptyState
      status="success"
      titleText={`You are signed in as ${auth.user.login}`}
      headingLevel="h1"
    >
      <EmptyStateBody>You have access to the Flight Wall appliance.</EmptyStateBody>
      <EmptyStateFooter>
        <Button variant="secondary" onClick={signOut}>
          Sign out
        </Button>
      </EmptyStateFooter>
    </EmptyState>
  );
}
