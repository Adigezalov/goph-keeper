import { useTranslation } from 'react-i18next'

import { ProgressSpinner } from '@shared/uikit/progress-spinner'

import styles from './page-loader.module.sass'

export const PageLoader = () => {
	const { t } = useTranslation()

	return (
		<div className={styles.container}>
			<div className={styles.loader}>
				<ProgressSpinner />
				<div className={styles.text}>{t('loading_page')}</div>
			</div>
		</div>
	)
}
