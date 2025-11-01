#!/bin/bash

# Production cleanup using direct API calls
# Usage: ./cleanup_production.sh <customer_id> <api_key>

if [ $# -ne 2 ]; then
    echo "Usage: $0 <customer_id> <api_key>"
    echo "Example: $0 cus_PRODUCTION123 sk_live_xxxx"
    exit 1
fi

CUSTOMER_ID="$1"
API_KEY="$2"

echo "WARNING: This will cancel ALL subscriptions for customer: $CUSTOMER_ID"
echo "Press Ctrl+C to cancel, or Enter to continue..."
read

# Get all subscriptions for the customer
echo "Fetching subscriptions..."
SUBSCRIPTIONS=$(curl -s -u "$API_KEY:" \
  "https://api.stripe.com/v1/subscriptions?customer=$CUSTOMER_ID&limit=100" | \
  jq -r '.data[].id')

if [ -z "$SUBSCRIPTIONS" ]; then
    echo "No subscriptions found."
    exit 0
fi

echo "Found subscriptions:"
echo "$SUBSCRIPTIONS"

# Cancel each subscription
for sub in $SUBSCRIPTIONS; do
    echo "Cancelling subscription: $sub"
    curl -s -u "$API_KEY:" \
      -X POST \
      "https://api.stripe.com/v1/subscriptions/$sub/cancel"
done

echo ""
echo "Cleanup complete!"
