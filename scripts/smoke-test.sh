#!/bin/bash
set -e

BASE_URL="${BASE_URL:-http://localhost:3000}"

echo "=== Smoke Test: Users API ==="
echo "Base URL: $BASE_URL"
echo

# Health check
echo "1. Health check..."
curl -s "$BASE_URL/health" | grep -q "ok" && echo "   ✓ Health check passed" || { echo "   ✗ Health check failed"; exit 1; }

# Create user
echo "2. Creating user..."
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/users" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test User", "email": "test@example.com", "password": "secret123"}')
USER_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')

if [ -n "$USER_ID" ]; then
  echo "   ✓ User created with ID: $USER_ID"
else
  echo "   ✗ Failed to create user"
  echo "   Response: $CREATE_RESPONSE"
  exit 1
fi

# Verify password not returned
if echo "$CREATE_RESPONSE" | grep -q '"password"'; then
  echo "   ✗ Password should not be in response"
  exit 1
else
  echo "   ✓ Password correctly omitted from response"
fi

# List users
echo "3. Listing users..."
LIST_RESPONSE=$(curl -s "$BASE_URL/users")
if echo "$LIST_RESPONSE" | grep -q '"pagination"'; then
  echo "   ✓ List returns pagination metadata"
else
  echo "   ✗ Missing pagination metadata"
  exit 1
fi

# Get user by ID
echo "4. Getting user by ID..."
GET_RESPONSE=$(curl -s "$BASE_URL/users/$USER_ID")
if echo "$GET_RESPONSE" | grep -q "Test User"; then
  echo "   ✓ User retrieved successfully"
else
  echo "   ✗ Failed to get user"
  exit 1
fi

# Update user (PUT)
echo "5. Updating user (PUT)..."
UPDATE_RESPONSE=$(curl -s -X PUT "$BASE_URL/users/$USER_ID" \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated User"}')
if echo "$UPDATE_RESPONSE" | grep -q "Updated User"; then
  echo "   ✓ User updated successfully"
else
  echo "   ✗ Failed to update user"
  exit 1
fi

# Partial update (PATCH)
echo "6. Partial update (PATCH)..."
PATCH_RESPONSE=$(curl -s -X PATCH "$BASE_URL/users/$USER_ID" \
  -H "Content-Type: application/json" \
  -d '{"email": "updated@example.com"}')
if echo "$PATCH_RESPONSE" | grep -q "updated@example.com"; then
  echo "   ✓ User patched successfully"
else
  echo "   ✗ Failed to patch user"
  exit 1
fi

# Delete user
echo "7. Deleting user..."
DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/users/$USER_ID")
if echo "$DELETE_RESPONSE" | grep -q "deleted"; then
  echo "   ✓ User deleted successfully"
else
  echo "   ✗ Failed to delete user"
  exit 1
fi

# Verify deletion (should 404)
echo "8. Verifying deletion..."
GET_DELETED=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/users/$USER_ID")
if [ "$GET_DELETED" = "404" ]; then
  echo "   ✓ Deleted user returns 404"
else
  echo "   ✗ Expected 404, got $GET_DELETED"
  exit 1
fi

# Test validation (invalid email)
echo "9. Testing validation (invalid email)..."
INVALID_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/users" \
  -H "Content-Type: application/json" \
  -d '{"name": "Test", "email": "invalid", "password": "secret"}')
if [ "$INVALID_RESPONSE" = "400" ]; then
  echo "   ✓ Invalid email returns 400"
else
  echo "   ✗ Expected 400, got $INVALID_RESPONSE"
  exit 1
fi

echo
echo "=== All smoke tests passed! ==="
