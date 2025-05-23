import { extendTheme, type ThemeConfig } from '@chakra-ui/react'

const config: ThemeConfig = {
  initialColorMode: 'light',
  useSystemColorMode: false,
}

const theme = extendTheme({ 
  config,
  colors: {
    brand: {
      50: '#f5f7fa',
      100: '#e4ebf5',
      200: '#cbd7e7',
      300: '#a6bcd8',
      400: '#7a9cc5',
      500: '#5981b0',
      600: '#43669a',
      700: '#385283',
      800: '#30456c',
      900: '#2c3c5b',
    },
  },
  fonts: {
    heading: `'Inter', sans-serif`,
    body: `'Inter', sans-serif`,
  },
})

export default theme
