import type { ReactNode } from 'react'
import { I18nextProvider } from 'react-i18next'

import { initI18next } from '@shared/localization'

type Props = {
	children?: ReactNode
}

export const LocalizationProvider = ({ children }: Props) => {
	const i18n = initI18next()

	return <I18nextProvider i18n={i18n}>{children}</I18nextProvider>
}
