/**
 * SherWare-Compatible Encryption/Decryption
 * This is a symmetric XOR stream cipher using MSVCRT's rand() PRNG
 * Compatible with legacy FoxPro FLL encryption
 * 
 * ⚠️ Security Notice: This is not cryptographically secure and should
 * only be used for legacy compatibility with existing encrypted data
 */

import logger from '../services/logger';

// Production encryption key
const SHERWARE_KEY = "SherWareKey_@8@2899909";

/**
 * Transform data using SherWare encryption algorithm
 * Same function is used for both encryption and decryption
 */
export function transform(data: Uint8Array, key: Uint8Array): void {
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

/**
 * UTF-8 encoding/decoding utilities
 */
export const utf8 = {
  enc: (s: string) => new TextEncoder().encode(s),
  dec: (b: Uint8Array) => new TextDecoder().decode(b),
};

/**
 * Convert Uint8Array to Base64 string
 */
export function toBase64(b: Uint8Array): string {
  if (typeof window === 'undefined') {
    return Buffer.from(b).toString('base64');
  }
  let s = '';
  for (let i = 0; i < b.length; i++) {
    s += String.fromCharCode(b[i]);
  }
  return btoa(s);
}

/**
 * Convert Base64 string to Uint8Array
 */
export function fromBase64(s: string): Uint8Array {
  if (typeof window === 'undefined') {
    return new Uint8Array(Buffer.from(s, 'base64'));
  }
  const bin = atob(s);
  const out = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) {
    out[i] = bin.charCodeAt(i);
  }
  return out;
}

/**
 * Encrypt plain text to Base64-encoded ciphertext
 */
export function encryptToBase64(plain: string, keyStr: string = SHERWARE_KEY): string {
  const buf = utf8.enc(plain);
  const out = new Uint8Array(buf);
  transform(out, utf8.enc(keyStr));
  return toBase64(out);
}

/**
 * Decrypt Base64-encoded ciphertext to plain text
 */
export function decryptFromBase64(b64: string, keyStr: string = SHERWARE_KEY): string {
  try {
    const data = fromBase64(b64);
    transform(data, utf8.enc(keyStr));
    return utf8.dec(data);
  } catch (error) {
    logger.error('Base64 decryption failed', { error: error.message, input: b64.substring(0, 20) });
    return '[Decryption Error]';
  }
}

/**
 * Check if a value looks like an encrypted CTAXID
 * Encrypted tax IDs can be either base64 encoded or raw binary
 */
export function isEncryptedTaxId(value: any): boolean {
  if (typeof value !== 'string') return false;
  
  // Trim like FoxPro ALLTRIM()
  const trimmed = value.trim();
  if (!trimmed || trimmed === '') return false;
  
  // Check if it looks like base64
  const base64Pattern = /^[A-Za-z0-9+/]+=*$/;
  const isBase64 = base64Pattern.test(trimmed) && trimmed.length >= 12 && trimmed.length <= 30;
  
  // Check if it contains non-printable characters (raw binary)
  const hasNonPrintable = /[\x00-\x08\x0B-\x0C\x0E-\x1F\x7F-\xFF]/.test(trimmed);
  
  // Check if it looks like a normal tax ID (should not be encrypted)
  const normalTaxIdPattern = /^\d{9}$|^\d{3}-\d{2}-\d{4}$|^\d{2}-\d{7}$/;
  const isNormalTaxId = normalTaxIdPattern.test(trimmed);
  
  return !isNormalTaxId && (isBase64 || hasNonPrintable);
}

/**
 * Decrypt raw binary data using SherWare algorithm
 */
export function decryptFromBinary(data: Uint8Array, keyStr: string = SHERWARE_KEY): string {
  try {
    const dataCopy = new Uint8Array(data);
    transform(dataCopy, utf8.enc(keyStr));
    return utf8.dec(dataCopy);
  } catch (error) {
    logger.error('Binary decryption failed', { error: error.message, dataLength: data.length });
    return '[Decryption Error]';
  }
}

/**
 * Convert string with potential binary data to Uint8Array
 */
function stringToBytes(str: string): Uint8Array {
  const bytes = new Uint8Array(str.length);
  for (let i = 0; i < str.length; i++) {
    bytes[i] = str.charCodeAt(i) & 0xff;
  }
  return bytes;
}

/**
 * Decrypt a tax ID value if it appears to be encrypted
 * Returns the decrypted value or the original if not encrypted
 */
export function decryptTaxId(value: any): string {
  if (!value) return '';
  
  // IMPORTANT: Trim spaces like FoxPro's ALLTRIM() function
  // DBF files store fixed-width fields padded with spaces
  const trimmedValue = String(value).trim();
  
  // Debug logging with more detail
  const hexBytes = trimmedValue.split('').map((c: string) => '0x' + c.charCodeAt(0).toString(16).padStart(2, '0')).join(' ');
  const charCodes = trimmedValue.split('').map((c: string) => c.charCodeAt(0));
  
  logger.info('=== TAX ID DECRYPTION ATTEMPT ===', {
    originalValue: value,
    originalLength: value ? value.length : 0,
    trimmedValue: trimmedValue,
    trimmedLength: trimmedValue.length,
    type: typeof trimmedValue,
    charCodes: charCodes,
    hex: hexBytes,
    firstByte: charCodes[0] ? '0x' + charCodes[0].toString(16) : 'none',
    isEncrypted: isEncryptedTaxId(trimmedValue)
  });
  
  // If it's not a string or doesn't look encrypted, return as-is
  if (!isEncryptedTaxId(trimmedValue)) {
    logger.debug('Value does not appear to be encrypted', { value: trimmedValue });
    return trimmedValue;
  }
  
  try {
    let decrypted: string;
    
    // Check if it's base64 encoded
    const base64Pattern = /^[A-Za-z0-9+/]+=*$/;
    if (base64Pattern.test(trimmedValue)) {
      logger.debug('Attempting base64 decryption');
      // Try base64 decryption
      decrypted = decryptFromBase64(trimmedValue);
    } else {
      logger.debug('Attempting raw binary decryption');
      // Try raw binary decryption
      const bytes = stringToBytes(trimmedValue);
      logger.debug('Byte array', { bytes: Array.from(bytes) });
      decrypted = decryptFromBinary(bytes);
    }
    
    logger.debug('Raw decrypted result', {
      decrypted: decrypted,
      charCodes: decrypted.split('').map(c => c.charCodeAt(0)),
      hex: decrypted.split('').map(c => c.charCodeAt(0).toString(16).padStart(2, '0')).join(' ')
    });
    
    // Clean up any non-printable characters
    decrypted = decrypted.replace(/[\x00-\x08\x0B-\x0C\x0E-\x1F\x7F-\xFF]/g, '');
    
    logger.debug('Cleaned decrypted value', { decrypted: decrypted });
    
    // Validate that the decrypted value looks like a tax ID
    // US Tax IDs are typically 9 digits (SSN) or formatted with dashes
    // Trim the decrypted value to handle trailing spaces
    const trimmedDecrypted = decrypted.trim();
    const taxIdPattern = /^\d{9}$|^\d{3}-\d{2}-\d{4}$|^\d{2}-\d{7}$/;
    
    if (taxIdPattern.test(trimmedDecrypted)) {
      // Format as XXX-XX-XXXX for display
      const digits = trimmedDecrypted.replace(/\D/g, '');
      if (digits.length === 9) {
        const formatted = `${digits.slice(0, 3)}-${digits.slice(3, 5)}-${digits.slice(5)}`;
        logger.info('Successfully decrypted tax ID', { formatted });
        return formatted;
      }
      logger.info('Successfully decrypted tax ID', { decrypted: trimmedDecrypted });
      return trimmedDecrypted;
    }
    
    logger.warn('Decrypted value does not match tax ID pattern', { 
      decrypted,
      pattern: taxIdPattern.toString() 
    });
    
    // If decrypted value doesn't look like a tax ID, show indicator
    return `[Unable to decrypt: ${trimmedValue.substring(0, 10)}...]`;
  } catch (error) {
    logger.error('Failed to decrypt tax ID', { error: error.message, stack: error.stack });
    return `[Decryption error]`;
  }
}

/**
 * Test decryption function - can be called from console
 */
export function testTaxIdDecryption(): void {
  logger.info('=== STARTING TAX ID DECRYPTION TEST ===');
  
  // Test with the actual encrypted value from VENDOR.DBF
  // CTAXID contains: "µ|Çø %Z¥¹" (raw binary)
  const testValue = "µ|Çø %Z¥¹";
  
  logger.info('Test Input', {
    value: testValue,
    length: testValue.length,
    charCodes: testValue.split('').map(c => c.charCodeAt(0)),
    hex: testValue.split('').map(c => '0x' + c.charCodeAt(0).toString(16).padStart(2, '0')).join(' ')
  });
  
  // Try direct binary decryption
  const bytes = stringToBytes(testValue);
  logger.info('Bytes Array', {
    bytes: Array.from(bytes),
    hex: Array.from(bytes).map(b => '0x' + b.toString(16).padStart(2, '0')).join(' ')
  });
  
  try {
    const decrypted = decryptFromBinary(bytes);
    logger.info('Decryption Result', {
      success: true,
      decrypted: decrypted,
      cleaned: decrypted.replace(/[\x00-\x08\x0B-\x0C\x0E-\x1F\x7F-\xFF]/g, ''),
      charCodes: decrypted.split('').map(c => c.charCodeAt(0)),
      hex: decrypted.split('').map(c => '0x' + c.charCodeAt(0).toString(16).padStart(2, '0')).join(' ')
    });
  } catch (error: any) {
    logger.error('Decryption Failed', {
      error: error.message,
      stack: error.stack
    });
  }
  
  logger.info('=== TAX ID DECRYPTION TEST COMPLETE ===');
}

// Make test function available globally for console
if (typeof window !== 'undefined') {
  (window as any).testTaxIdDecryption = testTaxIdDecryption;
}