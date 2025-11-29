import { useTranslation } from 'react-i18next'

import { CryptoInfo } from '@entities/crypto'
import { NetworkStatus } from '@entities/network-status'
import { ServerStatus } from '@entities/server-status'

import { Button } from '@shared/uikit/button'

import styles from './header.module.sass'

type Props = {
	isOnline: boolean
	onLogout: () => void
	onLogoutAll: () => void
}

export const HeaderView = ({ isOnline, onLogout, onLogoutAll }: Props) => {
	const { t } = useTranslation()

	return (
		<div className={styles.root}>
			<div className={styles.information}>
				<ServerStatus />
				<NetworkStatus />
				<CryptoInfo />
			</div>
			<div className={styles.actions}>
				<Button label={t('logout')} text onClick={onLogout} disabled={!isOnline} />
				<Button
					label={t('logout_all')}
					text
					severity={'secondary'}
					onClick={onLogoutAll}
					disabled={!isOnline}
				/>
			</div>
		</div>
	)
}
