DELETE FROM consents
WHERE user_id IN ('user-1', 'user-2') OR client_id = 'spa-client';

DELETE FROM sessions
WHERE user_id IN ('user-1', 'user-2');

DELETE FROM authorization_codes
WHERE user_id IN ('user-1', 'user-2') OR client_id = 'spa-client';

DELETE FROM refresh_tokens
WHERE user_id IN ('user-1', 'user-2') OR client_id = 'spa-client';

DELETE FROM clients
WHERE id = 'spa-client';

DELETE FROM users
WHERE id IN ('user-1', 'user-2');
