import i18next, { type InitOptions } from 'i18next'
import { initReactI18next } from 'react-i18next'

import { ru } from '../lang'

const customTypeOptions: InitOptions = {
	resources: {
		ru: {
			translation: ru,
		},
	},
	lng: 'ru',
	fallbackLng: 'ru',
	plural: (value: any) => {
		if (typeof value !== 'number') {
			return ''
		}

		if (value % 10 === 1 && value % 100 !== 11) {
			return 'one'
		}

		if ([2, 3, 4].includes(value % 10) && ![12, 13, 14].includes(value % 100)) {
			return 'few'
		}

		if (
			[0, 5, 6, 7, 8, 9].includes(value % 10) ||
			[11, 12, 13, 14].includes(value % 100)
		) {
			return 'many'
		}

		return ''
	},
}

export const initI18next = () => {
	void i18next.use(initReactI18next).init(customTypeOptions)

	return i18next
}
