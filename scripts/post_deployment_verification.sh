#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

pass() { echo "✅ PASS: $*"; }
fail() { echo "❌ FAIL: $*"; exit 1; }

# curl_json METHOD PATH JSON EXPECTED_HTTP_CODE EXPECTED_BODY_SUBSTRING
curl_json() {
  local method="$1"
  local path="$2"
  local json="$3"
  local expected_code="$4"
  local expected_substr="$5"

  local tmp_headers tmp_body
  tmp_headers="$(mktemp)"
  tmp_body="$(mktemp)"
  trap 'rm -f "$tmp_headers" "$tmp_body"' RETURN

  local http_code
  http_code="$(curl -sS -X "$method" "${BASE_URL}${path}" \
    -H "Content-Type: application/json" \
    -d "$json" \
    -D "$tmp_headers" \
    -o "$tmp_body" \
    -w "%{http_code}")" || fail "curl failed for $method $path"

  local body
  body="$(cat "$tmp_body")"

  if [[ "$http_code" != "$expected_code" ]]; then
    echo "---- Response headers ----"
    cat "$tmp_headers"
    echo "---- Response body ----"
    echo "$body"
    echo "-------------------------"
    fail "$method $path expected HTTP $expected_code, got $http_code"
  fi

  echo "$body" | grep -qF "$expected_substr" || {
    echo "---- Response headers ----"
    cat "$tmp_headers"
    echo "---- Response body ----"
    echo "$body"
    echo "-------------------------"
    fail "$method $path expected body to contain: $expected_substr"
  }

  # show correlation id if present
  local cid
  cid="$(grep -i '^X-Correlation-Id:' "$tmp_headers" | head -n 1 | awk -F': ' '{print $2}' | tr -d '\r' || true)"
  if [[ -n "$cid" ]]; then
    pass "$method $path (HTTP $http_code, correlation_id=$cid)"
  else
    pass "$method $path (HTTP $http_code)"
  fi
}

# Helper for transfer requests (keeps quoting/escaping sane)
transfer_json() {
  local src="$1"
  local dst="$2"
  local amt="$3"
  echo "{\"source_account_id\":${src},\"destination_account_id\":${dst},\"amount\":\"${amt}\"}"
}

echo "Running Post Deployment Verification against: $BASE_URL"
echo

# Use a unique ID so this can be run repeatedly without manual cleanup
RUN_ID="$(date +%s)"
ACCOUNT_OK="$((100 + RUN_ID % 1000000))"
ACCOUNT_DUP="$((ACCOUNT_OK + 1))"
ACCOUNT_NEG="$((ACCOUNT_OK + 2))"

# 1️⃣ Create account - success
curl_json POST /accounts \
"{\"account_id\":${ACCOUNT_OK},\"balance\":\"5000.00\"}" \
"201" \
"\"success\":true"

# 2️⃣ Duplicate account - should fail with 409
curl_json POST /accounts \
"{\"account_id\":${ACCOUNT_DUP},\"balance\":\"5000.00\"}" \
"201" \
"\"success\":true"
curl_json POST /accounts \
"{\"account_id\":${ACCOUNT_DUP},\"balance\":\"5000.00\"}" \
"409" \
"Account already exists"

# 3️⃣ Negative balance - should fail with 400
curl_json POST /accounts \
"{\"account_id\":${ACCOUNT_NEG},\"balance\":\"-5000.00\"}" \
"400" \
"amount must be greater than or equal to zero"

# 4️⃣ Create destination account (balance = 0)
ACCOUNT_DEST="$((ACCOUNT_OK + 10))"

curl_json POST /accounts \
"{\"account_id\":${ACCOUNT_DEST},\"balance\":\"0.00\"}" \
"201" \
"\"success\":true"

# 5️⃣ Transfer funds (happy path)
TRANSFER_AMOUNT="100.00"

curl_json POST /transfers \
"$(transfer_json "${ACCOUNT_OK}" "${ACCOUNT_DEST}" "${TRANSFER_AMOUNT}")" \
"200" \
"\"success\":true"

# =========================
# Transfer validation cases
# =========================

# 6️⃣ source_account_id or destination_account_id is 0 / negative
curl_json POST /transfers \
"$(transfer_json 0 "${ACCOUNT_DEST}" "10.00")" \
"400" \
"invalid account_id: must be positive"

curl_json POST /transfers \
"$(transfer_json -1 "${ACCOUNT_DEST}" "10.00")" \
"400" \
"invalid account_id: must be positive"

curl_json POST /transfers \
"$(transfer_json "${ACCOUNT_OK}" 0 "10.00")" \
"400" \
"invalid account_id: must be positive"

curl_json POST /transfers \
"$(transfer_json "${ACCOUNT_OK}" -2 "10.00")" \
"400" \
"invalid account_id: must be positive"

# 7️⃣ both account IDs are same
curl_json POST /transfers \
"$(transfer_json "${ACCOUNT_OK}" "${ACCOUNT_OK}" "10.00")" \
"400" \
"source and destination"

# 8️⃣ amount greater than balance (pick something huge)
curl_json POST /transfers \
"$(transfer_json "${ACCOUNT_OK}" "${ACCOUNT_DEST}" "999999.00")" \
"422" \
"insufficient"

# 9️⃣ amount is negative
curl_json POST /transfers \
"$(transfer_json "${ACCOUNT_OK}" "${ACCOUNT_DEST}" "-10.00")" \
"400" \
"amount"

echo
echo "🎉 Post Deployment Verification completed successfully."
