# SherWare-Compatible Encryption/Decryption

This document describes the SherWare-compatible encryption/decryption algorithm implemented in Go and TypeScript for legacy compatibility with a FoxPro FLL.

It is a symmetric XOR stream cipher using MSVCRT's `rand()` PRNG seeded from the key string.

> ⚠️ **Security Notice**  
> This is not cryptographically secure and should not be used for new security-critical features.

---

## Algorithm Summary

1. **Seed calculation**  
   `seed = sum(eachKeyByte + 1)` (uint32 wraparound)
2. **PRNG**  
   ```c
   seed = seed * 214013 + 2531011;
   randVal = (seed >> 16) & 0x7FFF;
   ```
3. **Per-byte transform**  
   ```c
   data[i] ^= ((randVal & 0xFF) ^ key[i % keyLen]);
   ```
4. Same function is used for encryption and decryption.

---

## Key

Current production key:

```
SherWareKey_@8@2899909
```

---

## Go Implementation

```go
package sherware

import "encoding/base64"

func seedFromKey(key []byte) uint32 {
    var s uint32
    for _, b := range key {
        s += uint32(b) + 1
    }
    return s
}

func msvcrtRand(seed uint32) (uint32, uint32) {
    seed = seed*214013 + 2531011
    return seed, (seed >> 16) & 0x7FFF
}

func Transform(buf []byte, key []byte) {
    if len(key) == 0 || len(buf) == 0 {
        return
    }
    klen := len(key)
    seed := seedFromKey(key)
    for i := 0; i < len(buf); i++ {
        var r uint32
        seed, r = msvcrtRand(seed)
        buf[i] ^= byte(r&0xFF) ^ key[i%klen]
    }
}

func EncryptToBase64(plain []byte, key []byte) string {
    out := make([]byte, len(plain))
    copy(out, plain)
    Transform(out, key)
    return base64.StdEncoding.EncodeToString(out)
}

func DecryptFromBase64(b64 string, key []byte) ([]byte, error) {
    cipher, err := base64.StdEncoding.DecodeString(b64)
    if err != nil {
        return nil, err
    }
    Transform(cipher, key)
    return cipher, nil
}
```

---

## TypeScript Implementation

```ts
export function transform(data: Uint8Array, key: Uint8Array): void {
  if (!key.length || !data.length) return;

  let seed = 0 >>> 0;
  for (let i = 0; i < key.length; i++) seed = (seed + (key[i] + 1)) >>> 0;

  const rnd = () => {
    seed = (Math.imul(seed, 214013) + 2531011) >>> 0;
    return (seed >>> 16) & 0x7fff;
  };

  for (let i = 0; i < data.length; i++) {
    const r = rnd() & 0xff;
    data[i] ^= (r ^ key[i % key.length]) & 0xff;
  }
}

export const utf8 = {
  enc: (s: string) => new TextEncoder().encode(s),
  dec: (b: Uint8Array) => new TextDecoder().decode(b),
};

export function toBase64(b: Uint8Array): string {
  if (typeof window === 'undefined') return Buffer.from(b).toString('base64');
  let s = ''; for (let i = 0; i < b.length; i++) s += String.fromCharCode(b[i]);
  return btoa(s);
}

export function fromBase64(s: string): Uint8Array {
  if (typeof window === 'undefined') return new Uint8Array(Buffer.from(s, 'base64'));
  const bin = atob(s);
  const out = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i);
  return out;
}

export function encryptToBase64(plain: string, keyStr: string): string {
  const buf = utf8.enc(plain);
  const out = new Uint8Array(buf);
  transform(out, utf8.enc(keyStr));
  return toBase64(out);
}

export function decryptFromBase64(b64: string, keyStr: string): string {
  const data = fromBase64(b64);
  transform(data, utf8.enc(keyStr));
  return utf8.dec(data);
}
```

---

## Usage Examples

**Go**:
```go
key := []byte("SherWareKey_@8@2899909")
cipher := sherware.EncryptToBase64([]byte("Hello"), key)
plain, _ := sherware.DecryptFromBase64(cipher, key)
```

**TypeScript**:
```ts
const key = "SherWareKey_@8@2899909";
const enc = encryptToBase64("Hello", key);
const dec = decryptFromBase64(enc, key);
```

---

## DBF Integration

- For **Character/Varchar** fields: store **Base64** of ciphertext.
- For **Binary/Memo** fields: store raw encrypted bytes directly.
- Encrypt on write, decrypt on read.
