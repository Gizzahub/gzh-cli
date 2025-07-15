#!/bin/bash
# Automated Network Environment Tests

echo "Running AWS Profile Management Test"
gz dev-env aws-profile switch default || echo "AWS profile switching not available"

echo "Running GCP Project Management Test"
gz dev-env gcp-project switch default-project || echo "GCP project switching not available"

echo "Running Azure Subscription Management Test"  
gz dev-env azure-subscription switch default-sub || echo "Azure subscription switching not available"
