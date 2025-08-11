// Test SherWare decryption
const SHERWARE_KEY = "SherWareKey_@8@2899909";

function transform(data, key) {
  if (!key.length || !data.length) return;

  // Calculate seed from key
  let seed = 0 >>> 0;
  for (let i = 0; i < key.length; i++) {
    seed = (seed + (key[i] + 1)) >>> 0;
  }

  // MSVCRT-compatible PRNG
  const rnd = () => {
    seed = (Math.imul(seed, 214013) + 2531011) >>> 0;
    return (seed >>> 16) & 0x7fff;
  };

  // Apply XOR transformation
  for (let i = 0; i < data.length; i++) {
    const r = rnd() & 0xff;
    data[i] ^= (r ^ key[i % key.length]) & 0xff;
  }
}

function decryptFromBase64(b64, keyStr = SHERWARE_KEY) {
  try {
    // Convert base64 to bytes
    const data = Buffer.from(b64, 'base64');
    const keyBytes = Buffer.from(keyStr);
    
    // Transform in place
    transform(data, keyBytes);
    
    // Convert back to string
    return data.toString('utf8');
  } catch (error) {
    console.error('Decryption failed:', error);
    return '[Decryption Error]';
  }
}

// Test with some sample encrypted values that might be in the DBF
const testValues = [
  'VGVzdDEyMw==',  // Simple base64
  'MDAxMjM0NTY3OA==',  // Another test
  'CTAXID_ENC_TEST',  // Not base64
];

console.log('Testing SherWare decryption:');
testValues.forEach(val => {
  console.log(`\nInput: ${val}`);
  console.log(`Decrypted: ${decryptFromBase64(val)}`);
});

// Test with a known tax ID pattern
const encryptToBase64 = (plain, keyStr = SHERWARE_KEY) => {
  const data = Buffer.from(plain, 'utf8');
  const keyBytes = Buffer.from(keyStr);
  transform(data, keyBytes);
  return data.toString('base64');
};

// Encrypt a sample tax ID
const sampleTaxId = '123-45-6789';
const encrypted = encryptToBase64(sampleTaxId);
console.log(`\nOriginal Tax ID: ${sampleTaxId}`);
console.log(`Encrypted: ${encrypted}`);
console.log(`Decrypted back: ${decryptFromBase64(encrypted)}`);