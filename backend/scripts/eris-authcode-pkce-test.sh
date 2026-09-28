#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
CLIENT_ID="${CLIENT_ID:-spa-client}"
CLIENT_SECRET="${CLIENT_SECRET:-}"
REDIRECT_URI="${REDIRECT_URI:-http://localhost:3000/callback}"
TOKEN_CLIENT_ARGS=(--data-urlencode "client_id=$CLIENT_ID")
if [ -n "$CLIENT_SECRET" ]; then
  TOKEN_CLIENT_ARGS+=(--data-urlencode "client_secret=$CLIENT_SECRET")
fi
TEST_TMP_DIR=$(mktemp -d)
COOKIE_JAR="$TEST_TMP_DIR/cookies.txt"
LOGIN_HEADERS="$TEST_TMP_DIR/login-headers.txt"
LOGIN_BODY="$TEST_TMP_DIR/login-body.txt"
trap 'rm -rf "$TEST_TMP_DIR"' EXIT

pass() { echo "  ✅ $1"; }
fail() { echo "  ❌ $1"; exit 1; }

echo "== 1. Generating PKCE pair =="
CODE_VERIFIER=$(openssl rand -base64 32 | tr -d '=+/' | cut -c1-43)
CODE_CHALLENGE=$(printf '%s' "$CODE_VERIFIER" | openssl dgst -sha256 -binary | openssl base64 | tr -d '=' | tr '+/' '-_')
echo "  verifier:  $CODE_VERIFIER"
echo "  challenge: $CODE_CHALLENGE"

echo
echo "== 2. GET /authorize (no session) — expect login form =="
RESP=$(curl -sS -c "$COOKIE_JAR" \
  "$BASE_URL/authorize?client_id=$CLIENT_ID&redirect_uri=$REDIRECT_URI&scope=openid+profile&state=xyz123&code_challenge=$CODE_CHALLENGE&code_challenge_method=S256")
echo "$RESP" | grep -q "action=\"/authorize/login\"" \
  && pass "login form returned" \
  || fail "expected login form, got: $RESP"

echo
echo "== 3. POST /authorize/login — expect consent form + session cookie =="
LOGIN_STATUS=$(curl -sS -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
  -D "$LOGIN_HEADERS" \
  -o "$LOGIN_BODY" \
  -w '%{http_code}' \
  -X POST "$BASE_URL/authorize/login" \
  --data-urlencode "email=alice@example.com" \
  --data-urlencode "password=correct-horse" \
  --data-urlencode "client_id=$CLIENT_ID" \
  --data-urlencode "redirect_uri=$REDIRECT_URI" \
  --data-urlencode "scope=openid profile" \
  --data-urlencode "state=xyz123" \
  --data-urlencode "code_challenge=$CODE_CHALLENGE" \
  --data-urlencode "code_challenge_method=S256")
RESP=$(<"$LOGIN_BODY")
grep -q "eris_session" "$COOKIE_JAR" \
  && pass "session cookie set" \
  || fail "no session cookie found in $COOKIE_JAR"

AUTH_CODE=""
LOCATION=""
if echo "$RESP" | grep -q "action=\"/authorize/consent\""; then
  pass "consent form returned"
elif [[ "$LOGIN_STATUS" =~ ^3 ]] && grep -qi '^location:' "$LOGIN_HEADERS"; then
  LOCATION=$(grep -i '^location:' "$LOGIN_HEADERS" | tr -d '\r')
  AUTH_CODE=$(echo "$LOCATION" | sed -n 's/.*code=\([^&]*\).*/\1/p')
  [ -n "$AUTH_CODE" ] \
    && pass "existing consent reused; authorization code returned" \
    || fail "login redirected without an authorization code: $LOCATION"
else
  fail "expected consent form or authorization redirect; status=$LOGIN_STATUS body=$RESP"
fi

echo
echo "== 4. POST /authorize/consent — expect redirect with code =="
if [ -z "$AUTH_CODE" ]; then
  LOCATION=$(curl -sS -i -b "$COOKIE_JAR" \
    -X POST "$BASE_URL/authorize/consent" \
    --data-urlencode "client_id=$CLIENT_ID" \
    --data-urlencode "redirect_uri=$REDIRECT_URI" \
    --data-urlencode "scope=openid profile" \
    --data-urlencode "state=xyz123" \
    --data-urlencode "code_challenge=$CODE_CHALLENGE" \
    --data-urlencode "code_challenge_method=S256" \
    --data-urlencode "approved_scope=openid" \
    --data-urlencode "approved_scope=profile" \
    | grep -i "^location:")
  AUTH_CODE=$(echo "$LOCATION" | sed -n 's/.*code=\([^&]*\).*/\1/p')
else
  pass "consent already recorded; consent POST skipped"
fi

[ -n "$AUTH_CODE" ] && pass "got code: $AUTH_CODE" || fail "no code in redirect: $LOCATION"

echo "$LOCATION" | grep -q "state=xyz123" \
  && pass "state round-tripped correctly" \
  || fail "state missing or mismatched"

echo
echo "== 5. POST /token (authorization_code) — expect access/refresh/id tokens =="
TOKEN_RESP=$(curl -s -X POST "$BASE_URL/token" \
  --data-urlencode "grant_type=authorization_code" \
  --data-urlencode "code=$AUTH_CODE" \
  --data-urlencode "redirect_uri=$REDIRECT_URI" \
  --data-urlencode "code_verifier=$CODE_VERIFIER" \
  "${TOKEN_CLIENT_ARGS[@]}")

if [ -z "$CLIENT_SECRET" ] && echo "$TOKEN_RESP" | grep -q '"error":"invalid_client"'; then
  fail "client $CLIENT_ID requires authentication; set CLIENT_SECRET to the one-time secret shown when the client was created in the admin portal"
fi

ACCESS_TOKEN=$(echo "$TOKEN_RESP" | python3 -c "import sys,json;print(json.load(sys.stdin).get('access_token',''))")
REFRESH_TOKEN=$(echo "$TOKEN_RESP" | python3 -c "import sys,json;print(json.load(sys.stdin).get('refresh_token',''))")
ID_TOKEN=$(echo "$TOKEN_RESP" | python3 -c "import sys,json;print(json.load(sys.stdin).get('id_token',''))")

[ -n "$ACCESS_TOKEN" ] && pass "access_token received" || fail "no access_token: $TOKEN_RESP"
[ -n "$REFRESH_TOKEN" ] && pass "refresh_token received" || fail "no refresh_token: $TOKEN_RESP"
[ -n "$ID_TOKEN" ] && pass "id_token received (openid scope worked)" || fail "no id_token: $TOKEN_RESP"

echo
echo "== 6. Negative test: reuse the same authorization code — expect failure =="
REUSE_RESP=$(curl -s -X POST "$BASE_URL/token" \
  --data-urlencode "grant_type=authorization_code" \
  --data-urlencode "code=$AUTH_CODE" \
  --data-urlencode "redirect_uri=$REDIRECT_URI" \
  --data-urlencode "code_verifier=$CODE_VERIFIER" \
  "${TOKEN_CLIENT_ARGS[@]}")
echo "$REUSE_RESP" | grep -q "invalid_grant" \
  && pass "reused code correctly rejected" \
  || fail "SECURITY BUG: reused code was accepted! $REUSE_RESP"

echo
echo "== 7. Negative test: wrong code_verifier — expect PKCE mismatch =="
# Fresh code needed since the previous one is now consumed
LOCATION2=$(curl -sS -i -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
  -X POST "$BASE_URL/authorize/consent" \
  --data-urlencode "client_id=$CLIENT_ID" \
  --data-urlencode "redirect_uri=$REDIRECT_URI" \
  --data-urlencode "scope=openid profile" \
  --data-urlencode "state=abc999" \
  --data-urlencode "code_challenge=$CODE_CHALLENGE" \
  --data-urlencode "code_challenge_method=S256" \
  --data-urlencode "approved_scope=openid" \
  --data-urlencode "approved_scope=profile" \
  | grep -i "^location:")
AUTH_CODE_2=$(echo "$LOCATION2" | sed -n 's/.*code=\([^&]*\).*/\1/p')

WRONG_VERIFIER_RESP=$(curl -s -X POST "$BASE_URL/token" \
  --data-urlencode "grant_type=authorization_code" \
  --data-urlencode "code=$AUTH_CODE_2" \
  --data-urlencode "redirect_uri=$REDIRECT_URI" \
  --data-urlencode "code_verifier=this-is-the-wrong-verifier" \
  "${TOKEN_CLIENT_ARGS[@]}")
echo "$WRONG_VERIFIER_RESP" | grep -q "invalid_grant" \
  && pass "wrong code_verifier correctly rejected" \
  || fail "SECURITY BUG: wrong verifier was accepted! $WRONG_VERIFIER_RESP"

echo
echo "== 8. POST /token (refresh_token) — expect a new access token =="
REFRESH_RESP=$(curl -s -X POST "$BASE_URL/token" \
  --data-urlencode "grant_type=refresh_token" \
  --data-urlencode "refresh_token=$REFRESH_TOKEN" \
  "${TOKEN_CLIENT_ARGS[@]}")
NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESP" | python3 -c "import sys,json;print(json.load(sys.stdin).get('access_token',''))")
[ -n "$NEW_ACCESS_TOKEN" ] && pass "refresh grant issued a new access token" || fail "refresh failed: $REFRESH_RESP"

echo
echo "== 9. Negative test: reuse the OLD refresh token — expect rejection (rotation) =="
OLD_REFRESH_REUSE=$(curl -s -X POST "$BASE_URL/token" \
  --data-urlencode "grant_type=refresh_token" \
  --data-urlencode "refresh_token=$REFRESH_TOKEN" \
  "${TOKEN_CLIENT_ARGS[@]}")
echo "$OLD_REFRESH_REUSE" | grep -q "invalid_grant" \
  && pass "rotated-out refresh token correctly rejected" \
  || fail "SECURITY BUG: old refresh token still worked! $OLD_REFRESH_REUSE"

echo
echo "🎉 All checks passed."
