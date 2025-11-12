import { useTranslation } from 'react-i18next'

import { CryptoInfo } from '@entities/crypto'
import { NetworkStatus } from '@entities/network-status'

import { Button } from '@shared/uikit/button'

import styles from './header.module.sass'

type Props = {
	onLogout: () => void
	onLogoutAll: () => void
}

export const HeaderView = ({ onLogout, onLogoutAll }: Props) => {
	const { t } = useTranslation()

	return (
		<div className={styles.root}>
			<div className={styles.information}>
				<NetworkStatus />
				<CryptoInfo />
			</div>
			<div className={styles.actions}>
				<Button label={t('logout')} text onClick={onLogout} />
				<Button label={t('logout_all')} text severity="secondary" onClick={onLogoutAll} />
			</div>
		</div>
	)
}
