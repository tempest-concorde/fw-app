import {
  Button,
  EmptyState,
  EmptyStateBody,
  EmptyStateFooter,
} from '@patternfly/react-core';

interface ErrorMessage {
  title: string;
  body: string;
}

function messageFor(code: string | null): ErrorMessage {
  switch (code) {
    case 'not_member':
      return {
        title: 'Access denied',
        body: 'Access is limited to members of the configured organization.',
      };
    case 'invalid_state':
      return {
        title: 'Sign-in could not be completed',
        body: 'The sign-in request was invalid. Please try again.',
      };
    case 'auth_failed':
      return {
        title: 'Sign-in could not be completed',
        body: 'GitHub authentication failed. Please try again.',
      };
    case 'config_error':
      return {
        title: 'Sign-in is unavailable',
        body: 'The appliance is not configured for sign-in. Please contact the operator.',
      };
    default:
      return {
        title: 'Sign-in could not be completed',
        body: 'Something went wrong during sign-in. Please try again.',
      };
  }
}

export function LoginFailure() {
  const code = new URLSearchParams(window.location.search).get('error');
  const message = messageFor(code);

  return (
    <EmptyState status="danger" titleText={message.title} headingLevel="h1">
      <EmptyStateBody>{message.body}</EmptyStateBody>
      <EmptyStateFooter>
        <Button
          variant="primary"
          onClick={() => window.location.assign('/auth/login')}
        >
          Try again
        </Button>
      </EmptyStateFooter>
    </EmptyState>
  );
}
