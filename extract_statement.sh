#!/bin/bash
# Extract structured data from bank and card statements using the Photon Commerce API.
#
# Submits a statement (PDF or image) and returns account details, balances,
# and a full transaction ledger including client name, bank name, account number,
# starting/ending balances, and all transactions (date, type, description, amount).
#
# Processing times (Managed Agents):
#   Trial accounts:  up to 24 hours
#   Production:      5 minutes to 24 hours
#
# AI extraction (seconds, no Managed Agents):
#   Contact support@photoncommerce.com to activate.
#   Once active, submit to /api/v4 instead of /api/pro.
#
# Docs:    https://apidocs.photoncommerce.com
# Sandbox: https://sandbox-api.photoncommerce.com/api/v4/register (20 free calls)

CLIENT_ID="YOUR_CLIENT_ID"
USERNAME="YOUR_USERNAME"
API_KEY="YOUR_API_KEY"
PASSWORD="YOUR_PASSWORD"
SECRET_KEY="YOUR_SECRET_KEY"

# Sandbox: https://sandbox-api.photoncommerce.com  (20 free calls, no card needed)
# Production: https://api.photoncommerce.com
BASE_URL="https://sandbox-api.photoncommerce.com"

# For AI extraction (seconds), replace /api/pro with /api/v4 — contact support@photoncommerce.com to activate.

# --- Option A: submit from a local file ---
RESPONSE=$(curl -s -X POST "$BASE_URL/api/pro?doctype=statement" \
  -H "CLIENT-ID: $CLIENT_ID" \
  -H "AUTHORIZATION: apikey $USERNAME:$API_KEY" \
  -H "PASSWORD: $PASSWORD" \
  -H "SECRET-KEY: $SECRET_KEY" \
  -F "pdf=@statement.pdf")

# --- Option B: submit via a publicly accessible URL ---
# RESPONSE=$(curl -s -X POST "$BASE_URL/api/pro?doctype=statement&url=https://example.com/statement.pdf" \
#   -H "CLIENT-ID: $CLIENT_ID" \
#   -H "AUTHORIZATION: apikey $USERNAME:$API_KEY" \
#   -H "PASSWORD: $PASSWORD" \
#   -H "SECRET-KEY: $SECRET_KEY")

# Optional parameters (append to the query string as needed):
#   &webhook_url=https://your-server.com/webhook
#   &auth_token=YOUR_TOKEN
#   &ID=YOUR_CUSTOM_ID
#   &subaccount=SUBACCOUNT_NAME
#   &page_start=1&page_end=1

PHOTON_KEY=$(echo "$RESPONSE" | grep -o '"photon_key":"[^"]*"' | cut -d'"' -f4)
echo "Submitted. photon_key: $PHOTON_KEY"
echo "Waiting for extraction to complete..."

POLL_INTERVAL=20
while true; do
  RESULT=$(curl -s "$BASE_URL/api/v4/json?photon_key=$PHOTON_KEY" \
    -H "CLIENT-ID: $CLIENT_ID" \
    -H "AUTHORIZATION: apikey $USERNAME:$API_KEY" \
    -H "PASSWORD: $PASSWORD" \
    -H "SECRET-KEY: $SECRET_KEY")

  STATUS=$(echo "$RESULT" | grep -o '"Status":"[^"]*"' | cut -d'"' -f4)
  if [ "$STATUS" != "pending" ] && [ "$STATUS" != "processing" ] && [ -n "$STATUS" ]; then
    break
  fi
  echo "  Status: ${STATUS:-pending} — retrying in ${POLL_INTERVAL}s..."
  sleep $POLL_INTERVAL
done

echo ""
echo "--- Bank Statement Data ---"
echo "$RESULT" | python3 -c "
import sys, json
data = json.load(sys.stdin).get('data', {})
print('Client:          ', data.get('client_name'))
print('Bank:            ', data.get('bank_name'))
print('Account Number:  ', data.get('account_number'))
print('Account Type:    ', data.get('account_type'))
print('Period:          ', data.get('statement_start_date'), '->', data.get('statement_end_date'))
print('Starting Balance:', data.get('starting_balance'))
print('Ending Balance:  ', data.get('ending_balance'))
print('Total Credits:   ', data.get('tot_credit'))
print('Total Debits:    ', data.get('tot_debit'))
txns = data.get('transactions', [])
if txns:
    print(f'\n--- Transactions ({len(txns)}) ---')
    for t in txns:
        print(f'  {t.get(\"date\")}  {t.get(\"type\"):<6}  {t.get(\"amount\"):>10}  {t.get(\"description\")}')
"
