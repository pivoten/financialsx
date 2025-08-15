import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

// Import translation files
import enTranslations from '../locales/en.json';
import esTranslations from '../locales/es.json';

// Define available languages
export const languages = {
  en: { name: 'English', flag: '🇺🇸' },
  es: { name: 'Español', flag: '🇪🇸' }
};

// Initialize i18n
i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      en: { translation: enTranslations },
      es: { translation: esTranslations }
    },
    fallbackLng: 'en',
    debug: false,
    
    interpolation: {
      escapeValue: false // React already escapes values
    },
    
    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      caches: ['localStorage']
    }
  });

// Helper functions
export const changeLanguage = (lng: string) => {
  i18n.changeLanguage(lng);
  localStorage.setItem('i18nextLng', lng);
};

export const getCurrentLanguage = () => {
  return i18n.language || 'en';
};

export const getAvailableLanguages = () => {
  return Object.keys(languages);
};

// Format currency based on locale
export const formatCurrency = (amount: number): string => {
  const locale = i18n.language === 'es' ? 'es-ES' : 'en-US';
  const currency = i18n.language === 'es' ? 'EUR' : 'USD';
  
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency: currency
  }).format(amount);
};

// Format date based on locale  
export const formatDate = (date: Date | string): string => {
  const locale = i18n.language === 'es' ? 'es-ES' : 'en-US';
  const dateObj = typeof date === 'string' ? new Date(date) : date;
  
  return new Intl.DateTimeFormat(locale, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  }).format(dateObj);
};

// Format number based on locale
export const formatNumber = (num: number): string => {
  const locale = i18n.language === 'es' ? 'es-ES' : 'en-US';
  return new Intl.NumberFormat(locale).format(num);
};

export default i18n;