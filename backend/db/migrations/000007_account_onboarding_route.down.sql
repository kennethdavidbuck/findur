UPDATE oauth_attempts
SET terminal_route = '/connect/result'
WHERE terminal_route = '/onboarding/accounts';

ALTER TABLE oauth_attempts
    DROP CONSTRAINT IF EXISTS oauth_attempts_terminal_route_check;

ALTER TABLE oauth_attempts
    ADD CONSTRAINT oauth_attempts_terminal_route_check
    CHECK (terminal_route IN ('/connect', '/connect/result', '/portfolio'));
