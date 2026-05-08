import js from '@eslint/js'
import tseslint from 'typescript-eslint'
import pluginVue from 'eslint-plugin-vue'
import prettier from 'eslint-config-prettier'

export default tseslint.config(
  // Base JS rules
  js.configs.recommended,
  // TypeScript rules
  ...tseslint.configs.recommended,
  // Vue rules
  ...pluginVue.configs['flat/recommended'],
  // Prettier must be last — disables all formatting rules
  prettier,
  {
    files: ['src/**/*.{ts,vue}'],
    languageOptions: {
      parserOptions: {
        parser: tseslint.parser,
        extraFileExtensions: ['.vue'],
      },
    },
    rules: {
      // Allow @ts-ignore and explicit any in tests
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_', varsIgnorePattern: '^_' }],
      // Vue
      'vue/multi-word-component-names': 'off',
      'vue/require-default-prop': 'off',
      'vue/html-self-closing': 'off',
      // Allow console in dev files
      'no-console': ['warn', { allow: ['warn', 'error'] }],
    },
  },
  {
    files: ['src/__tests__/**'],
    rules: {
      '@typescript-eslint/no-explicit-any': 'off',
      'no-console': 'off',
    },
  },
  {
    ignores: ['dist/**', 'node_modules/**', '*.config.{js,ts}'],
  }
)
