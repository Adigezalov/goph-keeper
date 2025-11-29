import { ru } from '../lang'

const resources = {
	translation: ru,
} as const

declare module 'i18next' {
	interface CustomTypeOptions {
		resources: typeof resources
	}
	interface InitOptions {
		plural?: (value: number, options: any) => string
	}
}
