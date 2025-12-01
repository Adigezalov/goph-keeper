import js from '@eslint/js'
import tseslint from '@typescript-eslint/eslint-plugin'
import tsParser from '@typescript-eslint/parser'
import prettier from 'eslint-config-prettier'
import react from 'eslint-plugin-react'
import reactHooks from 'eslint-plugin-react-hooks'

export default [
	js.configs.recommended,
	{
		files: ['**/*.ts', '**/*.tsx'],
		languageOptions: {
			parser: tsParser,
			parserOptions: {
				ecmaVersion: 'latest',
				sourceType: 'module',
				project: ['./tsconfig.eslint.json'],
			},
			globals: {
				window: 'readonly',
				document: 'readonly',
				localStorage: 'readonly',
				navigator: 'readonly',
				setTimeout: 'readonly',
				clearTimeout: 'readonly',
				console: 'readonly',
				process: 'readonly',
				describe: 'readonly',
				test: 'readonly',
				expect: 'readonly',
				it: 'readonly',
				alert: 'readonly',
				btoa: 'readonly',
				Image: 'readonly',
				NodeJS: 'readonly',
			},
		},
		plugins: {
			'@typescript-eslint': tseslint,
			react,
			'react-hooks': reactHooks,
		},
		rules: {
			'react/react-in-jsx-scope': 'off',
			'no-trailing-spaces': [1, { skipBlankLines: true }],
			'eol-last': [0],
			'react/display-name': 'off',
			'react-hooks/rules-of-hooks': 'error',
			'react-hooks/exhaustive-deps': 'error',
			'no-mixed-spaces-and-tabs': ['error', 'smart-tabs'],
			'no-unused-vars': 'off',
			'@typescript-eslint/no-unused-vars': [
				'warn',
				{
					args: 'after-used',
					argsIgnorePattern: '^_',
					varsIgnorePattern: '^_',
					caughtErrors: 'all',
					caughtErrorsIgnorePattern: '^_',
				},
			],
			'@typescript-eslint/no-unsafe-argument': [0],
			'@typescript-eslint/consistent-type-definitions': 'off',
			'@typescript-eslint/explicit-function-return-type': 'off',
			'@typescript-eslint/strict-boolean-expressions': 'off',
			'@typescript-eslint/no-misused-promises': 'off',
			'@typescript-eslint/promise-function-async': 'off',
			'@typescript-eslint/no-dynamic-delete': 'off',
			'@typescript-eslint/no-invalid-void-type': 'off',
			'@typescript-eslint/prefer-nullish-coalescing': 'off',
			complexity: ['error', 30],
			'no-undef': 'off',
			'no-redeclare': 'off',
		},
	},
	{
		files: ['.eslintrc.{js,cjs}'],
		languageOptions: {
			sourceType: 'script',
		},
	},
	{
		ignores: [
			'./src/**/**/**/**/*.stories.tsx',
			'vite.config.ts',
			'vite-env.d.ts',
			'dist',
			'.idea',
			'./libs',
			'public/sw.js',
		],
	},
	prettier,
	{
		files: ['**/*.js'],
		languageOptions: {
			globals: {
				console: 'readonly',
			},
		},
	},
]
