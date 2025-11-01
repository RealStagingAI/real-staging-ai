#!/bin/bash

# Clean up old test subscriptions for a given Stripe customer
# Usage: ./cleanup_test_subscriptions.sh <customer_id>

if [ $# -eq 0 ]; then
    echo "Usage: $0 <stripe_customer_id>"
    echo "Example: $0 cus_TL6MHctSV2WPPH"
    exit 1
fi

CUSTOMER_ID="$1"

echo "Cleaning up subscriptions for customer: $CUSTOMER_ID"
echo "WARNING: This will cancel ALL subscriptions for this customer!"
echo "Press Ctrl+C to cancel, or Enter to continue..."
read

# Get Stripe CLI (assuming it's installed)
if ! command -v stripe &> /dev/null; then
    echo "Error: Stripe CLI not found. Please install it first."
    echo "Visit: https://stripe.com/docs/stripe-cli"
    exit 1
fi

# List all subscriptions for the customer
echo "Fetching subscriptions for customer $CUSTOMER_ID..."
stripe subscriptions list --customer "$CUSTOMER_ID" --limit 100

# Cancel each subscription (except the most recent active one)
echo ""
echo "Cancelling old subscriptions..."

# Get subscription IDs (skip the most recent one)
SUBSCRIPTIONS=$(stripe subscriptions list --customer "$CUSTOMER_ID" --limit 100 | grep -E "sub_[a-zA-Z0-9]+" | sort -r | tail -n +2)

for sub in $SUBSCRIPTIONS; do
    echo "Cancelling subscription: $sub"
    stripe subscriptions cancel "$sub"
done

echo ""
echo "Cleanup complete! Remaining subscriptions:"
stripe subscriptions list --customer "$CUSTOMER_ID" --limit 100
