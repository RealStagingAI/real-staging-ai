#!/bin/bash

# Clean up old subscriptions, keeping only the most recent one
# Usage: ./cleanup_old_subscriptions.sh <customer_id> [api_key]

if [ $# -eq 0 ]; then
    echo "Usage: $0 <stripe_customer_id> [api_key]"
    echo "Example: $0 cus_TL6MHctSV2WPPH"
    echo "Example (production): $0 cus_PRODUCTION123 sk_live_xxxx"
    exit 1
fi

CUSTOMER_ID="$1"
API_KEY="$2"

echo "Cleaning up OLD subscriptions for customer: $CUSTOMER_ID"
echo "This will cancel all subscriptions EXCEPT the most recent active one."
echo "Press Ctrl+C to cancel, or Enter to continue..."
read

# Function to run stripe command with optional API key
run_stripe() {
    if [ -n "$API_KEY" ]; then
        stripe --api-key "$API_KEY" "$@"
    else
        stripe "$@"
    fi
}

# Check if required tools are available
if ! command -v stripe &> /dev/null; then
    echo "Error: Stripe CLI not found. Please install it first."
    echo "Visit: https://stripe.com/docs/stripe-cli"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "Error: jq not found. Please install it first."
    echo "On macOS: brew install jq"
    echo "On Ubuntu: sudo apt-get install jq"
    exit 1
fi

# Test the connection first
echo "Testing Stripe connection..."
if ! run_stripe balance &> /dev/null; then
    echo "Error: Failed to connect to Stripe. Please check your authentication."
    if [ -z "$API_KEY" ]; then
        echo "Try running 'stripe login' or provide an API key as the second argument."
    fi
    exit 1
fi

# List all subscriptions for the customer
echo "Fetching subscriptions for customer $CUSTOMER_ID..."
run_stripe subscriptions list --customer "$CUSTOMER_ID" --limit 100

# Get all subscription IDs, sorted by creation date (newest first)
# Use jq to properly extract subscription IDs from JSON
ALL_SUBSCRIPTIONS=$(run_stripe subscriptions list --customer "$CUSTOMER_ID" --limit 100 | jq -r '.data[].id' | sort -r)

# Keep the most recent one, cancel the rest
if [ -z "$ALL_SUBSCRIPTIONS" ]; then
    echo "No subscriptions found."
    exit 0
fi

# Get the most recent subscription (first one in the list)
MOST_RECENT=$(echo "$ALL_SUBSCRIPTIONS" | head -n 1)
echo "Keeping most recent subscription: $MOST_RECENT"

# Cancel all others
echo ""
echo "Cancelling old subscriptions..."
CANCEL_COUNT=0

for sub in $ALL_SUBSCRIPTIONS; do
    if [ "$sub" != "$MOST_RECENT" ]; then
        echo "Cancelling subscription: $sub"
        run_stripe subscriptions cancel "$sub"
        CANCEL_COUNT=$((CANCEL_COUNT + 1))
    fi
done

echo ""
echo "Cleanup complete! Cancelled $CANCEL_COUNT old subscriptions."
echo "Remaining subscription for customer $CUSTOMER_ID:"
run_stripe subscriptions list --customer "$CUSTOMER_ID" --limit 100
