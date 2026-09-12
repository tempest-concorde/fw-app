import {
  Button,
  LoginMainBody,
  LoginMainHeader,
  LoginPage,
} from '@patternfly/react-core';

export function LandingPage() {
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
