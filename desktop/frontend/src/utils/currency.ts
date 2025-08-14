import Decimal from 'decimal.js';

// Configure Decimal for currency calculations (2 decimal places)
Decimal.set({ precision: 20, rounding: Decimal.ROUND_HALF_UP });

/**
 * Currency class for handling monetary values with proper decimal precision
 * Prevents floating-point arithmetic errors in financial calculations
 */
export class Currency {
  private value: Decimal;

  constructor(value: Decimal | number | string) {
    if (value instanceof Decimal) {
      this.value = value.toDecimalPlaces(2);
    } else {
      this.value = new Decimal(value).toDecimalPlaces(2);
    }
  }

  /**
   * Create a Currency from a number (use carefully with DBF data)
   */
  static fromNumber(num: number): Currency {
    return new Currency(num);
  }

  /**
   * Create a Currency from a string representation
   */
  static fromString(str: string): Currency {
    return new Currency(str);
  }

  /**
   * Create a Currency from cents
   */
  static fromCents(cents: number): Currency {
    return new Currency(new Decimal(cents).dividedBy(100));
  }

  /**
   * Create a zero Currency value
   */
  static zero(): Currency {
    return new Currency(0);
  }

  /**
   * Parse a value from DBF or API response
   */
  static parse(value: any): Currency {
    if (value === null || value === undefined || value === '') {
      return Currency.zero();
    }
    
    if (typeof value === 'number') {
      return Currency.fromNumber(value);
    }
    
    if (typeof value === 'string') {
      // Remove currency symbols and commas
      const cleaned = value.replace(/[$,]/g, '').trim();
      if (cleaned === '' || cleaned === '-') {
        return Currency.zero();
      }
      try {
        return Currency.fromString(cleaned);
      } catch {
        return Currency.zero();
      }
    }
    
    return Currency.zero();
  }

  /**
   * Add two Currency values
   */
  add(other: Currency): Currency {
    return new Currency(this.value.plus(other.value));
  }

  /**
   * Subtract a Currency value
   */
  subtract(other: Currency): Currency {
    return new Currency(this.value.minus(other.value));
  }

  /**
   * Multiply by a number
   */
  multiply(factor: number | Decimal): Currency {
    return new Currency(this.value.times(factor));
  }

  /**
   * Divide by a number
   */
  divide(divisor: number | Decimal): Currency {
    return new Currency(this.value.dividedBy(divisor));
  }

  /**
   * Get the negative of this Currency
   */
  negate(): Currency {
    return new Currency(this.value.negated());
  }

  /**
   * Get the absolute value
   */
  abs(): Currency {
    return new Currency(this.value.abs());
  }

  /**
   * Check if positive
   */
  isPositive(): boolean {
    return this.value.greaterThan(0);
  }

  /**
   * Check if negative
   */
  isNegative(): boolean {
    return this.value.lessThan(0);
  }

  /**
   * Check if zero
   */
  isZero(): boolean {
    return this.value.isZero();
  }

  /**
   * Compare with another Currency
   */
  greaterThan(other: Currency): boolean {
    return this.value.greaterThan(other.value);
  }

  /**
   * Compare with another Currency
   */
  lessThan(other: Currency): boolean {
    return this.value.lessThan(other.value);
  }

  /**
   * Check equality with another Currency
   */
  equals(other: Currency): boolean {
    return this.value.equals(other.value);
  }

  /**
   * Convert to cents (integer)
   */
  toCents(): number {
    return this.value.times(100).toNumber();
  }

  /**
   * Convert to number (use with caution)
   */
  toNumber(): number {
    return this.value.toNumber();
  }

  /**
   * Convert to string with 2 decimal places
   */
  toString(): string {
    return this.value.toFixed(2);
  }

  /**
   * Format as currency string with dollar sign
   */
  format(): string {
    const isNegative = this.isNegative();
    const absValue = this.abs().toString();
    
    // Add thousand separators
    const parts = absValue.split('.');
    parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, ',');
    const formatted = parts.join('.');
    
    return isNegative ? `-$${formatted}` : `$${formatted}`;
  }

  /**
   * Format for display with optional color coding
   */
  formatWithColor(): { value: string; className: string } {
    const formatted = this.format();
    const className = this.isNegative() ? 'text-red-600' : 'text-green-600';
    return { value: formatted, className };
  }

  /**
   * Create a Currency from a JSON value
   */
  static fromJSON(json: any): Currency {
    return Currency.parse(json);
  }

  /**
   * Convert to JSON-serializable value
   */
  toJSON(): string {
    return this.toString();
  }
}

/**
 * Sum an array of Currency values
 */
export function sumCurrencies(values: Currency[]): Currency {
  return values.reduce(
    (sum, value) => sum.add(value),
    Currency.zero()
  );
}

/**
 * Calculate the average of Currency values
 */
export function averageCurrencies(values: Currency[]): Currency {
  if (values.length === 0) {
    return Currency.zero();
  }
  const sum = sumCurrencies(values);
  return sum.divide(values.length);
}

/**
 * Find the maximum Currency value
 */
export function maxCurrency(values: Currency[]): Currency {
  if (values.length === 0) {
    return Currency.zero();
  }
  return values.reduce((max, value) => 
    value.greaterThan(max) ? value : max
  );
}

/**
 * Find the minimum Currency value
 */
export function minCurrency(values: Currency[]): Currency {
  if (values.length === 0) {
    return Currency.zero();
  }
  return values.reduce((min, value) => 
    value.lessThan(min) ? value : min
  );
}

/**
 * Format a number as currency (legacy support)
 * @deprecated Use Currency class instead
 */
export function formatCurrency(value: number | string | null | undefined): string {
  return Currency.parse(value).format();
}

/**
 * Parse and calculate a GL balance using proper decimal arithmetic
 */
export function calculateGLBalance(
  debits: number | string,
  credits: number | string,
  accountType: number
): Currency {
  const totalDebits = Currency.parse(debits);
  const totalCredits = Currency.parse(credits);
  
  // Account Types:
  // 1 = Assets (Debit normal balance)
  // 2 = Liabilities (Credit normal balance)
  // 3 = Equity (Credit normal balance)
  // 4 = Revenue/Income (Credit normal balance)
  // 5 = Expenses (Debit normal balance)
  
  switch (accountType) {
    case 1: // Assets
    case 5: // Expenses
      return totalDebits.subtract(totalCredits);
    case 2: // Liabilities
    case 3: // Equity
    case 4: // Revenue
      return totalCredits.subtract(totalDebits);
    default:
      // Default to asset behavior
      return totalDebits.subtract(totalCredits);
  }
}