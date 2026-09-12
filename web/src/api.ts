export interface MeResponse {
  login: string;
  user_id: number;
}

/**
 * Returns the current authenticated user, or null when there is no valid
 * session. A non-200 response resolves to null (never throws).
 */
export async function me(): Promise<MeResponse | null> {
  try {
    const res = await fetch('/auth/me', {
      headers: { Accept: 'application/json' },
    });
    if (!res.ok) {
      return null;
    }
    return (await res.json()) as MeResponse;
  } catch {
    return null;
  }
}

/**
 * Signs the current user out by clearing the session cookie.
 * Returns true when the logout request succeeded.
 */
export async function logout(): Promise<boolean> {
  try {
    const res = await fetch('/auth/logout', { method: 'POST' });
    return res.ok;
  } catch {
    return false;
  }
}
