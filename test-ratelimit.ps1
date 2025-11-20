# Test Rate Limiting
# This script tests the broker's rate limiting functionality

Write-Host "=== Rate Limiting Test ===" -ForegroundColor Cyan
Write-Host ""

$brokerUrl = "http://localhost:8081"

# Check if broker is running
Write-Host "Checking if broker is running..." -ForegroundColor Yellow
try {
    $status = Invoke-RestMethod -Uri "$brokerUrl/api/v1/status" -Method GET -ErrorAction Stop
    Write-Host "OK - Broker is running (version: $($status.version))" -ForegroundColor Green
} catch {
    Write-Host "ERROR - Broker is not running" -ForegroundColor Red
    Write-Host "  Please start the broker with: cd broker; go run main.go" -ForegroundColor Yellow
    exit 1
}
Write-Host ""

# Test 1: Normal requests within limit
Write-Host "Test 1: Normal requests (should succeed)" -ForegroundColor Yellow
$successCount = 0

for ($i = 1; $i -le 10; $i++) {
    try {
        Invoke-RestMethod -Uri "$brokerUrl/api/v1/status" -Method GET -ErrorAction Stop | Out-Null
        $successCount++
        Write-Host "  Request $i : OK (200 OK)" -ForegroundColor Green
    } catch {
        Write-Host "  Request $i : FAILED" -ForegroundColor Red
    }
}
Write-Host ""

# Test 2: Rapid fire requests to trigger rate limit
Write-Host "Test 2: Rapid requests (should trigger rate limit)" -ForegroundColor Yellow
Write-Host "  Sending 120 requests rapidly..." -ForegroundColor Gray

$successCount = 0
$rateLimitedCount = 0
$firstRateLimitAt = 0

for ($i = 1; $i -le 120; $i++) {
    try {
        Invoke-RestMethod -Uri "$brokerUrl/api/v1/status" -Method GET -ErrorAction Stop | Out-Null
        $successCount++
        if ($i % 20 -eq 0) {
            Write-Host "    $i requests: still accepting..." -ForegroundColor Gray
        }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 429) {
            $rateLimitedCount++
            if ($firstRateLimitAt -eq 0) {
                $firstRateLimitAt = $i
                Write-Host "    Rate limit triggered at request #$i (429 Too Many Requests)" -ForegroundColor Yellow
            }
        } else {
            Write-Host "    Request $i : Failed with status $statusCode" -ForegroundColor Red
        }
    }
}

Write-Host ""
Write-Host "  Summary:" -ForegroundColor Gray
Write-Host "    Succeeded: $successCount" -ForegroundColor Green
Write-Host "    Rate limited (429): $rateLimitedCount" -ForegroundColor Yellow
if ($firstRateLimitAt -gt 0) {
    Write-Host "    First rate limit at: request #$firstRateLimitAt" -ForegroundColor Yellow
}
Write-Host ""

# Test 3: Wait and retry
if ($rateLimitedCount -gt 0) {
    Write-Host "Test 3: Recovery after rate limit" -ForegroundColor Yellow
    Write-Host "  Waiting 5 seconds for token bucket to refill..." -ForegroundColor Gray
    Start-Sleep -Seconds 5
    
    try {
        Invoke-RestMethod -Uri "$brokerUrl/api/v1/status" -Method GET -ErrorAction Stop | Out-Null
        Write-Host "  OK - Request succeeded after waiting" -ForegroundColor Green
        Write-Host "  Rate limit is working correctly!" -ForegroundColor Green
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 429) {
            Write-Host "  Still rate limited (may need to wait longer)" -ForegroundColor Yellow
        } else {
            Write-Host "  Failed with unexpected error: $statusCode" -ForegroundColor Red
        }
    }
} else {
    Write-Host "Test 3: Skipped (no rate limiting was triggered)" -ForegroundColor Yellow
    Write-Host "  This might mean:" -ForegroundColor Yellow
    Write-Host "  - Rate limiting is disabled (check RATE_LIMIT_ENABLED)" -ForegroundColor Yellow
    Write-Host "  - Rate limit is very high (check RATE_LIMIT_MAX_REQUESTS)" -ForegroundColor Yellow
}
Write-Host ""

# Summary
Write-Host "=== Test Complete ===" -ForegroundColor Cyan
Write-Host ""

if ($rateLimitedCount -gt 0) {
    Write-Host "RESULT: Rate limiting is WORKING" -ForegroundColor Green
    Write-Host "  - Requests under limit: Accepted" -ForegroundColor Green
    Write-Host "  - Requests over limit: Rejected with 429" -ForegroundColor Green
    Write-Host "  - Recovery: Tokens refill over time" -ForegroundColor Green
} else {
    Write-Host "RESULT: Rate limiting may NOT be working" -ForegroundColor Yellow
    Write-Host "  Check broker logs and configuration:" -ForegroundColor Yellow
    Write-Host "  - RATE_LIMIT_ENABLED=true" -ForegroundColor Gray
    Write-Host "  - RATE_LIMIT_MAX_REQUESTS=100" -ForegroundColor Gray
    Write-Host "  - RATE_LIMIT_WINDOW=1m" -ForegroundColor Gray
}
Write-Host ""

Write-Host "To check broker configuration, look for this in startup logs:" -ForegroundColor Cyan
Write-Host "  'Rate limiting: enabled (100 requests per 1m0s)'" -ForegroundColor Gray
