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

