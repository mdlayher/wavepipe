<?php

require_once __DIR__ . "/wavepipe.php";

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
	"/api/v0/albums/1",
	"/api/v0/artists",
	"/api/v0/artists/1",
	"/api/v0/folders",
	"/api/v0/folders/1",
	"/api/v0/songs",
	"/api/v0/songs/1",
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
