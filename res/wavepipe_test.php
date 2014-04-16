<?php

// apiSignature creates a HMAC-SHA1 API signature
function apiSignature($public, $nonce, $method, $resource, $secret) {
	// Create API signature string
	$signString = sprintf("%s-%s-%s-%s", $public, $nonce, $method, $resource);

	// Return HMAC-SHA1 signature
	return hash_hmac("sha1", $signString, $secret);
}

// generateNonce creates a nonce value for use with the API
function generateNonce($length = 10) {
	$characters = '0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ';
	$randomString = '';
	for ($i = 0; $i < $length; $i++) {
		$randomString .= $characters[rand(0, strlen($characters) - 1)];
	}

	return $randomString;
}

// Attempt a login request using test credentials
$login = json_decode(file_get_contents("http://localhost:8080/api/v0/login?u=test&p=test"), true);
if (empty($login)) {
	printf("Failed to decode login JSON");
	exit(1);
}

// Store necessary login information
$publicKey = $login["session"]["publicKey"];
$secretKey = $login["session"]["secretKey"];

// Iterate and test all JSON APIs
$apiCalls = array(
	"/api/v0/albums",
	"/api/v0/artists",
	"/api/v0/songs",
	"/api/v0/logout",
);

foreach ($apiCalls as $a) {
	// Create a nonce
	$nonce = generateNonce();

	// Create the necessary API signature
	$signature = apiSignature($publicKey, $nonce, "GET", $a, $secretKey);

	// Generate URL
	$url = sprintf("http://localhost:8080%s?s=%s:%s:%s", $a, $publicKey, $nonce, $signature);

	// Perform API call
	printf("%s: %s\n", $a, file_get_contents($url));
}
