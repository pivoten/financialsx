import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'

// Import translations
import en from '../locales/en.json'
import es from '../locales/es.json'

// Configuration options for i18n
const i18nConfig = {
  resources: {
    en: { translation: en },
    es: { translation: es }
  },
  fallbackLng: 'en',
  debug: false,
  
  interpolation: {
    escapeValue: false // React already escapes values
  },
  
  detection: {
    order: ['localStorage', 'navigator', 'htmlTag'],
    caches: ['localStorage'],
    lookupLocalStorage: 'i18nextLng'
  }
}

// Initialize i18n
i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init(i18nConfig)
  .then(() => {
    console.log('i18n initialized successfully')
  })
  .catch((error) => {
    console.error('Failed to initialize i18n:', error)
  })

export default i18n

// Helper function to get app config from translations
export const getAppConfig = () => {
  const t = i18n.t
  return {
    appName: t('app.name'),
    tagline: t('app.tagline'),
    copyright: t('app.copyright')
  }
}

// Helper to change language
export const changeLanguage = (lng: string) => {
  i18n.changeLanguage(lng)
  localStorage.setItem('i18nextLng', lng)
}

// Get available languages
export const getAvailableLanguages = () => [
  { code: 'en', name: 'English' },
  { code: 'es', name: 'Español' }
]