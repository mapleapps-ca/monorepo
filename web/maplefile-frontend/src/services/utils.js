// UUID v4 Generator - Compatible with Cassandra
// Generates a random UUID v4 in the format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx

function generateUUID() {
  // Use crypto.getRandomValues for cryptographically secure random numbers
  const array = new Uint8Array(16);
  crypto.getRandomValues(array);

  // Set version (4) and variant bits according to UUID v4 spec
  array[6] = (array[6] & 0x0f) | 0x40; // Version 4
  array[8] = (array[8] & 0x3f) | 0x80; // Variant 10

  // Convert to hex string with proper formatting
  const hex = Array.from(array, (byte) =>
    byte.toString(16).padStart(2, "0"),
  ).join("");

  // Insert hyphens at the correct positions
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20, 32)}`;
}

// Alternative implementation using crypto.randomUUID (if available in browser)
function generateUUIDNative() {
  // This is the preferred method if the browser supports it (most modern browsers do)
  if (crypto.randomUUID) {
    return crypto.randomUUID();
  }
  // Fallback to manual implementation
  return generateUUID();
}
