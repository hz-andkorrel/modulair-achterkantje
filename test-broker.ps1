# Test script for broker service
# This script tests the /api/v1/status endpoint with and without authentication

Write-Host "=== Broker Service Tests ===" -ForegroundColor Cyan
Write-Host ""

$brokerUrl = "http://localhost:8081"
$authUrl = "http://localhost:8080"

# Test 1: Status without authentication
Write-Host "Test 1: GET /api/v1/status (without auth)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$brokerUrl/api/v1/status" -Method GET
    Write-Host "✓ Status endpoint accessible without auth" -ForegroundColor Green
    Write-Host "  Status: $($response.status)" -ForegroundColor Gray
    Write-Host "  Authenticated: $($response.authenticated)" -ForegroundColor Gray
    Write-Host "  Version: $($response.version)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Failed to access status endpoint" -ForegroundColor Red
    Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 2: Status with authentication (requires auth-service to be running)
Write-Host "Test 2: GET /api/v1/status (with auth)" -ForegroundColor Yellow
try {
    # First, try to login to get a token
    Write-Host "  Attempting login to auth-service..." -ForegroundColor Gray
    $loginBody = @{
        email = "admin@example.com"
        password = "admin123"
    } | ConvertTo-Json
    
    $loginResponse = Invoke-RestMethod -Uri "$authUrl/api/auth/login" -Method POST -Headers @{"Content-Type"="application/json"} -Body $loginBody
    $token = $loginResponse.access_token
    Write-Host "  ✓ Login successful, token obtained" -ForegroundColor Green
    
    # Now call status with the token
    $headers = @{
        "Authorization" = "Bearer $token"
    }
    $response = Invoke-RestMethod -Uri "$brokerUrl/api/v1/status" -Method GET -Headers $headers
    Write-Host "✓ Status endpoint accessible with auth" -ForegroundColor Green
    Write-Host "  Status: $($response.status)" -ForegroundColor Gray
    Write-Host "  Authenticated: $($response.authenticated)" -ForegroundColor Gray
    Write-Host "  User ID: $($response.user.id)" -ForegroundColor Gray
    Write-Host "  User Email: $($response.user.email)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Failed to access status endpoint with auth" -ForegroundColor Red
    Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "  Note: Make sure auth-service is running on port 8080" -ForegroundColor Yellow
}
Write-Host ""

# Test 3: Invalid token
Write-Host "Test 3: GET /api/v1/status (with invalid token)" -ForegroundColor Yellow
try {
    $headers = @{
        "Authorization" = "Bearer invalid-token-here"
    }
    $response = Invoke-RestMethod -Uri "$brokerUrl/api/v1/status" -Method GET -Headers $headers
    Write-Host "✓ Status endpoint handles invalid token gracefully" -ForegroundColor Green
    Write-Host "  Status: $($response.status)" -ForegroundColor Gray
    Write-Host "  Authenticated: $($response.authenticated)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Unexpected error with invalid token" -ForegroundColor Red
    Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Write-Host "=== Tests Complete ===" -ForegroundColor Cyan
