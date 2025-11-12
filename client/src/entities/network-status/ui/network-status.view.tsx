import { classNames } from 'primereact/utils'

import { Icon } from '@shared/uikit/icon'

import styles from './network-status.module.sass'

type TProps = {
	isOnline: boolean
}

export const NetworkStatusView = ({ isOnline }: TProps) => {
	return (
		<div className={styles.root}>
			<span
				className={classNames(Icon.WIFI, styles.icon, { [styles.error]: !isOnline })}
			/>
			<div>{isOnline ? 'online' : 'offline'}</div>
		</div>
	)
}
