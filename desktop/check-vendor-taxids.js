const dbase = require('@valentin-kaiser/go-dbase/dbase');
const fs = require('fs');
const path = require('path');

async function checkVendorTaxIds() {
  const dbfPath = path.join(__dirname, 'datafiles', 'limecreekenergyllcdata', 'VENDOR.DBF');
  
  try {
    const db = await dbase.Open(dbfPath);
    const records = await db.Records();
    
    let encryptedCount = 0;
    let samples = [];
    
    for (let i = 0; i < Math.min(10, records.length); i++) {
      const record = records[i];
      const taxId = record.CTAXID;
      
      if (taxId && taxId.trim()) {
        const isBase64 = /^[A-Za-z0-9+/]+=*$/.test(taxId);
        console.log(`Record ${i}: CTAXID = "${taxId}" (Base64: ${isBase64})`);
        
        if (isBase64 && taxId.length >= 12 && taxId.length <= 30) {
          encryptedCount++;
          samples.push(taxId);
        }
      }
    }
    
    console.log(`\nFound ${encryptedCount} potentially encrypted tax IDs`);
    console.log('Samples:', samples);
    
    await db.Close();
  } catch (error) {
    console.error('Error:', error);
  }
}

checkVendorTaxIds();
